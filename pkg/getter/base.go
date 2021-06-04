package getter

import (
	"errors"
	"sync"

	"github.com/Sansui233/proxypool/pkg/proxy"
	"github.com/Sansui233/proxypool/pkg/tool"
)

// functions for getters
type Getter interface {
	Get() proxy.ProxyList
	Get2Chan(pc chan proxy.Proxy)
	Get2ChanWG(pc chan proxy.Proxy, wg *sync.WaitGroup)
}

// function type that creates getters
type creator func(options tool.Options) (getter Getter, err error)

// map str sourceType -> func creating getters,
// registered in package init()
var creatorMap = make(map[string]creator)

func Register(sourceType string, c creator) {
	creatorMap[sourceType] = c
}

func NewGetter(sourceType string, options tool.Options) (getter Getter, err error) {
	c, ok := creatorMap[sourceType]
	if ok {
		return c(options)
	}
	return nil, ErrorCreaterNotSupported
}

func StringArray2ProxyArray(origin []string) proxy.ProxyList {
	results := make(proxy.ProxyList, 0)
	for _, link := range origin {
		p, err := proxy.ParseProxyFromLink(link)
		if err == nil && p != nil {
			results = append(results, p)
		}
	}
	return results
}

func ClashProxy2ProxyArray(origin []map[string]interface{}) proxy.ProxyList {
	results := make(proxy.ProxyList, 0, len(origin))
	for _, pjson := range origin {
		p, err := proxy.ParseProxyFromClashProxy(pjson)
		if err == nil && p != nil {
			results = append(results, p)
		}
	}
	return results
}

func GrepLinksFromString(text string) []string {
	results := proxy.GrepSSRLinkFromString(text)
	results = append(results, proxy.GrepVmessLinkFromString(text)...)
	results = append(results, proxy.GrepSSLinkFromString(text)...)
	results = append(results, proxy.GrepTrojanLinkFromString(text)...)
	return results
}

func FuzzParseProxyFromString(text string) proxy.ProxyList {
	return StringArray2ProxyArray(GrepLinksFromString(text))
}

var (
	ErrorUrlNotFound         = errors.New("url should be specified")
	ErrorCreaterNotSupported = errors.New("type not supported")
)

func AssertTypeStringNotNull(i interface{}) (str string, err error) {
	switch i.(type) {
	case string:
		str = i.(string)
		if str == "" {
			return "", errors.New("string is null")
		}
		return str, nil
	default:
		return "", errors.New("type is not string")
	}
	return "", nil
}
