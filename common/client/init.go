package client

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/hoshinonyaruko/gensokyo-llm/common/config"
	"github.com/hoshinonyaruko/gensokyo-llm/common/logger"
	llmconfig "github.com/hoshinonyaruko/gensokyo-llm/config"
)

var HTTPClient *http.Client
var ImpatientHTTPClient *http.Client
var UserContentRequestHTTPClient *http.Client

func Init() {
	proxyURLstr := llmconfig.GetProxy()
	if proxyURLstr != "" {
		logger.SysLog(fmt.Sprintf("using %s as proxy to fetch user content", config.UserContentRequestProxy))
		// 解析 proxyURL 字符串
		proxyURL, err := url.Parse(proxyURLstr)
		if err != nil {
			log.Fatalf("Invalid proxy URL: %v", err)
		}

		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		UserContentRequestHTTPClient = &http.Client{
			Transport: transport,
			Timeout:   time.Second * time.Duration(config.UserContentRequestTimeout),
		}
	} else {
		UserContentRequestHTTPClient = &http.Client{}
	}
	var transport http.RoundTripper
	if config.RelayProxy != "" {
		logger.SysLog(fmt.Sprintf("using %s as api relay proxy", config.RelayProxy))
		proxyURL, err := url.Parse(config.RelayProxy)
		if err != nil {
			logger.FatalLog(fmt.Sprintf("USER_CONTENT_REQUEST_PROXY set but invalid: %s", config.UserContentRequestProxy))
		}
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	if config.RelayTimeout == 0 {
		HTTPClient = &http.Client{
			Transport: transport,
		}
	} else {
		HTTPClient = &http.Client{
			Timeout:   time.Duration(config.RelayTimeout) * time.Second,
			Transport: transport,
		}
	}

	ImpatientHTTPClient = &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}
}
