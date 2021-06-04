package api

import (
	binhtml "github.com/Sansui233/proxypool/internal/bindata/html"
	"github.com/Sansui233/proxypool/log"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sansui233/proxypool/config"
	appcache "github.com/Sansui233/proxypool/internal/cache"
	"github.com/Sansui233/proxypool/pkg/provider"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

const version = "v0.5.3"

var router *gin.Engine

func setupRouter() {
	gin.SetMode(gin.ReleaseMode)
	router = gin.New() // 没有任何中间件的路由
	store := persistence.NewInMemoryStore(time.Minute)
	router.Use(gin.Recovery(), cache.SiteCache(store, time.Minute)) // 加上处理panic的中间件，防止遇到panic退出程序

	_ = binhtml.RestoreAssets("", "assets/html") // 恢复静态文件（不恢复问题也不大就是难修改）
	_ = binhtml.RestoreAssets("", "assets/static")

	temp, err := loadHTMLTemplate() // 加载html模板，模板源存放于html.go中的类似_assetsHtmlSurgeHtml的变量
	if err != nil {
		panic(err)
	}
	router.SetHTMLTemplate(temp) // 应用模板

	router.StaticFile("/static/index.js", "assets/static/index.js")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/index.html", gin.H{
			"domain":               config.Config.Domain,
			"getters_count":        appcache.GettersCount,
			"all_proxies_count":    appcache.AllProxiesCount,
			"ss_proxies_count":     appcache.SSProxiesCount,
			"ssr_proxies_count":    appcache.SSRProxiesCount,
			"vmess_proxies_count":  appcache.VmessProxiesCount,
			"trojan_proxies_count": appcache.TrojanProxiesCount,
			"useful_proxies_count": appcache.UsefullProxiesCount,
			"last_crawl_time":      appcache.LastCrawlTime,
			"is_speed_test":        appcache.IsSpeedTest,
			"version":              version,
		})
	})

	router.GET("/clash", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/clash.html", gin.H{
			"domain": config.Config.Domain,
			"port":   config.Config.Port,
		})
	})

	router.GET("/surge", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/surge.html", gin.H{
			"domain": config.Config.Domain,
		})
	})

	router.GET("/shadowrocket", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/shadowrocket.html", gin.H{
			"domain": config.Config.Domain,
		})
	})

	router.GET("/clash/config", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/clash-config.yaml", gin.H{
			"domain": config.Config.Domain,
		})
	})
	router.GET("/clash/localconfig", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/clash-config-local.yaml", gin.H{
			"port": config.Config.Port,
		})
	})

	router.GET("/surge/config", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/surge.conf", gin.H{
			"domain": config.Config.Domain,
		})
	})

	router.GET("/clash/proxies", func(c *gin.Context) {
		proxyTypes := c.DefaultQuery("type", "")
		proxyCountry := c.DefaultQuery("c", "")
		proxyNotCountry := c.DefaultQuery("nc", "")
		proxySpeed := c.DefaultQuery("speed", "")
		text := ""
		if proxyTypes == "" && proxyCountry == "" && proxyNotCountry == "" && proxySpeed == "" {
			text = appcache.GetString("clashproxies") // A string. To show speed in this if condition, this must be updated after speedtest
			if text == "" {
				proxies := appcache.GetProxies("proxies")
				clash := provider.Clash{
					Base: provider.Base{
						Proxies: &proxies,
					},
				}
				text = clash.Provide() // 根据Query筛选节点
				appcache.SetString("clashproxies", text)
			}
		} else if proxyTypes == "all" {
			proxies := appcache.GetProxies("allproxies")
			clash := provider.Clash{
				provider.Base{
					Proxies:    &proxies,
					Types:      proxyTypes,
					Country:    proxyCountry,
					NotCountry: proxyNotCountry,
					Speed:      proxySpeed,
				},
			}
			text = clash.Provide() // 根据Query筛选节点
		} else {
			proxies := appcache.GetProxies("proxies")
			clash := provider.Clash{
				provider.Base{
					Proxies:    &proxies,
					Types:      proxyTypes,
					Country:    proxyCountry,
					NotCountry: proxyNotCountry,
					Speed:      proxySpeed,
				},
			}
			text = clash.Provide() // 根据Query筛选节点
		}
		c.String(200, text)
	})
	router.GET("/surge/proxies", func(c *gin.Context) {
		proxyTypes := c.DefaultQuery("type", "")
		proxyCountry := c.DefaultQuery("c", "")
		proxyNotCountry := c.DefaultQuery("nc", "")
		proxySpeed := c.DefaultQuery("speed", "")
		text := ""
		if proxyTypes == "" && proxyCountry == "" && proxyNotCountry == "" && proxySpeed == "" {
			text = appcache.GetString("surgeproxies") // A string. To show speed in this if condition, this must be updated after speedtest
			if text == "" {
				proxies := appcache.GetProxies("proxies")
				surge := provider.Surge{
					Base: provider.Base{
						Proxies: &proxies,
					},
				}
				text = surge.Provide()
				appcache.SetString("surgeproxies", text)
			}
		} else if proxyTypes == "all" {
			proxies := appcache.GetProxies("allproxies")
			surge := provider.Surge{
				Base: provider.Base{
					Proxies:    &proxies,
					Types:      proxyTypes,
					Country:    proxyCountry,
					NotCountry: proxyNotCountry,
					Speed:      proxySpeed,
				},
			}
			text = surge.Provide()
		} else {
			proxies := appcache.GetProxies("proxies")
			surge := provider.Surge{
				Base: provider.Base{
					Proxies:    &proxies,
					Types:      proxyTypes,
					Country:    proxyCountry,
					NotCountry: proxyNotCountry,
				},
			}
			text = surge.Provide()
		}
		c.String(200, text)
	})

	router.GET("/ss/sub", func(c *gin.Context) {
		proxies := appcache.GetProxies("proxies")
		ssSub := provider.SSSub{
			Base: provider.Base{
				Proxies: &proxies,
				Types:   "ss",
			},
		}
		c.String(200, ssSub.Provide())
	})
	router.GET("/ssr/sub", func(c *gin.Context) {
		proxies := appcache.GetProxies("proxies")
		ssrSub := provider.SSRSub{
			Base: provider.Base{
				Proxies: &proxies,
				Types:   "ssr",
			},
		}
		c.String(200, ssrSub.Provide())
	})
	router.GET("/vmess/sub", func(c *gin.Context) {
		proxies := appcache.GetProxies("proxies")
		vmessSub := provider.VmessSub{
			Base: provider.Base{
				Proxies: &proxies,
				Types:   "vmess",
			},
		}
		c.String(200, vmessSub.Provide())
	})
	router.GET("/sip002/sub", func(c *gin.Context) {
		proxies := appcache.GetProxies("proxies")
		sip002Sub := provider.SIP002Sub{
			Base: provider.Base{
				Proxies: &proxies,
				Types:   "ss",
			},
		}
		c.String(200, sip002Sub.Provide())
	})
	router.GET("/trojan/sub", func(c *gin.Context) {
		proxies := appcache.GetProxies("proxies")
		trojanSub := provider.TrojanSub{
			Base: provider.Base{
				Proxies: &proxies,
				Types:   "trojan",
			},
		}
		c.String(200, trojanSub.Provide())
	})
	router.GET("/link/:id", func(c *gin.Context) {
		idx := c.Param("id")
		proxies := appcache.GetProxies("allproxies")
		id, err := strconv.Atoi(idx)
		if err != nil {
			c.String(500, err.Error())
		}
		if id >= proxies.Len() || id < 0 {
			c.String(500, "id out of range")
		}
		c.String(200, proxies[id].Link())
	})
}

func Run() {
	setupRouter()
	servePort := config.Config.Port
	envp := os.Getenv("PORT") // environment port for heroku app
	if envp != "" {
		servePort = envp
	}
	// Run on this server
	err := router.Run(":" + servePort)
	if err != nil {
		log.Errorln("router: Web server starting failed. Make sure your port %s has not been used. \n%s", servePort, err.Error())
	} else {
		log.Infoln("Proxypool is serving on port: %s", servePort)
	}
}

// 返回页面templates
func loadHTMLTemplate() (t *template.Template, err error) {
	t = template.New("")
	for _, fileName := range binhtml.AssetNames() { //fileName带有路径前缀
		if strings.Contains(fileName, "css") {
			continue
		}
		data := binhtml.MustAsset(fileName)          //读取页面数据
		t, err = t.New(fileName).Parse(string(data)) //生成带路径名称的模板
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
