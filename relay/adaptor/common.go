package adaptor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/gensokyo-llm/common/client"
	"github.com/hoshinonyaruko/gensokyo-llm/config"
)

func SetupCommonRequestHeader(c *gin.Context, req *http.Request) {
	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))
	if config.GetuseSse() == 2 && c.Request.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "text/event-stream")
	}
}

func DoRequestHelper(a Adaptor, c *gin.Context, requestBody io.Reader) (*http.Response, error) {
	fullRequestURL, err := a.GetRequestURL()
	if err != nil {
		return nil, fmt.Errorf("get request url failed: %w", err)
	}
	fmt.Printf("请求地址:%v\n", fullRequestURL)
	fmt.Printf("请求体:%v\n", requestBody)
	req, err := http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("new request failed: %w", err)
	}
	err = a.SetupRequestHeader(c, req)
	if err != nil {
		return nil, fmt.Errorf("setup request header failed: %w", err)
	}
	resp, err := DoRequest(c, req)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}

	defer resp.Body.Close()
	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}
	// 重新设置 resp.Body 以便后续使用
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 打印响应体的 JSON 内容
	fmt.Printf("返回值: %s\n", string(bodyBytes))
	return resp, nil
}

func DoRequest(c *gin.Context, req *http.Request) (*http.Response, error) {
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("resp is nil")
	}
	_ = req.Body.Close()
	_ = c.Request.Body.Close()
	return resp, nil
}
