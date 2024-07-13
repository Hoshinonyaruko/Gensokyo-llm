package applogic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/hunyuan"
	"github.com/hoshinonyaruko/gensokyo-llm/prompt"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
	"github.com/hoshinonyaruko/gensokyo-llm/utils"
)

var (
	groupUserMessages            sync.Map
	mutexhunyuan                 sync.Mutex
	lastCompleteResponseshunyuan sync.Map // 存储每个conversationId的完整累积信息
)

func (app *App) ChatHandlerHunyuan(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取访问者的IP地址
	ip := r.RemoteAddr             // 注意：这可能包含端口号
	ip = strings.Split(ip, ":")[0] // 去除端口号，仅保留IP地址

	// 获取IP白名单
	whiteList := config.IPWhiteList()

	// 检查IP是否在白名单中
	if !utils.Contains(whiteList, ip) {
		http.Error(w, "Access denied", http.StatusInternalServerError)
		return
	}

	var msg structs.Message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 读取URL参数 "prompt"
	promptstr := r.URL.Query().Get("prompt")
	if promptstr != "" {
		// prompt 参数存在，可以根据需要进一步处理或记录
		fmtf.Printf("Received prompt parameter: %s\n", promptstr)
	}

	msg.Role = "user"
	//颠倒用户输入
	if config.GetReverseUserPrompt() {
		msg.Text = utils.ReverseString(msg.Text)
	}

	if msg.ConversationID == "" {
		msg.ConversationID = utils.GenerateUUID()
		app.createConversation(msg.ConversationID)
	}

	userMessageID, err := app.addMessage(msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var history []structs.Message

	//根据是否有prompt参数 选择是否载入config.yml的prompt还是prompts文件夹的
	if promptstr == "" {
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

		// 分别获取FirstQ&A, SecondQ&A, ThirdQ&A
		pairs := []struct {
			Q     string
			A     string
			RoleQ string // 问题的角色
			RoleA string // 答案的角色
		}{
			{config.GetFirstQ(), config.GetFirstA(), "user", "assistant"},
			{config.GetSecondQ(), config.GetSecondA(), "user", "assistant"},
			{config.GetThirdQ(), config.GetThirdA(), "user", "assistant"},
		}

		// 检查每一对Q&A是否均不为空，并追加到历史信息中
		for _, pair := range pairs {
			if pair.Q != "" && pair.A != "" {
				qMessage := structs.Message{
					Text: pair.Q,
					Role: pair.RoleQ,
				}
				aMessage := structs.Message{
					Text: pair.A,
					Role: pair.RoleA,
				}

				// 注意追加的顺序，确保问题在答案之前
				history = append(history, qMessage, aMessage)
			}
		}
	} else {
		// 只获取系统提示词
		systemMessage, err := prompt.GetFirstSystemMessageStruct(promptstr)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			// 如果找到system消息，将其添加到历史数组中
			history = append(history, systemMessage)
			fmt.Println("Added system message back to history.")
		}
	}

	// 获取历史信息
	if msg.ParentMessageID != "" {
		userhistory, err := app.getHistory(msg.ConversationID, msg.ParentMessageID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 截断历史信息
		userHistory := truncateHistoryHunYuan(userhistory, msg.Text, promptstr)

		if promptstr != "" {
			// 注意追加的顺序，确保问题在系统提示词之后
			// 使用...操作符来展开userhistory切片并追加到history切片
			// 获取系统级预埋的系统自定义QA对
			systemHistory, err := prompt.GetMessagesExcludingSystem(promptstr)
			if err != nil {
				fmtf.Printf("Error getting system history: %v,promptstr[%v]\n", err, promptstr)
				return
			}

			// 处理增强QA逻辑
			if config.GetEnhancedQA(promptstr) {
				systemHistory, err := prompt.GetMessagesExcludingSystem(promptstr)
				if err != nil {
					fmt.Printf("Error getting system history: %v\n", err)
					return
				}

				// 计算需要补足的历史记录数量
				neededHistoryCount := len(systemHistory) - 2 // 最后两条留给当前QA处理
				if neededHistoryCount > len(userHistory) {
					// 补足用户或助手历史
					difference := neededHistoryCount - len(userHistory)
					for i := 0; i < difference; i++ {
						if i%2 != 0 {
							userHistory = append(userHistory, structs.Message{Text: "", Role: "user"})
						} else {
							userHistory = append(userHistory, structs.Message{Text: "", Role: "assistant"})
						}
					}
				}

				// 附加系统历史到用户或助手历史，除了最后两条
				for i := 0; i < len(systemHistory)-2; i++ {
					sysMsg := systemHistory[i]
					index := len(userHistory) - neededHistoryCount + i
					if index >= 0 && index < len(userHistory) {
						userHistory[index].Text += fmt.Sprintf(" (%s)", sysMsg.Text)
					}
				}
			} else {
				// 将系统级别QA简单的附加在用户对话前方的位置(ai会知道,但不会主动引导)
				history = append(history, systemHistory...)
			}

			// 留下最后一个systemHistory成员进行后续处理
		}

		// 添加用户历史到总历史中
		history = append(history, userHistory...)
	} else {
		var systemHistory []structs.Message
		if promptstr != "" {
			systemHistory, err = prompt.GetMessagesExcludingSystem(promptstr)
			if err != nil {
				fmtf.Printf("prompt.GetMessagesExcludingSystem error: %v\n", err)
			}
			history = append(history, systemHistory...)
		}
	}

	fmtf.Printf("混元上下文history:%v\n", history)

	if config.GetHunyuanType() == 0 {
		// 构建 hunyuan 请求
		request := hunyuan.NewChatProRequest()
		// 配置块
		request.StreamModeration = new(bool)
		*request.StreamModeration = config.GetHunyuanStreamModeration(promptstr)
		request.Stream = new(bool)
		*request.Stream = config.GetHunyuanStreamModeration(promptstr)
		request.TopP = new(float64)
		*request.TopP = config.GetTopPHunyuan(promptstr)
		request.Temperature = new(float64)
		*request.Temperature = config.GetTemperatureHunyuan(promptstr)

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
			http.Error(w, fmtf.Sprintf("hunyuanapi返回错误: %v", err), http.StatusInternalServerError)
			return
		}
		if config.GetuseSse(promptstr) < 2 {
			// 解析响应
			var responseTextBuilder strings.Builder
			var totalUsage structs.UsageInfo
			for event := range response.BaseSSEResponse.Events {
				if event.Err != nil {
					http.Error(w, fmtf.Sprintf("接收事件时发生错误: %v", event.Err), http.StatusInternalServerError)
					return
				}

				// 解析事件数据
				var eventData map[string]interface{}
				if err := json.Unmarshal(event.Data, &eventData); err != nil {
					http.Error(w, fmtf.Sprintf("解析事件数据出错: %v", err), http.StatusInternalServerError)
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
					fmtf.Fprintf(w, "data: %s\n\n", fmtf.Sprintf("接收事件时发生错误: %v", event.Err))
					flusher.Flush()
					continue
				}

				// 解析事件数据和提取信息
				var eventData map[string]interface{}
				if err := json.Unmarshal(event.Data, &eventData); err != nil {
					fmtf.Fprintf(w, "data: %s\n\n", fmtf.Sprintf("解析事件数据出错: %v", err))
					flusher.Flush()
					continue
				}

				responseText, usageInfo := utils.ExtractEventDetails(eventData)
				responseTextBuilder.WriteString(responseText)
				totalUsage.PromptTokens += usageInfo.PromptTokens
				totalUsage.CompletionTokens += usageInfo.CompletionTokens

				// 发送当前事件的响应数据，但不包含assistantMessageID
				//fmtf.Printf("发送当前事件的响应数据，但不包含assistantMessageID\n")
				tempResponseMap := map[string]interface{}{
					"response":       responseText,
					"conversationId": msg.ConversationID,
					"details": map[string]interface{}{
						"usage": usageInfo,
					},
				}
				tempResponseJSON, _ := json.Marshal(tempResponseMap)
				fmtf.Fprintf(w, "data: %s\n\n", string(tempResponseJSON))
				flusher.Flush()
			}

			// 处理完所有事件后，生成并发送包含assistantMessageID的最终响应
			responseText := responseTextBuilder.String()
			fmtf.Printf("处理完所有事件后,生成并发送包含assistantMessageID的最终响应:%v\n", responseText)
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
			fmtf.Fprintf(w, "data: %s\n\n", string(finalResponseJSON))
			flusher.Flush()
		}
	} else if config.GetHunyuanType() == 1 {
		// 构建 hunyuan 标准版请求
		request := hunyuan.NewChatStdRequest()

		// 配置块
		request.StreamModeration = new(bool)
		*request.StreamModeration = config.GetHunyuanStreamModeration(promptstr)
		request.Stream = new(bool)
		*request.Stream = config.GetHunyuanStreamModeration(promptstr)
		request.TopP = new(float64)
		*request.TopP = config.GetTopPHunyuan(promptstr)
		request.Temperature = new(float64)
		*request.Temperature = config.GetTemperatureHunyuan(promptstr)

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
		utils.PrintChatStdRequest(request)

		// 发送请求并获取响应
		response, err := app.Client.ChatStd(request)
		if err != nil {
			http.Error(w, fmtf.Sprintf("hunyuanapi返回错误: %v", err), http.StatusInternalServerError)
			return
		}
		if config.GetuseSse(promptstr) < 2 {
			// 解析响应
			var responseTextBuilder strings.Builder
			var totalUsage structs.UsageInfo
			for event := range response.BaseSSEResponse.Events {
				if event.Err != nil {
					http.Error(w, fmtf.Sprintf("接收事件时发生错误: %v", event.Err), http.StatusInternalServerError)
					return
				}

				// 解析事件数据
				var eventData map[string]interface{}
				if err := json.Unmarshal(event.Data, &eventData); err != nil {
					http.Error(w, fmtf.Sprintf("解析事件数据出错: %v", err), http.StatusInternalServerError)
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
					fmtf.Fprintf(w, "data: %s\n\n", fmtf.Sprintf("接收事件时发生错误: %v", event.Err))
					flusher.Flush()
					continue
				}

				// 解析事件数据和提取信息
				var eventData map[string]interface{}
				if err := json.Unmarshal(event.Data, &eventData); err != nil {
					fmtf.Fprintf(w, "data: %s\n\n", fmtf.Sprintf("解析事件数据出错: %v", err))
					flusher.Flush()
					continue
				}

				responseText, usageInfo := utils.ExtractEventDetails(eventData)
				responseTextBuilder.WriteString(responseText)
				totalUsage.PromptTokens += usageInfo.PromptTokens
				totalUsage.CompletionTokens += usageInfo.CompletionTokens

				// 发送当前事件的响应数据，但不包含assistantMessageID
				//fmtf.Printf("发送当前事件的响应数据，但不包含assistantMessageID\n")
				tempResponseMap := map[string]interface{}{
					"response":       responseText,
					"conversationId": msg.ConversationID,
					"details": map[string]interface{}{
						"usage": usageInfo,
					},
				}
				tempResponseJSON, _ := json.Marshal(tempResponseMap)
				fmtf.Fprintf(w, "data: %s\n\n", string(tempResponseJSON))
				flusher.Flush()
			}

			// 处理完所有事件后，生成并发送包含assistantMessageID的最终响应
			responseText := responseTextBuilder.String()
			fmtf.Printf("处理完所有事件后,生成并发送包含assistantMessageID的最终响应:%v\n", responseText)
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
			fmtf.Fprintf(w, "data: %s\n\n", string(finalResponseJSON))
			flusher.Flush()

		}
	} else if config.GetHunyuanType() == 2 || config.GetHunyuanType() == 3 || config.GetHunyuanType() == 4 || config.GetHunyuanType() == 5 {
		// 构建 hunyuan 请求
		request := hunyuan.NewChatCompletionsRequest()
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

		// 获取HunyuanType并设置对应的Model
		switch config.GetHunyuanType() {
		case 2:
			request.Model = new(string)
			*request.Model = "hunyuan-lite"
		case 3:
			request.Model = new(string)
			*request.Model = "hunyuan-standard"
		case 4:
			request.Model = new(string)
			*request.Model = "hunyuan-standard-256K"
		case 5:
			request.Model = new(string)
			*request.Model = "hunyuan-pro"
		default:
			request.Model = new(string)
			*request.Model = "default-value"
		}
		fmtf.Printf("请求的混元模型类型:%v", *request.Model)
		request.StreamModeration = new(bool)
		*request.StreamModeration = config.GetHunyuanStreamModeration(promptstr)
		request.Stream = new(bool)
		*request.Stream = config.GetHunyuanStreamModeration(promptstr)
		request.TopP = new(float64)
		*request.TopP = config.GetTopPHunyuan(promptstr)
		request.Temperature = new(float64)
		*request.Temperature = config.GetTemperatureHunyuan(promptstr)

		// 打印请求以进行调试
		utils.PrintChatCompletionsRequest(request)

		// 发送请求并获取响应
		response, err := app.Client.ChatCompletions(request)
		if err != nil {
			http.Error(w, fmtf.Sprintf("hunyuanapi返回错误: %v", err), http.StatusInternalServerError)
			return
		}
		if config.GetuseSse(promptstr) < 2 {
			// 解析响应
			var responseTextBuilder strings.Builder
			var totalUsage structs.UsageInfo
			for event := range response.BaseSSEResponse.Events {
				if event.Err != nil {
					http.Error(w, fmtf.Sprintf("接收事件时发生错误: %v", event.Err), http.StatusInternalServerError)
					return
				}

				// 解析事件数据
				var eventData map[string]interface{}
				if err := json.Unmarshal(event.Data, &eventData); err != nil {
					http.Error(w, fmtf.Sprintf("解析事件数据出错: %v", err), http.StatusInternalServerError)
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

			// 生成一个随机的UUID
			randomUUID, err := uuid.NewRandom()
			if err != nil {
				http.Error(w, "Failed to generate UUID", http.StatusInternalServerError)
				return
			}

			var totalUsage structs.UsageInfo // 有并发问题
			for event := range response.BaseSSEResponse.Events {

				if event.Err != nil {
					fmtf.Fprintf(w, "data: %s\n\n", fmtf.Sprintf("接收事件时发生错误: %v", event.Err))
					flusher.Flush()
					continue
				}

				// 解析事件数据和提取信息
				var eventData map[string]interface{}
				if err := json.Unmarshal(event.Data, &eventData); err != nil {
					fmtf.Fprintf(w, "data: %s\n\n", fmtf.Sprintf("解析事件数据出错: %v", err))
					flusher.Flush()
					continue
				}

				// 在修改共享资源之前锁定Mutex
				mutexhunyuan.Lock()
				// 由于同一个上下文中 msg.ConversationID是相同的,而我们要区分更细粒度 所以添加UUID openai的api则设计了更细粒度的stramid,可以直接使用
				conversationId := msg.ConversationID + randomUUID.String()
				// 读取完整信息
				completeResponse, _ := lastCompleteResponseshunyuan.LoadOrStore(conversationId, "")
				// 提取出本次请求的响应
				responseText, usageInfo := utils.ExtractEventDetails(eventData)
				// 更新存储的完整累积信息
				updatedCompleteResponse := completeResponse.(string) + responseText
				lastCompleteResponseshunyuan.Store(conversationId, updatedCompleteResponse)
				// 完成修改后解锁Mutex
				mutexhunyuan.Unlock()

				totalUsage.PromptTokens += usageInfo.PromptTokens
				totalUsage.CompletionTokens += usageInfo.CompletionTokens

				// 构建并发送当前事件的响应数据
				tempResponseMap := map[string]interface{}{
					"response":       responseText,
					"conversationId": msg.ConversationID,
					"details": map[string]interface{}{
						"usage": usageInfo,
					},
				}
				tempResponseJSON, _ := json.Marshal(tempResponseMap)
				fmtf.Fprintf(w, "data: %s\n\n", string(tempResponseJSON))
				flusher.Flush()
			}

			// 处理完所有事件后，生成并发送包含assistantMessageID的最终响应
			conversationId := msg.ConversationID + randomUUID.String()
			completeResponse, _ := lastCompleteResponseshunyuan.LoadOrStore(conversationId, "")
			fmtf.Printf("处理完所有事件后,生成并发送包含assistantMessageID的最终响应:%v\n", completeResponse.(string))
			assistantMessageID, err := app.addMessage(structs.Message{
				ConversationID:  msg.ConversationID,
				ParentMessageID: userMessageID,
				Text:            completeResponse.(string),
				Role:            "assistant",
			})

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			finalResponseMap := map[string]interface{}{
				"response":       completeResponse.(string),
				"conversationId": msg.ConversationID,
				"messageId":      assistantMessageID,
				"details": map[string]interface{}{
					"usage": totalUsage,
				},
			}
			finalResponseJSON, _ := json.Marshal(finalResponseMap)
			fmtf.Fprintf(w, "data: %s\n\n", string(finalResponseJSON))
			flusher.Flush()
		}
	}
}

func truncateHistoryHunYuan(history []structs.Message, prompt string, promptstr string) []structs.Message {
	MAX_TOKENS := config.GetMaxTokensHunyuan(promptstr)

	tokenCount := len(prompt)
	for _, msg := range history {
		tokenCount += len(msg.Text)
	}

	if tokenCount >= MAX_TOKENS {
		// 第一步：逐个移除消息直到满足令牌数量限制，同时保证成对的消息交替出现
		for tokenCount > MAX_TOKENS && len(history) > 0 {
			tokenCount -= len(history[0].Text)
			history = history[1:]

			// 确保移除后，历史记录仍然以user消息结尾
			if len(history) > 0 && history[0].Role == "assistant" {
				tokenCount -= len(history[0].Text)
				history = history[1:]
			}
		}
	}

	// 第二步：检查并移除包含空文本的QA对
	i := 0
	for i < len(history)-1 { // 需要检查成对的消息，所以用len(history)-1
		// 检查是否成对，且下一条消息的角色正确
		if history[i].Role == "user" && history[i+1].Role == "assistant" && (len(history[i].Text) == 0 || len(history[i+1].Text) == 0) {
			// 移除这对QA
			fmtf.Println("hunyuan-找到了空的对话: ", history[i].Text, history[i+1].Text)
			history = append(history[:i], history[i+2:]...)
			continue // 继续检查下一对，不增加i因为切片已经缩短
		}
		i++
	}

	// 第三步：确保以assistant结尾，如果不是则尝试移除直到满足条件
	if len(history) > 0 && history[len(history)-1].Role == "user" {
		// 尝试找到最近的"assistant"消息并截断至该点
		for i := len(history) - 2; i >= 0; i-- {
			if history[i].Role == "assistant" {
				history = history[:i+1]
				break
			}
		}
	}

	return history
}
