package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/gensokyo-llm/common/logger"
	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/relay"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/adaptor/openai"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/adaptor/tencent"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/controller/validator"
	relaymodel "github.com/hoshinonyaruko/gensokyo-llm/relay/model"
)

func RelayTextHelper(c *gin.Context) *relaymodel.ErrorWithStatusCode {
	ctx := c.Request.Context()
	// get & validate textRequest
	textRequest, err := getAndValidateTextRequest(c)
	if err != nil {
		logger.Errorf(ctx, "getAndValidateTextRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "invalid_text_request", http.StatusBadRequest)
	}

	adaptor := relay.GetAdaptor(config.GetApiType())
	if adaptor == nil {
		return openai.ErrorWrapper(fmt.Errorf("invalid api type"), "invalid_api_type", http.StatusBadRequest)
	}

	adaptor.Init()

	if config.GetModelInterceptor() {
		// 读取并拦截请求体
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
			return openai.ErrorWrapper(err, "Failed to read request body", http.StatusInternalServerError)
		}

		// 解析 JSON
		var requestBodyTemp map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &requestBodyTemp); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
			return openai.ErrorWrapper(err, "Failed to read request body", http.StatusInternalServerError)
		}

		// 根据 API 类型修改 model 字段
		switch config.GetApiType() {
		case 0:
			// 根据 hunyuanType 修改 model 字段
			hunyuanType := config.GetHunyuanType()
			modelName := tencent.GetModelNameByHunyuanType(hunyuanType)
			requestBodyTemp["model"] = modelName
			textRequest.Model = modelName
			// 处理 messages 数组，保留字数最多的一个 system 角色
			if messages, ok := requestBodyTemp["messages"].([]interface{}); ok {
				filteredMessages := tencent.FilterSystemMessages(messages)
				//满足hunyuan的癖好
				filteredMessages = tencent.AdjustMessageOrder(filteredMessages)
				requestBodyTemp["Messages"] = filteredMessages
				textRequest.Messages = filteredMessages
			}
		case 2, 6, 7:
			requestBodyTemp["model"] = config.GetGptModel()
			textRequest.Model = config.GetGptModel()
		}

		// 将修改后的 JSON 转换回字节
		modifiedBodyBytes, err := json.Marshal(requestBodyTemp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal modified request body"})
			return openai.ErrorWrapper(err, "Failed to marshal modified request body", http.StatusInternalServerError)
		}

		// 重新设置 c.Request.Body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(modifiedBodyBytes))

		// 打印修改后的请求体
		//fmt.Printf("modifiedRequestBody: %s\n", string(modifiedBodyBytes))
	}

	// get request body
	var requestBody io.Reader
	switch config.GetApiType() {
	case 2, 6, 7:
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return openai.ErrorWrapper(err, "read_request_body_failed", http.StatusInternalServerError)
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		requestBody = bytes.NewBuffer(bodyBytes)
		//fmt.Printf("requestBody: %s\n", string(bodyBytes))
	default:
		convertedRequest, err := adaptor.ConvertRequest(c, textRequest)
		if err != nil {
			fmt.Printf("converted request err: %s\n", err)
			return openai.ErrorWrapper(err, "convert_request_failed", http.StatusInternalServerError)
		}
		jsonData, err := json.Marshal(convertedRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "json_marshal_failed", http.StatusInternalServerError)
		}
		fmt.Printf("converted request: %s\n", string(jsonData))
		requestBody = bytes.NewBuffer(jsonData)
	}

	// do request
	resp, err := adaptor.DoRequest(c, requestBody)
	if err != nil {
		fmt.Printf("DoRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	// 错误判断
	if isErrorHappened(resp) {
		return RelayErrorHandler(resp)
	}

	// do response
	_, respErr := adaptor.DoResponse(c, resp)
	if respErr != nil {
		fmt.Printf("respErr is not nil: %+v", respErr)
		return respErr
	}

	return nil
}

func getAndValidateTextRequest(c *gin.Context) (*relaymodel.GeneralOpenAIRequest, error) {
	textRequest := &relaymodel.GeneralOpenAIRequest{}
	err := UnmarshalBodyReusable(c, textRequest)
	if err != nil {
		return nil, err
	}
	err = validator.ValidateTextRequest(textRequest)
	if err != nil {
		return nil, err
	}
	return textRequest, nil
}

func UnmarshalBodyReusable(c *gin.Context, v any) error {
	requestBody, err := GetRequestBody(c)
	if err != nil {
		return err
	}
	contentType := c.Request.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		err = json.Unmarshal(requestBody, &v)
	}
	if err != nil {
		return err
	}
	// Reset request body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	return nil
}

func GetRequestBody(c *gin.Context) ([]byte, error) {
	requestBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	_ = c.Request.Body.Close()

	return requestBody, nil
}

func isErrorHappened(resp *http.Response) bool {
	if resp == nil {
		return true
	}
	if resp.StatusCode != http.StatusOK {
		return true
	}
	return false
}
