package provider

import (
	"fmt"
	"github.com/Sansui233/proxypool/log"
	"github.com/Sansui233/proxypool/pkg/healthcheck"
	"math"
	"strconv"
	"strings"

	"github.com/Sansui233/proxypool/pkg/proxy"
)

type Provider interface {
	Provide() string
}

type Base struct {
	Proxies    *proxy.ProxyList `yaml:"proxies"`
	Types      string           `yaml:"type"`
	Country    string           `yaml:"country"`
	NotCountry string           `yaml:"not_country"`
	Speed      string           `yaml:"speed"`
}

// 根据子类的的Provide()传入的信息筛选节点，结果会改变传入的proxylist。
func (b *Base) preFilter() {
	proxies := make(proxy.ProxyList, 0)

	if ok := checkErrorProxies(*b.Proxies); !ok {
		log.Warnln("provider: nothing to provide")
		b.Proxies = &proxies
		return
	}

	needFilterType := true
	needFilterCountry := true
	needFilterNotCountry := true
	needFilterSpeed := true
	if b.Types == "" || b.Types == "all" {
		needFilterType = false
	}
	if b.Country == "" || b.Country == "all" {
		needFilterCountry = false
	}
	if b.NotCountry == "" {
		needFilterNotCountry = false
	}
	if b.Speed == "" {
		needFilterSpeed = true
	}
	types := strings.Split(b.Types, ",")
	countries := strings.Split(b.Country, ",")
	notCountries := strings.Split(b.NotCountry, ",")
	speedMin, speedMax := checkSpeed(strings.Split(b.Speed, ","))

	if speedMin == -1 {
		needFilterSpeed = false
	}

	bProxies := *b.Proxies
	for _, p := range bProxies {
		if needFilterType {
			typeOk := false
			for _, t := range types {
				if p.TypeName() == t {
					typeOk = true
					break
				}
			}
			if !typeOk {
				goto exclude
			}
		}

		if needFilterNotCountry {
			for _, c := range notCountries {
				if strings.Contains(p.BaseInfo().Name, c) {
					goto exclude
				}
			}
		}

		if needFilterCountry {
			countryOk := false
			for _, c := range countries {
				if strings.Contains(p.BaseInfo().Name, c) {
					countryOk = true
					break
				}
			}
			if !countryOk {
				goto exclude
			}
		}

		if needFilterSpeed && len(healthcheck.ProxyStats) != 0 {
			if ps, ok := healthcheck.ProxyStats.Find(p); ok {
				if ps.Speed != 0 {
					// clear history speed tag
					names := strings.Split(p.BaseInfo().Name, " |")
					if len(names) > 1 {
						p.BaseInfo().Name = names[0]
					}
					// check speed
					if ps.Speed > speedMin && ps.Speed < speedMax {
						p.AddToName(fmt.Sprintf(" |%5.2fMb", ps.Speed))
					} else {
						goto exclude
					}
				} else {
					if speedMin != 0 { // still show 0 speed proxy when speed Min is 0
						goto exclude
					}
				}
			} else {
				if speedMin != 0 { // still show no speed result proxy when speed Min is 0
					goto exclude
				}
			}
		} else { // When no filter needed: clear speed tag. But I don't know why speed is stored in name while provider get proxies from cache everytime. It's name should be refreshed without speed tag. Because of gin-cache?
			names := strings.Split(p.BaseInfo().Name, " |")
			if len(names) > 1 {
				p.BaseInfo().Name = names[0]
			}
		}

		proxies = append(proxies, p)
		// update statistic
		if ps, ok := healthcheck.ProxyStats.Find(p); ok {
			ps.UpdatePSCount()
		} else {
			healthcheck.ProxyStats = append(healthcheck.ProxyStats, healthcheck.Stat{
				Id:       p.Identifier(),
				ReqCount: 1,
			})
		}
	exclude:
	}

	b.Proxies = &proxies
}

func checkErrorProxies(proxies []proxy.Proxy) bool {
	if proxies == nil {
		return false
	}
	if len(proxies) == 0 {
		return false
	}
	if proxies[0] == nil {
		return false
	}
	return true
}

func checkSpeed(speed []string) (speedMin float64, speedMax float64) {
	speedMin, speedMax = 0, 1000
	var err1, err2 error
	switch len(speed) {
	case 1:
		if speed[0] != "" {
			speedMin, err1 = strconv.ParseFloat(speed[0], 64)
		}
	case 2:
		speedMin, err1 = strconv.ParseFloat(speed[0], 64)
		speedMax, err2 = strconv.ParseFloat(speed[1], 64)
	}
	if math.IsNaN(speedMin) || err1 != nil {
		speedMin = 0.00
	}
	if math.IsNaN(speedMax) || err2 != nil {
		speedMax = 1000.00
	}
	return speedMin, speedMax
}
