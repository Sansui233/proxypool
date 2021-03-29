package healthcheck

import "github.com/Sansui233/proxypool/pkg/proxy"

// Statistic for a proxy
type Stat struct {
	Speed    float64
	Delay    uint16
	ReqCount uint16
	Relay    bool
	Pool     bool
	OutIp    string
	Id       string
}

// Statistic array for proxies
type StatList []Stat

// ProxyStats stores proxies' statistics
var ProxyStats StatList

func init() {
	ProxyStats = make(StatList, 0)
}

// Update speed for a Stat
func (ps *Stat) UpdatePSSpeed(speed float64) {
	if ps.Speed < 60 && ps.Speed != 0 {
		ps.Speed = 0.3*ps.Speed + 0.7*speed
	} else {
		ps.Speed = speed
	}
}

// Update delay for a Stat
func (ps *Stat) UpdatePSDelay(delay uint16) {
	ps.Delay = delay
}

// Update out ip for a Stat
func (ps *Stat) UpdatePSOutIp(outIp string) {
	ps.OutIp = outIp
}

// Count + 1 for a Stat
func (ps *Stat) UpdatePSCount() {
	ps.ReqCount++
}

// Find a proxy's Stat in StatList
func (psList StatList) Find(p proxy.Proxy) (*Stat, bool) {
	s := p.Identifier()
	for i, _ := range psList {
		if psList[i].Id == s {
			return &psList[i], true
		}
	}
	return nil, false
}

// Return proxies that request count more than a given nubmer
func (psList StatList) ReqCountThan(n uint16, pl []proxy.Proxy, reset bool) []proxy.Proxy {
	proxies := make([]proxy.Proxy, 0)
	for _, p := range pl {
		for j, _ := range psList {
			if psList[j].ReqCount > n && p.Identifier() == psList[j].Id {
				proxies = append(proxies, p)
			}
		}
	}
	// reset request count
	if reset {
		for i, _ := range psList {
			psList[i].ReqCount = 0
		}
	}
	return proxies
}

// Sort proxies by speed. Notice that this returns the same pointer.
func (psList StatList) SortProxiesBySpeed(proxies []proxy.Proxy) []proxy.Proxy {
	if ok := checkErrorProxies(proxies); !ok {
		return proxies
	}
	l := len(proxies)
	if l == 1 {
		return proxies
	}
	// Classic bubble Sort. Biggest the first
	for i := 0; i < l-1; i++ { // i defines unsorted list bound
		flag := false
		for j := 0; j < l-1-i; j++ {
			ps1, ok1 := psList.Find(proxies[j])
			ps2, ok2 := psList.Find(proxies[j+1])
			// validate records, put no record proxy behind
			if !ok2 {
				continue
			} else if !ok1 && ok2 {
				t := proxies[j]
				proxies[j] = proxies[j+1]
				proxies[j+1] = t
				flag = true
				continue
			}
			// else: validate speed value, put zero speed proxy behind
			if ps2.Speed == 0 {
				continue
			} else if ps1.Speed == 0 { // when ps2.speed != 0, validate ps1
				t := proxies[j]
				proxies[j] = proxies[j+1]
				proxies[j+1] = t
				flag = true
				continue
			} else {
				// Reach the real speed sort. Too much code on validation. I'm so tired
				if ps1.Speed < ps2.Speed {
					t := proxies[j]
					proxies[j] = proxies[j+1]
					proxies[j+1] = t
					flag = true
				}
			}
		}
		if flag == false {
			break
		}
	}
	return proxies
}
