package applogic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

	// 预处理响应数据，移除可能的换行符
	preprocessedResponse := strings.TrimSpace(responseData.Response)

	// 尝试直接解析JSON
	err = json.Unmarshal([]byte(preprocessedResponse), &nestedResponse)

	// 如果直接解析失败，尝试容错处理
	if err != nil {
		// 检查是否为纯浮点数字符串，尝试解析为浮点数
		var floatValue float64
		if err := json.Unmarshal([]byte(preprocessedResponse), &floatValue); err == nil {
			// 如果是纯浮点数，构造JSON格式字符串并重新尝试解析
			jsonFloat := fmt.Sprintf("{\"result\":%s}", preprocessedResponse)
			err = json.Unmarshal([]byte(jsonFloat), &nestedResponse)
			if err != nil {
				// 如果仍然失败，则记录错误并返回
				fmt.Printf("Error unmarshalling adjusted response data: %v\n", err)
				return false
			}
		} else {
			// 如果不是纯浮点数，也不是正确的JSON格式，则记录原始错误并返回
			fmt.Printf("Error unmarshalling nested response data: %v\n", err)
			return false
		}
	}
	fmtf.Printf("大模型agent安全检查结果: %v\n", nestedResponse.Result)
	return nestedResponse.Result > config.GetAntiPromptLimit()
}
