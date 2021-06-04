package getter

import (
	"fmt"
	"github.com/Sansui233/proxypool/log"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/Sansui233/proxypool/pkg/proxy"
	"github.com/Sansui233/proxypool/pkg/tool"
	"github.com/gocolly/colly"
)

func init() {
	Register("tgchannel", NewTGChannelGetter)
}

type TGChannelGetter struct {
	c         *colly.Collector
	NumNeeded int
	results   []string
	Url       string
	apiUrl    string
}

func NewTGChannelGetter(options tool.Options) (getter Getter, err error) {
	num, found := options["num"]
	t := 200
	switch num.(type) {
	case int:
		t = num.(int)
	case float64:
		t = int(num.(float64))
	}

	if !found || t <= 0 {
		t = 200
	}
	urlInterface, found := options["channel"]
	if found {
		url, err := AssertTypeStringNotNull(urlInterface)
		if err != nil {
			return nil, err
		}
		return &TGChannelGetter{
			c:         tool.GetColly(),
			NumNeeded: t,
			Url:       "https://t.me/s/" + url,
			apiUrl:    "https://tg.i-c-a.su/rss/" + url,
		}, nil
	}
	return nil, ErrorUrlNotFound
}

func (g *TGChannelGetter) Get() proxy.ProxyList {
	result := make(proxy.ProxyList, 0)
	g.results = make([]string, 0)
	// 找到所有的文字消息
	g.c.OnHTML("div.tgme_widget_message_text", func(e *colly.HTMLElement) {
		g.results = append(g.results, GrepLinksFromString(e.Text)...)
		// 抓取到http链接，有可能是订阅链接或其他链接，无论如何试一下
		subUrls := urlRe.FindAllString(e.Text, -1)
		for _, url := range subUrls {
			result = append(result, (&Subscribe{Url: url}).Get()...)
		}
	})

	// 找到之前消息页面的链接，加入访问队列
	g.c.OnHTML("link[rel=prev]", func(e *colly.HTMLElement) {
		if len(g.results) < g.NumNeeded {
			_ = e.Request.Visit(e.Attr("href"))
		}
	})

	g.results = make([]string, 0)
	err := g.c.Visit(g.Url)
	if err != nil {
		_ = fmt.Errorf("%s", err.Error())
	}
	result = append(result, StringArray2ProxyArray(g.results)...)

	// 获取文件(api需要维护)
	resp, err := tool.GetHttpClient().Get(g.apiUrl)
	if err != nil {
		return result
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	items := strings.Split(string(body), "\n")
	for _, s := range items {
		if strings.Contains(s, "enclosure url") { // get to xml node
			elements := strings.Split(s, "\"")
			for _, e := range elements {
				if strings.Contains(e, "https://") {
					// Webfuzz的可能性比较大，也有可能是订阅链接，为了不拖慢运行速度不写了
					result = append(result, (&WebFuzz{Url: e}).Get()...)
				}
			}
		}
	}
	return result
}

func (g *TGChannelGetter) Get2ChanWG(pc chan proxy.Proxy, wg *sync.WaitGroup) {
	defer wg.Done()
	nodes := g.Get()
	log.Infoln("STATISTIC: TGChannel\tcount=%d\turl=%s\n", len(nodes), g.Url)
	for _, node := range nodes {
		pc <- node
	}
}
func (g *TGChannelGetter) Get2Chan(pc chan proxy.Proxy) {
	nodes := g.Get()
	log.Infoln("STATISTIC: TGChannel\tcount=%d\turl=%s\n", len(nodes), g.Url)
	for _, node := range nodes {
		pc <- node
	}
}
