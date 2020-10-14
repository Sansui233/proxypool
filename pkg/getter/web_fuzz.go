package getter

import (
	"io/ioutil"
	"log"
	"sync"

	"github.com/13263955567/proxypool/pkg/proxy"
	"github.com/13263955567/proxypool/pkg/tool"
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

func (w *WebFuzz) Get2Chan(pc chan proxy.Proxy, wg *sync.WaitGroup) {
	defer wg.Done()
	nodes := w.Get()
	log.Printf("STATISTIC: WebFuzz\tcount=%d\turl=%s\n", len(nodes), w.Url)
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
