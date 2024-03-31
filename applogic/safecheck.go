package applogic

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
)

// ResponseData 用于解析外层响应
type ResponseData struct {
	ConversationID string `json:"conversationId"`
	MessageID      string `json:"messageId"`
	Response       string `json:"response"` // 这里是嵌套的JSON字符串
}

// NestedResponse 用于解析嵌套的response字符串
type NestedResponse struct {
	Result float64 `json:"result"`
}

// checkResponseThreshold 发送消息并根据返回值决定是否超过阈值
func checkResponseThreshold(msg string) bool {
	url := config.GetAntiPromptAttackPath()
	requestBody, err := json.Marshal(map[string]interface{}{
		"message":         msg,
		"conversationId":  "",
		"parentMessageId": "",
		"user_id":         "",
	})
	if err != nil {
		fmtf.Printf("Error marshalling request: %v\n", err)
		return false
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmtf.Printf("Error sending request: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmtf.Printf("Error reading response body: %v\n", err)
		return false
	}
	fmtf.Printf("Response: %s\n", string(responseBody))

	var responseData ResponseData
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		fmtf.Printf("Error unmarshalling response data: %v\n", err)
		return false
	}

	var nestedResponse NestedResponse
	if err := json.Unmarshal([]byte(responseData.Response), &nestedResponse); err != nil {
		fmtf.Printf("Error unmarshalling nested response data: %v\n", err)
		return false
	}
	fmtf.Printf("大模型agent安全检查结果: %v\n", nestedResponse.Result)
	return nestedResponse.Result > config.GetAntiPromptLimit()
}
