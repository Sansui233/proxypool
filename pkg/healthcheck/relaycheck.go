package healthcheck

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Dreamacro/clash/adapters/outbound"
	"github.com/Sansui233/proxypool/log"
	"github.com/Sansui233/proxypool/pkg/proxy"
	"github.com/ivpusic/grpool"
	"net"
	"sync"
	"time"
)

func RelayCheck(proxies proxy.ProxyList) {
	pool := grpool.NewPool(500, 200)
	pool.WaitCount(len(proxies))
	m := sync.Mutex{}

	log.Infoln("Relay Test ON")
	doneCount := 0
	go func() {
		for _, p := range proxies {
			pp := p
			pool.JobQueue <- func() {
				defer pool.JobDone()
				out, err := testRelay(pp)
				if err == nil && out != "" {
					m.Lock()
					// Relay or pool
					if isRelay(pp.BaseInfo().Server, out) {
						if ps, ok := ProxyStats.Find(pp); ok {
							ps.UpdatePSOutIp(out)
							ps.Relay = true
						} else {
							ps = &Stat{
								Id:    pp.Identifier(),
								Relay: true,
								OutIp: out,
							}
							ProxyStats = append(ProxyStats, *ps)
						}
					} else { // is pool ip
						if ps, ok := ProxyStats.Find(pp); ok {
							ps.UpdatePSOutIp(out)
							ps.Pool = true
						} else {
							ps = &Stat{
								Id:    pp.Identifier(),
								Pool:  true,
								OutIp: out,
							}
							ProxyStats = append(ProxyStats, *ps)
						}
					}
					m.Unlock()
				}
				doneCount++
				progress := float64(doneCount) * 100 / float64(len(proxies))
				fmt.Printf("\r\t[%5.1f%% DONE]", progress)
			}
		}
	}()
	pool.WaitAll()
	pool.Release()
	fmt.Println()
}

// Get outbound relay ip
func testRelay(p proxy.Proxy) (outip string, err error) {
	pmap := make(map[string]interface{})
	err = json.Unmarshal([]byte(p.String()), &pmap)
	if err != nil {
		return "", err
	}

	pmap["port"] = int(pmap["port"].(float64))
	if p.TypeName() == "vmess" {
		pmap["alterId"] = int(pmap["alterId"].(float64))
		if network, ok := pmap["network"]; ok && network.(string) == "h2" {
			return "", nil // todo 暂无方法测试h2的延迟，clash对于h2的connection会阻塞
		}
	}

	clashProxy, err := outbound.ParseProxy(pmap)
	if err != nil {
		return "", err
	}

	b, err := HTTPGetBodyViaProxyWithTime(clashProxy, "http://ipinfo.io/ip", time.Second*10)
	if err != nil {
		return "", err
	}

	if string(b) == p.BaseInfo().Server {
		return "", nil // not relay
	}

	address := net.ParseIP(string(b))
	if address == nil {
		return "", errors.New("error outbound ip format")
	}

	return string(b), nil
}

// Distinguish pool ip from relay. false for pool, true for relay
func isRelay(src string, out string) bool {
	ipv4Mask := net.CIDRMask(24, 32)
	ip1 := net.ParseIP(src)
	ip2 := net.ParseIP(out)
	if fmt.Sprint(ip1.Mask(ipv4Mask)) == fmt.Sprint(ip2.Mask(ipv4Mask)) { // Pool
		return false
	}
	return true
}
