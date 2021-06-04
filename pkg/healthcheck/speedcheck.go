package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Dreamacro/clash/adapters/outbound"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Sansui233/proxypool/log"
	"github.com/Sansui233/proxypool/pkg/proxy"
	"github.com/ivpusic/grpool"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var SpeedTimeout = time.Second * 10

// SpeedTestAll tests speed of a group of proxies. Results are stored in ProxyStats
func SpeedTestAll(proxies []proxy.Proxy, conns int) {
	if ok := checkErrorProxies(proxies); !ok {
		return
	}
	numWorker := conns
	if numWorker <= 0 {
		numWorker = 5
	}
	numJob := 1
	if numWorker > 4 {
		numJob = (numWorker + 2) / 4
	}
	resultCount := 0
	m := sync.Mutex{}

	log.Infoln("Speed Test ON")
	doneCount := 0
	// use grpool
	pool := grpool.NewPool(numWorker, numJob)
	pool.WaitCount(len(proxies))
	for _, p := range proxies {
		pp := p
		pool.JobQueue <- func() {
			defer pool.JobDone()
			speed, err := ProxySpeedTest(pp)
			if err == nil || speed > 0 {
				m.Lock()
				if proxyStat, ok := ProxyStats.Find(pp); ok {
					proxyStat.UpdatePSSpeed(speed)
				} else {
					ProxyStats = append(ProxyStats, Stat{
						Id:    pp.Identifier(),
						Speed: speed,
					})
				}
				resultCount++
				m.Unlock()
			}
			doneCount++
			progress := float64(doneCount) * 100 / float64(len(proxies))
			fmt.Printf("\r\t[%5.1f%% DONE]", progress)
		}
	}
	pool.WaitAll()
	pool.Release()
	fmt.Println()
	log.Infoln("Speed Test Done. Count all speed results: %d", resultCount)
}

// SpeedTestNew tests speed of new proxies which is not in ProxyStats. Then appended to ProxyStats
func SpeedTestNew(proxies []proxy.Proxy, conns int) {
	if ok := checkErrorProxies(proxies); !ok {
		return
	}
	numWorker := conns
	if numWorker <= 0 {
		numWorker = 5
	}
	numJob := 1
	if numWorker > 4 {
		numJob = (numWorker + 2) / 4
	}
	resultCount := 0
	m := sync.Mutex{}

	log.Infoln("Speed Test ON")
	doneCount := 0
	// use grpool
	pool := grpool.NewPool(numWorker, numJob)
	pool.WaitCount(len(proxies))
	for _, p := range proxies {
		pp := p
		pool.JobQueue <- func() {
			defer pool.JobDone()
			m.Lock()
			if proxyStat, ok := ProxyStats.Find(pp); !ok {
				// when proxy's Stat not exits
				speed, err := ProxySpeedTest(pp)
				if err == nil || speed > 0 {
					ProxyStats = append(ProxyStats, Stat{
						Id:    pp.Identifier(),
						Speed: speed,
					})
					resultCount++
				}
			} else if proxyStat.Speed == 0 {
				speed, err := ProxySpeedTest(pp)
				if err == nil || speed > 0 {
					proxyStat.UpdatePSSpeed(speed)
					resultCount++
				}
			}
			m.Unlock()
			doneCount++
			progress := float64(doneCount) * 100 / float64(len(proxies))
			fmt.Printf("\r\t[%5.1f%% DONE]", progress)
		}
	}
	pool.WaitAll()
	pool.Release()
	fmt.Println()
	log.Infoln("Speed Test Done. New speed results count: %d", resultCount)
}

// ProxySpeedTest returns a speed result of a proxy. The speed result is like 20Mbit/s. -1 for error.
func ProxySpeedTest(p proxy.Proxy) (speedResult float64, err error) {
	// convert to clash proxy struct
	pmap := make(map[string]interface{})
	err = json.Unmarshal([]byte(p.String()), &pmap)
	if err != nil {
		return -1, err
	}
	pmap["port"] = int(pmap["port"].(float64))
	if p.TypeName() == "vmess" {
		pmap["alterId"] = int(pmap["alterId"].(float64))
		if network, ok := pmap["network"]; ok && network.(string) == "h2" {
			return 0, nil // todo 暂无方法测试h2的速度，clash对于h2的connection会阻塞
		}
	}

	clashProxy, err := outbound.ParseProxy(pmap)
	if err != nil {
		return -1, err
	}

	// start speedtest using speedtest.net
	var user *User
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		user, _ = fetchUserInfo(clashProxy)
	}()
	serverList, err := fetchServerList(clashProxy)
	if err != nil {
		return -1, err
	}

	// deal fetchUserInfo routine
	wg.Wait()

	// some logically unexpected error handling
	if user == nil {
		return -1, errors.New("fetch User Infoln failed in go routine") // 我真的不会用channel抛出err，go routine的不明原因阻塞我服了。下面的两个BUG现在都不知道原因，逻辑上不该出现的
	}
	if &serverList == nil {
		return -1, errors.New("unexpected error when fetching serverlist: addr of var serverlist nil")
	}
	if len(serverList.Servers) == 0 {
		return -1, errors.New("unexpected error when fetching serverlist: unexpected 0 server")
	}

	// Calculate distance
	for i := range serverList.Servers {
		server := serverList.Servers[i]
		sLat, _ := strconv.ParseFloat(server.Lat, 64)
		sLon, _ := strconv.ParseFloat(server.Lon, 64)
		uLat, _ := strconv.ParseFloat(user.Lat, 64)
		uLon, _ := strconv.ParseFloat(user.Lon, 64)
		server.Distance = distance(sLat, sLon, uLat, uLon)
	}
	// Sort by distance
	sort.Sort(ByDistance{serverList.Servers})

	var targets Servers
	targets = append(serverList.Servers[:3])

	// Test
	targets.StartTest(clashProxy)
	speedResult = targets.GetResult()

	return speedResult, nil

}

/* Test with SpeedTest.net */
// Download Size(MB)  0.245 0.5 1.125  2   5     8     12.5  18    24.5  32
var dlSizes = [...]int{350, 500, 750, 1000, 1500, 2000, 2500, 3000, 3500, 4000}

//var ulSizes = [...]int{100, 300, 500, 800, 1000, 1500, 2500, 3000, 3500, 4000} //kB

func pingTest(clashProxy C.Proxy, sURL string) time.Duration {
	pingURL := strings.Split(sURL, "/upload")[0] + "/latency.txt"

	l := time.Second * 10
	for i := 0; i < 2; i++ {
		sTime := time.Now()
		err := HTTPGetViaProxy(clashProxy, pingURL)
		fTime := time.Now()
		if err != nil {
			continue
		}
		if fTime.Sub(sTime) < l {
			l = fTime.Sub(sTime)
		}
	}
	return l / 2.0
}

// return a speed(Mbps)
func downloadTest(clashProxy C.Proxy, sURL string, latency time.Duration) float64 {
	dlURL := strings.Split(sURL, "/upload")[0]

	// Warming up
	sTime := time.Now()
	err := dlWarmUp(clashProxy, dlURL)
	fTime := time.Now()
	if err != nil {
		return 0
	}
	// 1.125MB for each request (750 * 750 * 2)
	wuSpeed := 1.125 * 8 * 2 / fTime.Sub(sTime.Add(latency)).Seconds()

	// Decide workload by warm up speed. Weight is the level of size.
	weight := 0
	if 10.0 < wuSpeed {
		weight = 5
	} else if 5 < wuSpeed {
		weight = 4
	} else if 2.5 < wuSpeed {
		weight = 3
	} else { // if too slow, skip main test to save time
		return wuSpeed
	}

	// Main speedtest
	dlSpeed := wuSpeed
	sTime = time.Now()
	err = downloadRequest(clashProxy, dlURL, weight)
	fTime = time.Now()
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		return wuSpeed // todo Incorrect Result
	}
	reqMB := dlSizes[weight] * dlSizes[weight] * 2 / 1000 / 1000
	dlSpeed = float64(reqMB) * 8 / fTime.Sub(sTime).Seconds()
	return dlSpeed
}

func dlWarmUp(clashProxy C.Proxy, dlURL string) error {
	size := dlSizes[2]
	url := dlURL + "/random" + strconv.Itoa(size) + "x" + strconv.Itoa(size) + ".jpg"
	err := HTTPGetBodyForSpeedTest(clashProxy, url, SpeedTimeout)
	if err != nil {
		return err
	}
	return nil
}

func downloadRequest(clashProxy C.Proxy, dlURL string, w int) error {
	size := dlSizes[w]
	url := dlURL + "/random" + strconv.Itoa(size) + "x" + strconv.Itoa(size) + ".jpg"
	err := HTTPGetBodyForSpeedTest(clashProxy, url, SpeedTimeout)
	if err != nil {
		return err
	}
	return nil
}
