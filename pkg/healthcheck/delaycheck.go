package healthcheck

import (
	"encoding/json"
	"fmt"
	"github.com/Sansui233/proxypool/pkg/proxy"
	"sync"
	"time"

	"github.com/ivpusic/grpool"

	"github.com/Dreamacro/clash/adapters/outbound"
)

const defaultURLTestTimeout = time.Second * 5

func CleanBadProxiesWithGrpool(proxies []proxy.Proxy) (cproxies []proxy.Proxy) {
	// Note: Grpool实现对go并发管理的封装，主要是在数据量大时减少内存占用，不会提高效率。
	pool := grpool.NewPool(500, 200)

	c := make(chan *Stat)
	defer close(c)
	m := sync.Mutex{}

	pool.WaitCount(len(proxies))
	// 线程：延迟测试，测试过程通过grpool的job并发
	go func() {
		for _, p := range proxies {
			pp := p // 捕获，否则job执行时是按当前的p测试的
			pool.JobQueue <- func() {
				defer pool.JobDone()
				delay, err := testDelay(pp)
				if err == nil && delay != 0 {
					m.Lock()
					if ps, ok := ProxyStats.Find(pp); ok {
						ps.UpdatePSDelay(delay)
						c <- ps
					} else {
						ps = &Stat{
							Id:    pp.Identifier(),
							Delay: delay,
						}
						ProxyStats = append(ProxyStats, *ps)
						c <- ps
					}
					m.Unlock()
				}
			}
		}
	}()
	done := make(chan struct{}) // 用于多线程的运行结束标识
	defer close(done)

	go func() {
		pool.WaitAll()
		pool.Release()
		done <- struct{}{}
	}()

	okMap := make(map[string]struct{})
	for { // Note: 无限循环，直到能读取到done
		select {
		case ps := <-c:
			if ps.Delay > 0 {
				okMap[ps.Id] = struct{}{}
			}
		case <-done:
			cproxies = make(proxy.ProxyList, 0, 500) // 定义返回的proxylist
			// check usable proxy
			for i, _ := range proxies {
				if _, ok := okMap[proxies[i].Identifier()]; ok {
					//cproxies = append(cproxies, p.Clone())
					cproxies = append(cproxies, proxies[i]) // 返回对GC不友好的指针看会怎么样
				}
			}
			return
		}
	}
}

// Return 0 for error
func testDelay(p proxy.Proxy) (delay uint16, err error) {
	pmap := make(map[string]interface{})
	err = json.Unmarshal([]byte(p.String()), &pmap)
	if err != nil {
		return
	}

	pmap["port"] = int(pmap["port"].(float64))
	if p.TypeName() == "vmess" {
		pmap["alterId"] = int(pmap["alterId"].(float64))
		if network, ok := pmap["network"]; ok && network.(string) == "h2" {
			return 0, nil // todo 暂无方法测试h2的延迟，clash对于h2的connection会阻塞
		}
	}

	clashProxy, err := outbound.ParseProxy(pmap)
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}

	sTime := time.Now()
	err = HTTPHeadViaProxy(clashProxy, "http://www.gstatic.com/generate_204")
	if err != nil {
		return 0, err
	}
	fTime := time.Now()
	delay = uint16(fTime.Sub(sTime) / time.Millisecond)

	return delay, nil
}
