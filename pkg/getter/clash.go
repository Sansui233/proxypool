package getter

import (
	"github.com/Sansui233/proxypool/log"
	"github.com/Sansui233/proxypool/pkg/proxy"
	"github.com/Sansui233/proxypool/pkg/tool"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"sync"
)

func init() {
	Register("clash", NewClashGetter)
}

type Clash struct {
	Url string
}

type config struct {
	Proxy []map[string]interface{} `json:"proxies" yaml:"proxies"`
}

func (c *Clash) Get() proxy.ProxyList {
	resp, err := tool.GetHttpClient().Get(c.Url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	conf := config{}
	err = yaml.Unmarshal(body, &conf)
	if err != nil {
		return nil
	}

	return ClashProxy2ProxyArray(conf.Proxy)

}

func (c *Clash) Get2Chan(pc chan proxy.Proxy) {
	nodes := c.Get()
	log.Infoln("STATISTIC: Clash\tcount=%d\turl=%s\n", len(nodes), c.Url)
	for _, node := range nodes {
		pc <- node
	}
}

func (c *Clash) Get2ChanWG(pc chan proxy.Proxy, wg *sync.WaitGroup) {
	defer wg.Done()
	nodes := c.Get()
	log.Infoln("STATISTIC: Clash\tcount=%d\turl=%s\n", len(nodes), c.Url)
	for _, node := range nodes {
		pc <- node
	}
}

func NewClashGetter(options tool.Options) (getter Getter, err error) {
	urlInterface, found := options["url"]
	if found {
		url, err := AssertTypeStringNotNull(urlInterface)
		if err != nil {
			return nil, err
		}
		return &Clash{
			Url: url,
		}, nil
	}
	return nil, ErrorUrlNotFound
}
