package tencent

import (
	"sort"

	relaymodel "github.com/hoshinonyaruko/gensokyo-llm/relay/model"
)

var ModelList = []string{
	"hunyuan-lite",
	"hunyuan-standard",
	"hunyuan-standard-256K",
	"hunyuan-pro",
}

// getModelNameByHunyuanType - 根据 hunyuanType 获取模型名称
func GetModelNameByHunyuanType(hunyuanType int) string {
	switch hunyuanType {
	case 0:
		return "hunyuan-standard"
	case 1:
		return "hunyuan-standard"
	case 2:
		return "hunyuan-lite"
	case 3:
		return "hunyuan-standard"
	case 4:
		return "hunyuan-standard-256K"
	case 5:
		return "hunyuan-pro"
	default:
		return "hunyuan-default"
	}
}

// FilterSystemMessages - 保留字数最多的一个 system 角色
func FilterSystemMessages(messages []interface{}) []relaymodel.Message {
	var systemMessages []relaymodel.Message
	var otherMessages []relaymodel.Message

	// 分离 system 角色和其他角色
	for _, msg := range messages {
		if messageMap, ok := msg.(map[string]interface{}); ok {
			role, okRole := messageMap["role"].(string)
			content, okContent := messageMap["content"]
			if okRole && okContent {
				message := relaymodel.Message{
					Role:    role,
					Content: content, // 注意 Content 是 any 类型，可能需要进一步的类型断言处理
				}

				if role == "system" {
					systemMessages = append(systemMessages, message)
				} else {
					otherMessages = append(otherMessages, message)
				}
			}
		}
	}

	// 保留字数最多的一个 system 角色
	if len(systemMessages) > 0 {
		sort.Slice(systemMessages, func(i, j int) bool {
			iContent, okI := systemMessages[i].Content.(string)
			jContent, okJ := systemMessages[j].Content.(string)
			if okI && okJ {
				return len(iContent) > len(jContent)
			}
			return false
		})
		otherMessages = append([]relaymodel.Message{systemMessages[0]}, otherMessages...)
	}

	return otherMessages
}

func AdjustMessageOrder(messages []relaymodel.Message) []relaymodel.Message {
	var adjustedMessages []relaymodel.Message
	userStarted := false

	// 查找第一个 user 或 assistant 消息的索引
	var firstUserIndex int = -1
	var firstAssistantIndex int = -1
	for i, msg := range messages {
		if msg.Role == "user" && firstUserIndex == -1 {
			firstUserIndex = i
			break
		}
		if msg.Role == "assistant" && firstAssistantIndex == -1 {
			firstAssistantIndex = i
		}
	}

	// 确保从 user 消息开始
	if firstUserIndex == -1 && firstAssistantIndex != -1 {
		// 如果没有 user 消息但有 assistant 消息，则移除第一个 assistant 消息
		if firstAssistantIndex == 0 {
			messages = messages[1:]
		}
	} else {
		userStarted = true
	}

	// 如果从 user 开始，按顺序重建消息列表
	if userStarted {
		for i := firstUserIndex; i < len(messages); i++ {
			if i%2 == 0 && messages[i].Role == "user" || i%2 == 1 && messages[i].Role == "assistant" {
				adjustedMessages = append(adjustedMessages, messages[i])
			}
		}
	}

	return adjustedMessages
}
