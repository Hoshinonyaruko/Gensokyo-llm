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
	lastCompleteResponsesGlm sync.Map // 存储每个conversationId的完整累积信息
)

func (app *App) ChatHandlerGlm(w http.ResponseWriter, r *http.Request) {
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

	// 读取URL参数 "userid"
	useridstr := r.URL.Query().Get("userid")
	if useridstr != "" {
		// prompt 参数存在，可以根据需要进一步处理或记录
		fmtf.Printf("Received userid parameter: %s\n", useridstr)
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
		userHistory := truncateHistoryGlm(userhistory, msg.Text, promptstr)

		if promptstr != "" {
			// 注意追加的顺序，确保问题在系统提示词之后
			// 使用...操作符来展开userhistory切片并追加到history切片
			// 获取系统级预埋的系统自定义QA对
			systemHistory, err := prompt.GetMessagesExcludingSystem(promptstr)
			if err != nil {
				fmtf.Printf("Error getting system history: %v,promptstr[%v]\n", err, promptstr)
				return
			}
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

	fmtf.Printf("Glm上下文history:%v\n", history)

	// 构建请求到Glm API
	apiURL := config.GetGlmApiPath()

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

	if useridstr == "" {
		useridstr = "123"
	}

	// 获取配置信息
	apiKey := config.GetGlmApiKey(promptstr)
	// 创建请求体的映射结构
	requestBody := map[string]interface{}{
		"model":       config.GetGlmModel(promptstr),
		"messages":    messages,
		"do_sample":   config.GetGlmDoSample(),
		"stream":      config.GetuseSse(promptstr),
		"temperature": config.GetGlmTemperature(),
		"top_p":       config.GetGlmTopP(),
		"max_tokens":  config.GetGlmMaxTokens(promptstr),
		"stop":        config.GetGlmStop(),
		//"tools":       config.GetGlmTools(), 不太清楚参数格式
		"tool_choice": config.GetGlmToolChoice(),
		"user_id":     useridstr,
	}

	fmtf.Printf("glm requestBody :%v", requestBody)
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

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmtf.Sprintf("Error sending request to glm API: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if !config.GetuseSse(promptstr) {
		// 处理响应
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read response body: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Printf("glm 返回: %v", string(responseBody))

		var glmApiResponse struct {
			Choices []struct {
				FinishReason string `json:"finish_reason"`
				Message      struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
			Created   int64  `json:"created"`
			ID        string `json:"id"`
			Model     string `json:"model"`
			RequestID string `json:"request_id"`
			Usage     struct {
				CompletionTokens int `json:"completion_tokens"`
				PromptTokens     int `json:"prompt_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal(responseBody, &glmApiResponse); err != nil {
			http.Error(w, fmt.Sprintf("Error unmarshaling response: %v", err), http.StatusInternalServerError)
			return
		}

		// 从API响应中获取回复文本
		if len(glmApiResponse.Choices) > 0 {
			responseText := glmApiResponse.Choices[0].Message.Content

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
						PromptTokens:     glmApiResponse.Usage.PromptTokens,     // 实际值
						CompletionTokens: glmApiResponse.Usage.CompletionTokens, // 实际值
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
			http.Error(w, "No response data available from glm API", http.StatusInternalServerError)
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
				var eventData structs.GlmSSEData
				if err := json.Unmarshal([]byte(eventDataJSON), &eventData); err != nil {
					fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("解析事件数据出错: %v", err))
					flusher.Flush()
					continue
				}

				// 如果存在需要发送的临时响应数据（例如，在事件流中间点）
				tempResponseMap := map[string]interface{}{
					"response":       eventData.Choices[0].Delta.Content,
					"conversationId": msg.ConversationID, // 确保msg.ConversationID已经定义并初始化
					// "details" 字段留待进一步处理，如有必要
				}
				tempResponseJSON, _ := json.Marshal(tempResponseMap)
				fmt.Fprintf(w, "data: %s\n\n", string(tempResponseJSON))
				flusher.Flush()

				// 维护累加信息,发送最后事件
				conversationId := msg.ConversationID + randomUUID.String()
				// 读取完整信息
				completeResponse, _ := lastCompleteResponsesGlm.LoadOrStore(conversationId, "")
				// 更新存储的完整累积信息
				updatedCompleteResponse := completeResponse.(string) + eventData.Choices[0].Delta.Content
				lastCompleteResponsesGlm.Store(conversationId, updatedCompleteResponse)
			}
		}

		//一点点奇怪的转换
		conversationId := msg.ConversationID + randomUUID.String()
		completeResponse, _ := lastCompleteResponsesGlm.LoadOrStore(conversationId, "")
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
		// 首先从 conversationMap 获取真实的 conversationId
		if finalContent, ok := lastCompleteResponsesGlm.Load(conversationId); ok {
			finalResponseMap := map[string]interface{}{
				"response":       finalContent,
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

}

func truncateHistoryGlm(history []structs.Message, prompt string, promptstr string) []structs.Message {
	MAX_TOKENS := config.GetGlmMaxTokens(promptstr)

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
