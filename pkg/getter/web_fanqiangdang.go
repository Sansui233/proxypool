package getter

import (
	"fmt"
	"github.com/Sansui233/proxypool/log"
	"strings"
	"sync"

	"github.com/Sansui233/proxypool/pkg/proxy"
	"github.com/Sansui233/proxypool/pkg/tool"
	"github.com/gocolly/colly"
)

func init() {
	Register("web-fanqiangdang", NewWebFanqiangdangGetter)
}

type WebFanqiangdang struct {
	c       *colly.Collector
	Url     string
	results proxy.ProxyList
}

func NewWebFanqiangdangGetter(options tool.Options) (getter Getter, err error) {
	urlInterface, found := options["url"]
	if found {
		url, err := AssertTypeStringNotNull(urlInterface)
		if err != nil {
			return nil, err
		}
		return &WebFanqiangdang{
			c:   colly.NewCollector(),
			Url: url,
		}, nil
	}
	return nil, ErrorUrlNotFound
}

func (w *WebFanqiangdang) Get() proxy.ProxyList {
	w.results = make(proxy.ProxyList, 0)
	w.c.OnHTML("td.t_f", func(e *colly.HTMLElement) {
		innerHTML, err := e.DOM.Html()
		if err != nil {
			return
		}
		if strings.Contains(innerHTML, "data-cfemail") {
			decoded, err := tool.CFEmailDecode(tool.GetCFEmailPayload(innerHTML))
			if err == nil {
				e.Text = strings.ReplaceAll(e.Text, "[email protected]", decoded)
			}
		}
		w.results = append(w.results, FuzzParseProxyFromString(e.Text)...)
		subUrls := urlRe.FindAllString(e.Text, -1)
		for _, url := range subUrls {
			w.results = append(w.results, (&Subscribe{Url: url}).Get()...)
		}
	})

	w.c.OnHTML("th.new>a[href]", func(e *colly.HTMLElement) {
		url := e.Attr("href")
		if url == "javascript:;" {
			return
		}
		url, err := tool.CFScriptRedirect(url)
		if err == nil && url[0] == '/' {
			url = "https://fanqiangdang.com" + url
		}
		if strings.HasPrefix(url, "https://fanqiangdang.com/thread") {
			_ = e.Request.Visit(url)
		}
	})

	w.results = make(proxy.ProxyList, 0)
	err := w.c.Visit(w.Url)
	if err != nil {
		_ = fmt.Errorf("%s", err.Error())
	}

	return w.results
}

func (w *WebFanqiangdang) Get2ChanWG(pc chan proxy.Proxy, wg *sync.WaitGroup) {
	defer wg.Done()
	nodes := w.Get()
	log.Infoln("STATISTIC: Fanqiangdang\tcount=%d\turl=%s\n", len(nodes), w.Url)
	for _, node := range nodes {
		pc <- node
	}
}

func (w *WebFanqiangdang) Get2Chan(pc chan proxy.Proxy) {
	nodes := w.Get()
	log.Infoln("STATISTIC: Fanqiangdang\tcount=%d\turl=%s\n", len(nodes), w.Url)
	for _, node := range nodes {
		pc <- node
	}
}
