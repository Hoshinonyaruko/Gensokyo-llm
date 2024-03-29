package applogic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/hunyuan"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
	"github.com/hoshinonyaruko/gensokyo-llm/utils"
)

var messageBuilder strings.Builder
var groupUserMessages = make(map[string]string)

func (app *App) ChatHandlerHunyuan(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg structs.Message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	msg.Role = "user"

	if msg.ConversationID == "" {
		msg.ConversationID = utils.GenerateUUID()
		app.createConversation(msg.ConversationID)
	}

	userMessageID, err := app.addMessage(msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取历史信息
	var history []structs.Message
	if msg.ParentMessageID != "" {
		history, err = app.getHistory(msg.ConversationID, msg.ParentMessageID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 截断历史信息
		history = truncateHistoryHunYuan(history, msg.Text)
	}

	// 获取系统提示词
	systemPromptContent := config.SystemPrompt() // 注意检查实际的函数名是否正确

	// 如果系统提示词不为空，则添加到历史信息的开始
	if systemPromptContent != "0" {
		systemPromptRole := "system"
		systemPromptMsg := structs.Message{
			Text: systemPromptContent,
			Role: systemPromptRole,
		}
		// 将系统提示作为第一条消息
		history = append([]structs.Message{systemPromptMsg}, history...)
	}

	fmt.Printf("history:%v\n", history)

	// 构建 hunyuan 请求
	request := hunyuan.NewChatProRequest()

	// 添加历史信息
	for _, hMsg := range history {
		content := hMsg.Text // 创建新变量
		role := hMsg.Role    // 创建新变量
		hunyuanMsg := hunyuan.Message{
			Content: &content, // 引用新变量的地址
			Role:    &role,    // 引用新变量的地址
		}
		request.Messages = append(request.Messages, &hunyuanMsg)
	}

	// 添加当前用户消息
	currentUserContent := msg.Text // 创建新变量
	currentUserRole := msg.Role    // 创建新变量
	currentUserMsg := hunyuan.Message{
		Content: &currentUserContent, // 引用新变量的地址
		Role:    &currentUserRole,    // 引用新变量的地址
	}
	request.Messages = append(request.Messages, &currentUserMsg)

	// 打印请求以进行调试
	utils.PrintChatProRequest(request)

	// 发送请求并获取响应
	response, err := app.Client.ChatPro(request)
	if err != nil {
		http.Error(w, fmt.Sprintf("hunyuanapi返回错误: %v", err), http.StatusInternalServerError)
		return
	}

	if !config.GetuseSse() {
		// 解析响应
		var responseTextBuilder strings.Builder
		var totalUsage structs.UsageInfo
		for event := range response.BaseSSEResponse.Events {
			if event.Err != nil {
				http.Error(w, fmt.Sprintf("接收事件时发生错误: %v", event.Err), http.StatusInternalServerError)
				return
			}

			// 解析事件数据
			var eventData map[string]interface{}
			if err := json.Unmarshal(event.Data, &eventData); err != nil {
				http.Error(w, fmt.Sprintf("解析事件数据出错: %v", err), http.StatusInternalServerError)
				return
			}

			// 使用extractEventDetails函数提取信息
			responseText, usageInfo := utils.ExtractEventDetails(eventData)
			responseTextBuilder.WriteString(responseText)
			totalUsage.PromptTokens += usageInfo.PromptTokens
			totalUsage.CompletionTokens += usageInfo.CompletionTokens
		}
		// 现在responseTextBuilder中的内容是所有AI助手回复的组合
		responseText := responseTextBuilder.String()

		assistantMessageID, err := app.addMessage(structs.Message{
			ConversationID:  msg.ConversationID,
			ParentMessageID: userMessageID,
			Text:            responseText,
			Role:            "assistant",
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 构造响应
		responseMap := map[string]interface{}{
			"response":       responseText,
			"conversationId": msg.ConversationID,
			"messageId":      assistantMessageID,
			"details": map[string]interface{}{
				"usage": totalUsage,
			},
		}

		json.NewEncoder(w).Encode(responseMap)
	} else {
		// 设置SSE相关的响应头部
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		var responseTextBuilder strings.Builder
		var totalUsage structs.UsageInfo

		for event := range response.BaseSSEResponse.Events {
			if event.Err != nil {
				fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("接收事件时发生错误: %v", event.Err))
				flusher.Flush()
				continue
			}

			// 解析事件数据和提取信息
			var eventData map[string]interface{}
			if err := json.Unmarshal(event.Data, &eventData); err != nil {
				fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("解析事件数据出错: %v", err))
				flusher.Flush()
				continue
			}

			responseText, usageInfo := utils.ExtractEventDetails(eventData)
			responseTextBuilder.WriteString(responseText)
			totalUsage.PromptTokens += usageInfo.PromptTokens
			totalUsage.CompletionTokens += usageInfo.CompletionTokens

			// 发送当前事件的响应数据，但不包含assistantMessageID
			//fmt.Printf("发送当前事件的响应数据，但不包含assistantMessageID\n")
			tempResponseMap := map[string]interface{}{
				"response":       responseText,
				"conversationId": msg.ConversationID,
				"details": map[string]interface{}{
					"usage": usageInfo,
				},
			}
			tempResponseJSON, _ := json.Marshal(tempResponseMap)
			fmt.Fprintf(w, "data: %s\n\n", string(tempResponseJSON))
			flusher.Flush()
		}

		// 处理完所有事件后，生成并发送包含assistantMessageID的最终响应
		//fmt.Printf("处理完所有事件后，生成并发送包含assistantMessageID的最终响应\n")
		responseText := responseTextBuilder.String()
		assistantMessageID, err := app.addMessage(structs.Message{
			ConversationID:  msg.ConversationID,
			ParentMessageID: userMessageID,
			Text:            responseText,
			Role:            "assistant",
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		finalResponseMap := map[string]interface{}{
			"response":       responseText,
			"conversationId": msg.ConversationID,
			"messageId":      assistantMessageID,
			"details": map[string]interface{}{
				"usage": totalUsage,
			},
		}
		finalResponseJSON, _ := json.Marshal(finalResponseMap)
		fmt.Fprintf(w, "data: %s\n\n", string(finalResponseJSON))
		flusher.Flush()
	}
}

func splitAndSendMessages(groupid int64, userid int64, line string) {
	// 提取JSON部分
	dataPrefix := "data: "
	jsonStr := strings.TrimPrefix(line, dataPrefix)

	// 解析JSON数据
	var sseData struct {
		Response string `json:"response"`
	}
	err := json.Unmarshal([]byte(jsonStr), &sseData)
	if err != nil {
		fmt.Printf("Error unmarshalling SSE data: %v\n", err)
		return
	}

	// 处理提取出的信息
	processMessage(groupid, userid, sseData.Response)
}

func processMessage(groupid int64, userid int64, message string) {
	key := utils.GetKey(groupid, userid)

	// 定义中文全角和英文标点符号
	punctuations := []rune{'。', '！', '？', '，', ',', '.', '!', '?'}

	for _, char := range message {
		messageBuilder.WriteRune(char)
		if utils.ContainsRune(punctuations, char) {
			// 达到标点符号，发送累积的整个消息
			if messageBuilder.Len() > 0 {
				groupUserMessages[key] += messageBuilder.String()
				utils.SendGroupMessage(groupid, messageBuilder.String())
				messageBuilder.Reset() // 重置消息构建器
			}
		}
	}
}
