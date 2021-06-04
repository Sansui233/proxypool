package cache

import (
	"time"

	"github.com/Sansui233/proxypool/pkg/proxy"
	"github.com/patrickmn/go-cache"
)

var c = cache.New(cache.NoExpiration, 10*time.Minute)

func GetProxies(key string) proxy.ProxyList {
	result, found := c.Get(key)
	if found {
		return result.(proxy.ProxyList) //Get返回的是interface
	}
	return nil
}

func SetProxies(key string, proxies proxy.ProxyList) {
	c.Set(key, proxies, cache.NoExpiration)
}

func SetString(key, value string) {
	c.Set(key, value, cache.NoExpiration)
}

func GetString(key string) string {
	result, found := c.Get(key)
	if found {
		return result.(string)
	}
	return ""
}
