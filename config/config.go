package config

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sansui233/proxypool/pkg/tool"
	"github.com/ghodss/yaml"
)

var configFilePath = "config.yaml"

type ConfigOptions struct {
	Domain            string   `json:"domain" yaml:"domain"`
	Port              string   `json:"port" yaml:"port"`
	DatabaseUrl       string   `json:"database_url" yaml:"database_url"`
	CrawlInterval     uint64   `json:"crawl-interval" yaml:"crawl-interval"`
	CFEmail           string   `json:"cf_email" yaml:"cf_email"`
	CFKey             string   `json:"cf_key" yaml:"cf_key"`
	SourceFiles       []string `json:"source-files" yaml:"source-files"`
	SpeedTest         bool     `json:"speedtest" yaml:"speedtest"`
	SpeedTestInterval uint64   `json:"speedtest-interval" yaml:"speedtest-interval"`
	Connection        int      `json:"connection" yaml:"connection"`
	Timeout           int      `json:"timeout" yaml:"timeout"`
	ActiveFrequency   uint16   `json:"active-frequency" yaml:"active-frequency" `
	ActiveInterval    uint64   `json:"active-interval" yaml:"active-interval"`
	ActiveMaxNumber   uint16   `json:"active-max-number" yaml:"active-max-number"`
}

// Config 配置
var Config ConfigOptions

// Parse 解析配置文件，支持本地文件系统和网络链接
func Parse(path string) error {
	if path == "" {
		path = configFilePath
	} else {
		configFilePath = path
	}
	fileData, err := ReadFile(path)
	if err != nil {
		return err
	}
	Config = ConfigOptions{}
	err = yaml.Unmarshal(fileData, &Config)
	if err != nil {
		return err
	}

	// set default
	if Config.Connection <= 0 {
		Config.Connection = 5
	}
	if Config.Port == "" {
		Config.Port = "12580"
	}
	if Config.CrawlInterval == 0 {
		Config.CrawlInterval = 60
	}
	if Config.SpeedTestInterval == 0 {
		Config.SpeedTestInterval = 720
	}
	if Config.ActiveInterval == 0 {
		Config.ActiveInterval = 60
	}
	if Config.ActiveFrequency == 0 {
		Config.ActiveFrequency = 100
	}
	if Config.ActiveMaxNumber == 0 {
		Config.ActiveMaxNumber = 100
	}

	// 部分配置环境变量优先
	if domain := os.Getenv("DOMAIN"); domain != "" {
		Config.Domain = domain
	}
	if cfEmail := os.Getenv("CF_API_EMAIL"); cfEmail != "" {
		Config.CFEmail = cfEmail
	}
	if cfKey := os.Getenv("CF_API_KEY"); cfKey != "" {
		Config.CFKey = cfKey
	}

	return nil
}

// 从本地文件或者http链接读取配置文件内容
func ReadFile(path string) ([]byte, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		resp, err := tool.GetHttpClient().Get(path)
		if err != nil {
			return nil, errors.New("config file http get fail")
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	} else {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, err
		}
		return ioutil.ReadFile(path)
	}
}
