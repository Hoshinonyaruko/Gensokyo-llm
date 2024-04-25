package applogic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/prompt"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
	"github.com/hoshinonyaruko/gensokyo-llm/utils"
)

//var mutexErnie sync.Mutex

func (app *App) ChatHandlerErnie(w http.ResponseWriter, r *http.Request) {
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

	// 是否从参数获取prompt
	if promptstr == "" {
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
		history, err = prompt.GetMessagesExcludingSystem(promptstr)
		if err != nil {
			fmtf.Printf("prompt.GetMessagesExcludingSystem error: %v\n", err)
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
		userhistory = truncateHistoryErnie(userhistory, msg.Text)

		// 注意追加的顺序，确保问题在系统提示词之后
		// 使用...操作符来展开userhistory切片并追加到history切片
		history = append(history, userhistory...)
	}

	fmtf.Printf("文心上下文history:%v\n", history)

	// 构建请求负载
	var payload structs.WXRequestPayload
	for _, hMsg := range history {
		payload.Messages = append(payload.Messages, structs.WXMessage{
			Content: hMsg.Text,
			Role:    hMsg.Role,
		})
	}

	// 添加当前用户消息
	payload.Messages = append(payload.Messages, structs.WXMessage{
		Content: msg.Text,
		Role:    "user",
	})

	TopP := config.GetWenxinTopp()
	PenaltyScore := config.GetWnxinPenaltyScore()
	MaxOutputTokens := config.GetWenxinMaxOutputTokens()

	// 设置其他可选参数
	payload.TopP = TopP
	payload.PenaltyScore = PenaltyScore
	payload.MaxOutputTokens = MaxOutputTokens

	// 是否sse
	if config.GetuseSse(promptstr) {
		payload.Stream = true
	}

	// 是否从参数中获取prompt
	if promptstr == "" {
		// 获取系统提示词，并设置system字段，如果它不为空
		systemPromptContent := config.SystemPrompt() // 确保函数名正确
		if systemPromptContent != "0" {
			payload.System = systemPromptContent // 直接在请求负载中设置system字段
		}
	} else {
		// 获取系统提示词，并设置system字段，如果它不为空
		systemPromptContent, err := prompt.GetFirstSystemMessage(promptstr)
		if err != nil {
			fmtf.Printf("prompt.GetFirstSystemMessage error: %v\n", err)
		}
		if systemPromptContent != "" {
			payload.System = systemPromptContent // 直接在请求负载中设置system字段
		}
	}

	// 获取访问凭证和API路径
	accessToken := config.GetWenxinAccessToken()
	apiPath := config.GetWenxinApiPath(promptstr)

	// 构建请求URL
	url := fmtf.Sprintf("%s?access_token=%s", apiPath, accessToken)
	fmtf.Printf("%v\n", url)

	// 序列化请求负载
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error occurred during marshaling. Error: %s", err.Error())
	}

	fmtf.Printf("文心一言请求:%v\n", string(jsonData))

	// 创建并发送POST请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error occurred during request creation. Error: %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error occurred during sending the request. Error: %s", err.Error())
	}
	defer resp.Body.Close()

	// 读取响应头中的速率限制信息
	rateLimitRequests := resp.Header.Get("X-Ratelimit-Limit-Requests")
	rateLimitTokens := resp.Header.Get("X-Ratelimit-Limit-Tokens")
	remainingRequests := resp.Header.Get("X-Ratelimit-Remaining-Requests")
	remainingTokens := resp.Header.Get("X-Ratelimit-Remaining-Tokens")

	fmtf.Printf("RateLimit: Requests %s, Tokens %s, Remaining Requests %s, Remaining Tokens %s\n",
		rateLimitRequests, rateLimitTokens, remainingRequests, remainingTokens)

	// 检查是否不使用SSE
	if !config.GetuseSse(promptstr) {
		// 读取整个响应体到内存中
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error occurred during response body reading. Error: %s", err)
		}

		// 首先尝试解析为简单的map来查看响应概览
		var response map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &response); err != nil {
			log.Fatalf("Error occurred during response decoding to map. Error: %s", err)
		}
		fmtf.Printf("%v\n", response)

		// 然后尝试解析为具体的结构体以获取详细信息
		var responseStruct struct {
			ID               string `json:"id"`
			Object           string `json:"object"`
			Created          int    `json:"created"`
			SentenceID       int    `json:"sentence_id,omitempty"`
			IsEnd            bool   `json:"is_end,omitempty"`
			IsTruncated      bool   `json:"is_truncated"`
			Result           string `json:"result"`
			NeedClearHistory bool   `json:"need_clear_history"`
			BanRound         int    `json:"ban_round"`
			Usage            struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal(bodyBytes, &responseStruct); err != nil {
			http.Error(w, fmtf.Sprintf("解析响应体出错: %v", err), http.StatusInternalServerError)
			return
		}
		// 根据API响应构造消息和响应给客户端
		assistantMessageID, err := app.addMessage(structs.Message{
			ConversationID:  msg.ConversationID,
			ParentMessageID: userMessageID,
			Text:            responseStruct.Result,
			Role:            "assistant",
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 构造响应
		responseMap := map[string]interface{}{
			"response":       responseStruct.Result,
			"conversationId": msg.ConversationID,
			"messageId":      assistantMessageID,
			"details": map[string]interface{}{
				"usage": map[string]int{
					"prompt_tokens":     responseStruct.Usage.PromptTokens,
					"completion_tokens": responseStruct.Usage.CompletionTokens,
					"total_tokens":      responseStruct.Usage.TotalTokens,
				},
			},
		}

		// 设置响应头信息以反映速率限制状态
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Ratelimit-Limit-Requests", rateLimitRequests)
		w.Header().Set("X-Ratelimit-Limit-Tokens", rateLimitTokens)
		w.Header().Set("X-Ratelimit-Remaining-Requests", remainingRequests)
		w.Header().Set("X-Ratelimit-Remaining-Tokens", remainingTokens)

		// 发送JSON响应
		json.NewEncoder(w).Encode(responseMap)
	} else {
		// SSE响应模式
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

		// 假设我们已经建立了与API的连接并且开始接收流式响应
		// reader代表从API接收数据的流
		reader := bufio.NewReader(resp.Body)
		for {
			// 读取流中的一行，即一个事件数据块
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// 流结束
					break
				}
				// 处理错误
				fmtf.Fprintf(w, "data: %s\n\n", fmtf.Sprintf("读取流数据时发生错误: %v", err))
				flusher.Flush()
				continue
			}

			// 处理流式数据行
			if strings.HasPrefix(line, "data: ") {
				eventDataJSON := line[6:] // 去掉"data: "前缀

				var eventData struct {
					ID               string `json:"id"`
					Object           string `json:"object"`
					Created          int    `json:"created"`
					SentenceID       int    `json:"sentence_id,omitempty"`
					IsEnd            bool   `json:"is_end,omitempty"`
					IsTruncated      bool   `json:"is_truncated"`
					Result           string `json:"result"`
					NeedClearHistory bool   `json:"need_clear_history"`
					BanRound         int    `json:"ban_round"`
					Usage            struct {
						PromptTokens     int `json:"prompt_tokens"`
						CompletionTokens int `json:"completion_tokens"`
						TotalTokens      int `json:"total_tokens"`
					} `json:"usage"`
				}
				// 解析JSON数据
				if err := json.Unmarshal([]byte(eventDataJSON), &eventData); err != nil {
					fmtf.Fprintf(w, "data: %s\n\n", fmtf.Sprintf("解析事件数据出错: %v", err))
					flusher.Flush()
					continue
				}

				// 这里处理解析后的事件数据
				responseTextBuilder.WriteString(eventData.Result)
				totalUsage.PromptTokens += eventData.Usage.PromptTokens
				totalUsage.CompletionTokens += eventData.Usage.CompletionTokens

				// 发送当前事件的响应数据，但不包含assistantMessageID
				tempResponseMap := map[string]interface{}{
					"response":       eventData.Result,
					"conversationId": msg.ConversationID,
					"details": map[string]interface{}{
						"usage": eventData.Usage,
					},
				}
				tempResponseJSON, _ := json.Marshal(tempResponseMap)
				fmtf.Fprintf(w, "data: %s\n\n", string(tempResponseJSON))
				flusher.Flush()

				// 如果这是最后一个消息
				if eventData.IsEnd {
					break
				}
			}
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

func truncateHistoryErnie(history []structs.Message, prompt string) []structs.Message {
	MAX_TOKENS := config.GetMaxTokenWenxin()

	tokenCount := len(prompt)
	for _, msg := range history {
		tokenCount += len(msg.Text)
	}

	if tokenCount >= MAX_TOKENS {
		// 第一步：逐个移除消息直到满足令牌数量限制
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
	for i := 0; i < len(history)-1; { // 注意这里去掉了自增部分
		if history[i].Role == "user" && history[i+1].Role == "assistant" &&
			(len(history[i].Text) == 0 || len(history[i+1].Text) == 0) {
			fmtf.Println("文心-找到了空的对话: ", history[i].Text, history[i+1].Text)
			history = append(history[:i], history[i+2:]...) // 移除这对QA
		} else {
			i++ // 只有在不删除元素时才增加i
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
