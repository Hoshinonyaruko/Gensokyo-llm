package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
)

// GetPromptkeyboard 请求并打印3个预测的问题
func GetPromptkeyboard(msg string) bool {
	url := config.GetFunctionPath()
	wxFunction := structs.WXFunction{
		Name:        "predict_followup_questions",
		Description: "根据用户输入,预测用户可能接下来提出的三个相关问题",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"question": map[string]interface{}{
					"type":        "string",
					"description": "用户提出的初始问题",
				},
			},
			"required": []string{"question"},
		},
		Responses: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"followup_questions": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "预测的后续问题列表",
				},
			},
		},
	}

	request := structs.WXRequestMessageF{
		Text:       msg,
		WXFunction: wxFunction,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error marshalling request: %v\n", err)
		return false
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return false
	}
	fmt.Printf("Response: %s\n", string(responseBody))

	// 这里可以添加逻辑以解析和处理响应数据

	return true // 根据实际情况可能需要调整返回值
}
