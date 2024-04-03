package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/hoshinonyaruko/gensokyo-llm/acnode"
	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/hunyuan"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func PrintChatProRequest(request *hunyuan.ChatProRequest) {

	// 打印Messages
	for i, msg := range request.Messages {
		fmtf.Printf("Message %d:\n", i)
		fmtf.Printf("Content: %s\n", *msg.Content)
		fmtf.Printf("Role: %s\n", *msg.Role)
	}

}

func PrintChatStdRequest(request *hunyuan.ChatStdRequest) {

	// 打印Messages
	for i, msg := range request.Messages {
		fmtf.Printf("Message %d:\n", i)
		fmtf.Printf("Content: %s\n", *msg.Content)
		fmtf.Printf("Role: %s\n", *msg.Role)
	}

}

// contains 检查一个字符串切片是否包含一个特定的字符串
func Contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// 获取复合键
func GetKey(groupid int64, userid int64) string {
	return fmt.Sprintf("%d.%d", groupid, userid)
}

// 随机的分布发送
func ContainsRune(slice []rune, value rune) bool {
	for _, item := range slice {
		if item == value {
			// 获取概率百分比
			probability := config.GetSplitByPuntuations()
			// 将概率转换为0到1之间的浮点数
			probabilityPercentage := float64(probability) / 100.0
			// 生成一个0到1之间的随机浮点数
			randomValue := rand.Float64()
			// 如果随机数小于或等于概率，则返回true
			return randomValue <= probabilityPercentage
		}
	}
	return false
}

// 取出ai回答
func ExtractEventDetails(eventData map[string]interface{}) (string, structs.UsageInfo) {
	var responseTextBuilder strings.Builder
	var totalUsage structs.UsageInfo

	// 提取使用信息
	if usage, ok := eventData["Usage"].(map[string]interface{}); ok {
		var usageInfo structs.UsageInfo
		if promptTokens, ok := usage["PromptTokens"].(float64); ok {
			usageInfo.PromptTokens = int(promptTokens)
		}
		if completionTokens, ok := usage["CompletionTokens"].(float64); ok {
			usageInfo.CompletionTokens = int(completionTokens)
		}
		totalUsage.PromptTokens += usageInfo.PromptTokens
		totalUsage.CompletionTokens += usageInfo.CompletionTokens
	}

	// 提取AI助手的回复
	if choices, ok := eventData["Choices"].([]interface{}); ok {
		for _, choice := range choices {
			if choiceMap, ok := choice.(map[string]interface{}); ok {
				if delta, ok := choiceMap["Delta"].(map[string]interface{}); ok {
					if role, ok := delta["Role"].(string); ok && role == "assistant" {
						if content, ok := delta["Content"].(string); ok {
							responseTextBuilder.WriteString(content)
						}
					}
				}
			}
		}
	}

	return responseTextBuilder.String(), totalUsage
}

func SendGroupMessage(groupID int64, message string) error {
	// 获取基础URL
	baseURL := config.GetHttpPath() // 假设config.getHttpPath()返回基础URL

	// 构建完整的URL
	url := baseURL + "/send_group_msg"

	if config.GetSensitiveModeType() == 1 {
		message = acnode.CheckWordOUT(message)
	}

	// 构造请求体
	requestBody, err := json.Marshal(map[string]interface{}{
		"group_id": groupID,
		"message":  message,
	})
	if err != nil {
		return fmtf.Errorf("failed to marshal request body: %w", err)
	}

	// 发送POST请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmtf.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmtf.Errorf("received non-OK response status: %s", resp.Status)
	}

	// TODO: 处理响应体（如果需要）

	return nil
}

func SendPrivateMessage(UserID int64, message string) error {
	// 获取基础URL
	baseURL := config.GetHttpPath() // 假设config.getHttpPath()返回基础URL

	// 构建完整的URL
	url := baseURL + "/send_private_msg"

	if config.GetSensitiveModeType() == 1 {
		message = acnode.CheckWordOUT(message)
	}

	// 构造请求体
	requestBody, err := json.Marshal(map[string]interface{}{
		"user_id": UserID,
		"message": message,
	})

	if err != nil {
		return fmtf.Errorf("failed to marshal request body: %w", err)
	}

	// 发送POST请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmtf.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmtf.Errorf("received non-OK response status: %s", resp.Status)
	}

	// TODO: 处理响应体（如果需要）

	return nil
}

func SendPrivateMessageSSE(UserID int64, message structs.InterfaceBody) error {
	// 获取基础URL
	baseURL := config.GetHttpPath() // 假设config.GetHttpPath()返回基础URL

	// 构建完整的URL
	url := baseURL + "/send_private_msg_sse"

	// 检查是否需要启用敏感词过滤
	if config.GetSensitiveModeType() == 1 && message.Content != "" {
		message.Content = acnode.CheckWordOUT(message.Content)
	}

	// 构造请求体，包括InterfaceBody
	requestBody, err := json.Marshal(map[string]interface{}{
		"user_id": UserID,
		"message": message,
	})
	if err != nil {
		return fmtf.Errorf("failed to marshal request body: %w", err)
	}

	// 发送POST请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmtf.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmtf.Errorf("received non-OK response status: %s", resp.Status)
	}

	// TODO: 处理响应体（如果需要）

	return nil
}

// ReverseString 颠倒给定字符串中的字符顺序
func ReverseString(s string) string {
	// // 将字符串转换为rune切片，以便处理多字节字符
	// runes := []rune(s)
	// for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
	// 	// 交换前后对应的字符
	// 	runes[i], runes[j] = runes[j], runes[i]
	// }
	// // 将颠倒后的rune切片转换回字符串
	// return string(runes)
	return "####" + s + "####"
}

// RemoveBracketsContent 接收一个字符串，并移除所有[[...]]的内容
func RemoveBracketsContent(input string) string {
	// 编译一个正则表达式，用于匹配[[任意字符]]的模式
	re := regexp.MustCompile(`\[\[.*?\]\]`)
	// 使用正则表达式的ReplaceAllString方法删除匹配的部分
	return re.ReplaceAllString(input, "")
}
