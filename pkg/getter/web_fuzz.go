package getter

import (
	"github.com/Sansui233/proxypool/log"
	"io/ioutil"
	"sync"

	"github.com/Sansui233/proxypool/pkg/proxy"
	"github.com/Sansui233/proxypool/pkg/tool"
)

// Add key value pair to creatorMap(string → creator) in base.go
func init() {
	// register to creator map
	Register("webfuzz", NewWebFuzzGetter)
}

/* A Getter with an additional property */
type WebFuzz struct {
	Url string
}

// Implement Getter interface
func (w *WebFuzz) Get() proxy.ProxyList {
	resp, err := tool.GetHttpClient().Get(w.Url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	return FuzzParseProxyFromString(string(body))
}

func (w *WebFuzz) Get2ChanWG(pc chan proxy.Proxy, wg *sync.WaitGroup) {
	defer wg.Done()
	nodes := w.Get()
	log.Infoln("STATISTIC: WebFuzz\tcount=%d\turl=%s\n", len(nodes), w.Url)
	for _, node := range nodes {
		pc <- node
	}
}

func (w *WebFuzz) Get2Chan(pc chan proxy.Proxy) {
	nodes := w.Get()
	log.Infoln("STATISTIC: WebFuzz\tcount=%d\turl=%s\n", len(nodes), w.Url)
	for _, node := range nodes {
		pc <- node
	}
}

func NewWebFuzzGetter(options tool.Options) (getter Getter, err error) {
	urlInterface, found := options["url"]
	if found {
		url, err := AssertTypeStringNotNull(urlInterface)
		if err != nil {
			return nil, err
		}
		return &WebFuzz{Url: url}, nil
	}
	return nil, ErrorUrlNotFound
}
