package applogic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/prompt"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
	"github.com/hoshinonyaruko/gensokyo-llm/utils"
)

// 用于存储每个conversationId的最后一条消息内容
var (
	// lastResponses 存储每个真实 conversationId 的最后响应文本
	lastResponsesTyqw         sync.Map
	lastCompleteResponsesTyqw sync.Map // 存储每个conversationId的完整累积信息
	mutexTyqw                 sync.Mutex
)

func (app *App) ChatHandlerTyqw(w http.ResponseWriter, r *http.Request) {
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
		systemPromptContent := config.SystemPrompt()
		if systemPromptContent != "0" {
			systemPrompt := structs.Message{
				Text: systemPromptContent,
				Role: "system",
			}
			// 将系统提示词添加到历史信息的开始
			history = append([]structs.Message{systemPrompt}, history...)
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
		// 默认执行 正常提示词顺序
		if !config.GetEnhancedQA(promptstr) {
			history, err = prompt.GetMessagesFromFilename(promptstr)
			if err != nil {
				fmtf.Printf("prompt.GetMessagesFromFilename error: %v\n", err)
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
	}

	// 获取历史信息
	if msg.ParentMessageID != "" {
		userhistory, err := app.getHistory(msg.ConversationID, msg.ParentMessageID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 截断历史信息
		userHistory := truncateHistoryTyqw(userhistory, msg.Text, promptstr)

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
	}

	fmtf.Printf("Tyqw上下文history:%v\n", history)

	// 构建请求到Tyqw API
	apiURL := config.GetTyqwApiPath(promptstr)
	fmtf.Printf("Tyqw请求地址:%v\n", apiURL)

	// 构造消息历史和当前消息
	messages := []map[string]interface{}{}
	for _, hMsg := range history {
		messages = append(messages, map[string]interface{}{
			"role":    hMsg.Role,
			"content": hMsg.Text,
		})
	}

	// 添加当前用户消息
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": msg.Text,
	})

	var isIncrementalOutput bool
	tyqwssetype := config.GetTyqwSseType(promptstr)
	if tyqwssetype == 1 {
		isIncrementalOutput = true
	} else if tyqwssetype == 2 {
		isIncrementalOutput = false
	}
	// 获取配置信息
	useSSE := config.GetuseSse(promptstr)
	apiKey := config.GetTyqwKey(promptstr)
	var requestBody map[string]interface{}
	if useSSE == 2 {
		// 构建请求体，根据提供的文档重新调整
		requestBody = map[string]interface{}{
			"parameters": map[string]interface{}{
				"max_tokens":         config.GetTyqwMaxTokens(promptstr),   // 最大生成的token数
				"temperature":        config.GetTyqwTemperature(promptstr), // 控制随机性和多样性的温度
				"top_p":              config.GetTyqwTopP(promptstr),        // 核采样方法的概率阈值
				"top_k":              config.GetTyqwTopK(promptstr),        // 采样候选集的大小
				"repetition_penalty": config.GetTyqwRepetitionPenalty(),    // 控制重复度的惩罚因子
				"stop":               config.GetTyqwStopTokens(),           // 停止标记
				"seed":               config.GetTyqwSeed(),                 // 随机数种子
				"result_format":      "message",                            // 返回结果的格式
				"enable_search":      config.GetTyqwEnableSearch(),         // 是否启用互联网搜索
				"incremental_output": isIncrementalOutput,                  // 是否使用增量SSE模式,使用增量模式会更快,rwkv和openai不支持增量模式
			},
			"model": config.GetTyqwModel(promptstr), // 指定对话模型
			"input": map[string]interface{}{
				"messages": messages, // 用户与模型的对话历史
			},
			"user_name":      config.GetTyqwUserName(),      // 用户名
			"assistant_name": config.GetTyqwAssistantName(), // 助手名
			"system_name":    config.GetTyqwSystemName(),    // 系统名
			"presystem":      config.GetTyqwPreSystem(),     // 预系统处理信息
		}
	} else {
		// 构建请求体，根据提供的文档重新调整
		requestBody = map[string]interface{}{
			"parameters": map[string]interface{}{
				"max_tokens":         config.GetTyqwMaxTokens(promptstr),   // 最大生成的token数
				"temperature":        config.GetTyqwTemperature(promptstr), // 控制随机性和多样性的温度
				"top_p":              config.GetTyqwTopP(promptstr),        // 核采样方法的概率阈值
				"top_k":              config.GetTyqwTopK(promptstr),        // 采样候选集的大小
				"repetition_penalty": config.GetTyqwRepetitionPenalty(),    // 控制重复度的惩罚因子
				"stop":               config.GetTyqwStopTokens(),           // 停止标记
				"seed":               config.GetTyqwSeed(),                 // 随机数种子
				"result_format":      "message",                            // 返回结果的格式
				"enable_search":      config.GetTyqwEnableSearch(),         // 是否启用互联网搜索
			},
			"model": config.GetTyqwModel(promptstr), // 指定对话模型
			"input": map[string]interface{}{
				"messages": messages, // 用户与模型的对话历史
			},
			"user_name":      config.GetTyqwUserName(),      // 用户名
			"assistant_name": config.GetTyqwAssistantName(), // 助手名
			"system_name":    config.GetTyqwSystemName(),    // 系统名
			"presystem":      config.GetTyqwPreSystem(),     // 预系统处理信息
		}
	}

	fmtf.Printf("tyqw requestBody :%v", requestBody)
	requestBodyJSON, _ := json.Marshal(requestBody)

	// 准备HTTP请求
	client := &http.Client{}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		http.Error(w, fmtf.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// 设置Content-Type和Authorization
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 根据是否使用SSE来设置Accept和X-DashScope-SSE
	if useSSE == 2 {
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("X-DashScope-SSE", "enable")
	}

	// 设置工作区
	workspace, _ := config.GetTyqworkspace() // 假设这是获取workspace的函数
	if workspace != "" {
		fmt.Println("X-DashScope-WorkSpace:", workspace)
		req.Header.Set("X-DashScope-WorkSpace", workspace)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmtf.Sprintf("Error sending request to tyqw API: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if config.GetuseSse(promptstr) < 2 {
		// 处理响应
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read response body: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Printf("TYQW 返回: %v", string(responseBody))

		var tyqwApiResponse struct {
			Output struct {
				Choices []struct {
					FinishReason string `json:"finish_reason"`
					Message      struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					} `json:"message"`
				} `json:"choices"`
			} `json:"output"`
			Usage struct {
				TotalTokens  int `json:"total_tokens"`
				OutputTokens int `json:"output_tokens"`
				InputTokens  int `json:"input_tokens"`
			} `json:"usage"`
			RequestID string `json:"request_id"`
		}

		if err := json.Unmarshal(responseBody, &tyqwApiResponse); err != nil {
			http.Error(w, fmt.Sprintf("Error unmarshaling response: %v", err), http.StatusInternalServerError)
			return
		}

		// 从API响应中获取回复文本
		if len(tyqwApiResponse.Output.Choices) > 0 {
			responseText := tyqwApiResponse.Output.Choices[0].Message.Content

			// 添加助理消息
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

			// 构造响应数据，包括回复文本、对话ID、消息ID，以及使用情况
			responseMap := map[string]interface{}{
				"response":       responseText,
				"conversationId": msg.ConversationID,
				"messageId":      assistantMessageID,
				"details": map[string]interface{}{
					"usage": structs.UsageInfo{
						PromptTokens:     tyqwApiResponse.Usage.InputTokens,  // 实际值
						CompletionTokens: tyqwApiResponse.Usage.OutputTokens, // 实际值
					},
				},
			}

			// 设置响应头部为JSON格式
			w.Header().Set("Content-Type", "application/json")
			// 将响应数据编码为JSON并发送
			if err := json.NewEncoder(w).Encode(responseMap); err != nil {
				http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "No response data available from TYQW API", http.StatusInternalServerError)
		}
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

		reader := bufio.NewReader(resp.Body)
		var totalUsage structs.GPTUsageInfo
		if config.GetTyqwSseType(promptstr) == 1 {
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break // 流结束
					}
					// 处理错误
					fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("读取流数据时发生错误: %v", err))
					flusher.Flush()
					continue
				}
				if strings.HasPrefix(line, "data:") {
					eventDataJSON := line[5:] // 去掉"data: "前缀
					// 解析JSON数据
					var eventData structs.TyqwSSEData
					if err := json.Unmarshal([]byte(eventDataJSON), &eventData); err != nil {
						fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("解析事件数据出错: %v", err))
						flusher.Flush()
						continue
					}

					// 如果存在需要发送的临时响应数据（例如，在事件流中间点）
					tempResponseMap := map[string]interface{}{
						"response":       eventData.Output.Choices[0].Message.Content,
						"conversationId": msg.ConversationID, // 确保msg.ConversationID已经定义并初始化
						// "details" 字段留待进一步处理，如有必要
					}
					tempResponseJSON, _ := json.Marshal(tempResponseMap)
					fmt.Fprintf(w, "data: %s\n\n", string(tempResponseJSON))
					flusher.Flush()

					// 维护累加信息,发送最后事件
					conversationId := msg.ConversationID + randomUUID.String()
					// 读取完整信息
					completeResponse, _ := lastCompleteResponsesTyqw.LoadOrStore(conversationId, "")
					// 更新存储的完整累积信息
					updatedCompleteResponse := completeResponse.(string) + eventData.Output.Choices[0].Message.Content
					lastCompleteResponsesTyqw.Store(conversationId, updatedCompleteResponse)
				}
			}
		} else {
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break // 流结束
					}
					fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("读取流数据时发生错误: %v", err))
					flusher.Flush()
					continue
				}

				if strings.HasPrefix(line, "data:") {
					eventDataJSON := line[5:] // 去掉"data: "前缀
					if eventDataJSON[0] != '{' {
						fmt.Println("非JSON数据,跳过:", eventDataJSON)
						continue
					}

					// 解析JSON数据
					var eventData structs.TyqwSSEData
					if err := json.Unmarshal([]byte(eventDataJSON), &eventData); err != nil {
						fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("解析事件数据出错: %v", err))
						flusher.Flush()
						continue
					}

					// 在修改共享资源之前锁定Mutex
					mutexTyqw.Lock()

					conversationId := msg.ConversationID + randomUUID.String()
					//读取完整信息
					completeResponse, _ := lastCompleteResponsesTyqw.LoadOrStore(conversationId, "")

					// 检索上一次的响应文本
					lastResponse, _ := lastResponsesTyqw.LoadOrStore(conversationId, "")
					lastResponseText := lastResponse.(string)

					newContent := ""
					for _, choice := range eventData.Output.Choices {
						content := choice.Message.Content
						// 如果新内容以旧内容开头
						if strings.HasPrefix(content, lastResponseText) {
							// 特殊情况：当新内容和旧内容完全相同时，处理逻辑应当与新内容不以旧内容开头时相同
							if content == lastResponseText {
								newContent += content
							} else {
								// 剔除旧内容部分，只保留新增的部分
								newContent += content[len(lastResponseText):]
							}
						} else {
							// 如果新内容不以旧内容开头，可能是并发情况下的新消息，直接使用新内容
							newContent += content
						}
					}

					// 更新存储的完整累积信息
					updatedCompleteResponse := completeResponse.(string) + newContent
					lastCompleteResponsesTyqw.Store(conversationId, updatedCompleteResponse)

					// 使用累加的新内容更新存储的最后响应状态
					if newContent != "" {
						lastResponsesTyqw.Store(conversationId, newContent)
					}

					// 完成修改后解锁Mutex
					mutexTyqw.Unlock()

					// 发送新增的内容
					if newContent != "" {
						tempResponseMap := map[string]interface{}{
							"response":       newContent,
							"conversationId": msg.ConversationID,
						}
						tempResponseJSON, _ := json.Marshal(tempResponseMap)
						fmt.Fprintf(w, "data: %s\n\n", string(tempResponseJSON))
						flusher.Flush()
					}
				}
			}
		}

		// 一点点奇怪的转换
		conversationId := msg.ConversationID + randomUUID.String()
		completeResponse, _ := lastCompleteResponsesTyqw.LoadOrStore(conversationId, "")
		// 在所有事件处理完毕后发送最终响应
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

		// 在所有事件处理完毕后发送最终响应
		finalResponseMap := map[string]interface{}{
			"response":       completeResponse.(string),
			"conversationId": conversationId,
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

func truncateHistoryTyqw(history []structs.Message, prompt string, promptstr string) []structs.Message {
	MAX_TOKENS := config.GetTyqwMaxTokens(promptstr)

	tokenCount := len(prompt)
	for _, msg := range history {
		tokenCount += len(msg.Text)
	}

	if tokenCount >= MAX_TOKENS {
		// 第一步：从开始逐个移除消息，直到满足令牌数量限制
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
	for i := 0; i < len(history)-1; i++ { // 使用len(history)-1是因为我们要检查成对的消息
		q := history[i]
		a := history[i+1]

		// 检查Q和A是否成对，且A的角色应为assistant，Q的角色为user，避免删除非QA对的消息
		if q.Role == "user" && a.Role == "assistant" && (len(q.Text) == 0 || len(a.Text) == 0) {
			fmtf.Println("closeai-找到了空的对话: ", q, a)
			// 移除这对QA
			history = append(history[:i], history[i+2:]...)
			i-- // 因为删除了元素，调整索引以正确检查下一个元素
		}
	}

	// 第三步：确保以assistant结尾
	if len(history) > 0 && history[len(history)-1].Role == "user" {
		for len(history) > 0 && history[len(history)-1].Role != "assistant" {
			history = history[:len(history)-1]
		}
	}

	return history
}
