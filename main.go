package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3" // 只导入，作为驱动

	"github.com/google/uuid"
	"github.com/hoshinonyaruko/gensokyo-hunyuan/config"
	"github.com/hoshinonyaruko/gensokyo-hunyuan/hunyuan"
	"github.com/hoshinonyaruko/gensokyo-hunyuan/template"
)

var messageBuilder strings.Builder

var groupUserMessages = make(map[string]string)

type App struct {
	DB     *sql.DB
	Client *hunyuan.Client
}

type Message struct {
	ConversationID  string `json:"conversationId"`
	ParentMessageID string `json:"parentMessageId"`
	Text            string `json:"message"`
	Role            string `json:"role"`
	CreatedAt       string `json:"created_at"`
}
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

// 群信息事件
type OnebotGroupMessage struct {
	RawMessage      string      `json:"raw_message"`
	MessageID       int         `json:"message_id"`
	GroupID         int64       `json:"group_id"` // Can be either string or int depending on p.Settings.CompleteFields
	MessageType     string      `json:"message_type"`
	PostType        string      `json:"post_type"`
	SelfID          int64       `json:"self_id"` // Can be either string or int
	Sender          Sender      `json:"sender"`
	SubType         string      `json:"sub_type"`
	Time            int64       `json:"time"`
	Avatar          string      `json:"avatar,omitempty"`
	Echo            string      `json:"echo,omitempty"`
	Message         interface{} `json:"message"` // For array format
	MessageSeq      int         `json:"message_seq"`
	Font            int         `json:"font"`
	UserID          int64       `json:"user_id"`
	RealMessageType string      `json:"real_message_type,omitempty"`  //当前信息的真实类型 group group_private guild guild_private
	IsBindedGroupId bool        `json:"is_binded_group_id,omitempty"` //当前群号是否是binded后的
	IsBindedUserId  bool        `json:"is_binded_user_id,omitempty"`  //当前用户号号是否是binded后的
}

type Sender struct {
	Nickname string `json:"nickname"`
	TinyID   string `json:"tiny_id"`
	UserID   int64  `json:"user_id"`
	Role     string `json:"role,omitempty"`
	Card     string `json:"card,omitempty"`
	Sex      string `json:"sex,omitempty"`
	Age      int32  `json:"age,omitempty"`
	Area     string `json:"area,omitempty"`
	Level    string `json:"level,omitempty"`
	Title    string `json:"title,omitempty"`
}

func (app *App) ensureTablesExist() error {
	createMessagesTableSQL := `
    CREATE TABLE IF NOT EXISTS messages (
        id VARCHAR(36) PRIMARY KEY,
        conversation_id VARCHAR(36) NOT NULL,
        parent_message_id VARCHAR(36),
        text TEXT NOT NULL,
        role TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`

	_, err := app.DB.Exec(createMessagesTableSQL)
	if err != nil {
		return fmt.Errorf("error creating messages table: %w", err)
	}

	// 其他创建

	return nil
}
func (app *App) createConversation(conversationID string) error {
	_, err := app.DB.Exec("INSERT INTO conversations (id) VALUES (?)", conversationID)
	return err
}

func (app *App) addMessage(msg Message) (string, error) {
	fmt.Printf("添加信息：%v\n", msg)
	// Generate a new UUID for message ID
	messageID := generateUUID() // Implement this function to generate a UUID

	_, err := app.DB.Exec("INSERT INTO messages (id, conversation_id, parent_message_id, text, role) VALUES (?, ?, ?, ?, ?)",
		messageID, msg.ConversationID, msg.ParentMessageID, msg.Text, msg.Role)
	return messageID, err
}

func truncateHistory(history []Message, prompt string) []Message {
	const MAX_TOKENS = 4096

	tokenCount := len(prompt)
	for _, msg := range history {
		tokenCount += len(msg.Text)
	}

	if tokenCount <= MAX_TOKENS {
		return history
	}

	// 第一步：移除所有助手回复
	truncatedHistory := []Message{}
	for _, msg := range history {
		if msg.Role == "user" {
			truncatedHistory = append(truncatedHistory, msg)
		}
	}

	tokenCount = len(prompt)
	for _, msg := range truncatedHistory {
		tokenCount += len(msg.Text)
	}

	if tokenCount <= MAX_TOKENS {
		return truncatedHistory
	}

	// 第二步：从开始逐个移除消息，直到满足令牌数量限制
	for tokenCount > MAX_TOKENS && len(truncatedHistory) > 0 {
		tokenCount -= len(truncatedHistory[0].Text)
		truncatedHistory = truncatedHistory[1:]
	}

	return truncatedHistory
}

func (app *App) getHistory(conversationID, parentMessageID string) ([]Message, error) {
	var history []Message

	// SQL 查询获取历史信息
	query := `SELECT text, role, created_at FROM messages
              WHERE conversation_id = ? AND created_at <= (SELECT created_at FROM messages WHERE id = ?)
              ORDER BY created_at ASC`
	rows, err := app.DB.Query(query, conversationID, parentMessageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var previousText string
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.Text, &msg.Role, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		if msg.Text == previousText {
			continue
		}
		previousText = msg.Text

		// 根据角色添加不同的消息格式
		historyEntry := Message{
			Role: msg.Role,
			Text: msg.Text,
		}
		fmt.Printf("加入:%v\n", historyEntry)
		history = append(history, historyEntry)
	}
	return history, nil
}

func (app *App) chatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg Message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	msg.Role = "user"

	if msg.ConversationID == "" {
		msg.ConversationID = generateUUID()
		app.createConversation(msg.ConversationID)
	}

	userMessageID, err := app.addMessage(msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取历史信息
	var history []Message
	if msg.ParentMessageID != "" {
		history, err = app.getHistory(msg.ConversationID, msg.ParentMessageID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 截断历史信息
		history = truncateHistory(history, msg.Text)
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
	printChatProRequest(request)

	// 发送请求并获取响应
	response, err := app.Client.ChatPro(request)
	if err != nil {
		http.Error(w, fmt.Sprintf("hunyuanapi返回错误: %v", err), http.StatusInternalServerError)
		return
	}

	if !config.GetuseSse() {
		// 解析响应
		var responseTextBuilder strings.Builder
		var totalUsage UsageInfo
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
			responseText, usageInfo := extractEventDetails(eventData)
			responseTextBuilder.WriteString(responseText)
			totalUsage.PromptTokens += usageInfo.PromptTokens
			totalUsage.CompletionTokens += usageInfo.CompletionTokens
		}
		// 现在responseTextBuilder中的内容是所有AI助手回复的组合
		responseText := responseTextBuilder.String()

		assistantMessageID, err := app.addMessage(Message{
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
		var totalUsage UsageInfo

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

			responseText, usageInfo := extractEventDetails(eventData)
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
		assistantMessageID, err := app.addMessage(Message{
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

func generateUUID() string {
	return uuid.New().String()
}

func main() {
	if _, err := os.Stat("config.yml"); os.IsNotExist(err) {

		// 将修改后的配置写入 config.yml
		err = os.WriteFile("config.yml", []byte(template.ConfigTemplate), 0644)
		if err != nil {
			fmt.Println("Error writing config.yml:", err)
			return
		}

		fmt.Println("请配置config.yml然后再次运行.")
		fmt.Print("按下 Enter 继续...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		os.Exit(0)
	}
	// 加载配置
	conf, err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	// Deprecated
	secretId := conf.Settings.SecretId
	secretKey := conf.Settings.SecretKey
	fmt.Printf("secretId:%v\n", secretId)
	fmt.Printf("secretKey:%v\n", secretKey)
	region := config.Getregion()
	client, err := hunyuan.NewClientWithSecretId(secretId, secretKey, region)
	if err != nil {
		fmt.Printf("创建hunyuanapi出错:%v", err)
	}

	db, err := sql.Open("sqlite3", "file:mydb.sqlite?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := &App{DB: db, Client: client}

	// 在启动服务之前确保所有必要的表都已创建
	err = app.ensureTablesExist()
	if err != nil {
		log.Fatalf("Failed to ensure database tables exist: %v", err)
	}
	// 确保user_context表存在
	err = app.ensureUserContextTableExists()
	if err != nil {
		log.Fatalf("Failed to ensure user_context table exists: %v", err)
	}
	http.HandleFunc("/conversation", app.chatHandler)
	http.HandleFunc("/gensokyo", app.gensokyoHandler)
	port := config.GetPort()
	portStr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(portStr, nil))
	// request := hunyuan.NewChatProRequest()
	// var message hunyuan.Message
	// var messages []*hunyuan.Message

	// content := "你好"
	// role := "user"

	// message.Content = &content
	// message.Role = &role

	// messages = append(messages, &message)
	// request.Messages = messages

	// response, err := client.ChatPro(request)
	// if err != nil {
	// 	fmt.Printf("hunyuanapi返回错误:%v", err)
	// }
	// for event := range response.BaseSSEResponse.Events {
	// 	if event.Err != nil {
	// 		fmt.Printf("接收事件时发生错误：%v\n", event.Err)
	// 		continue
	// 	}

	// 	// 处理事件数据
	// 	fmt.Printf("收到事件：%s\n", event.Event)
	// 	fmt.Printf("事件ID：%s\n", event.Id)
	// 	fmt.Printf("事件数据：%s\n", string(event.Data))
	// }
}
func (app *App) ensureUserContextTableExists() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS user_context (
        user_id INTEGER PRIMARY KEY,
        conversation_id TEXT NOT NULL,
        parent_message_id TEXT
    );`

	_, err := app.DB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating user_context table: %w", err)
	}

	return nil
}

func (app *App) handleUserContext(userID int64) (string, string, error) {
	var conversationID, parentMessageID string

	// 检查用户上下文是否存在
	query := `SELECT conversation_id, parent_message_id FROM user_context WHERE user_id = ?`
	err := app.DB.QueryRow(query, userID).Scan(&conversationID, &parentMessageID)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户上下文不存在，创建新的
			conversationID = generateUUID() // 假设generateUUID()是一个生成UUID的函数
			parentMessageID = ""

			// 插入新的用户上下文
			insertQuery := `INSERT INTO user_context (user_id, conversation_id, parent_message_id) VALUES (?, ?, ?)`
			_, err = app.DB.Exec(insertQuery, userID, conversationID, parentMessageID)
			if err != nil {
				return "", "", err
			}
		} else {
			// 查询过程中出现了其他错误
			return "", "", err
		}
	}

	// 返回conversationID和parentMessageID
	return conversationID, parentMessageID, nil
}

func (app *App) migrateUserToNewContext(userID int64) error {
	// 生成新的conversationID
	newConversationID := generateUUID() // 假设generateUUID()是一个生成UUID的函数

	// 更新用户上下文
	updateQuery := `UPDATE user_context SET conversation_id = ?, parent_message_id = '' WHERE user_id = ?`
	_, err := app.DB.Exec(updateQuery, newConversationID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) gensokyoHandler(w http.ResponseWriter, r *http.Request) {
	// 只处理POST请求
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// 解析请求体到OnebotGroupMessage结构体
	var message OnebotGroupMessage
	err = json.Unmarshal(body, &message)
	if err != nil {
		http.Error(w, "Error parsing request body", http.StatusInternalServerError)
		return
	}

	// 打印消息和其他相关信息
	fmt.Printf("Received message: %v\n", message.Message)
	fmt.Printf("Full message details: %+v\n", message)

	// 判断message.Message的类型
	switch msg := message.Message.(type) {
	case string:
		// message.Message是一个string
		fmt.Printf("Received string message: %s\n", msg)
		switch msg {
		case "重置":
			fmt.Println("处理重置操作")
			app.migrateUserToNewContext(message.UserID)
			sendGroupMessage(message.GroupID, "重置成功")

		default:
			// 当msg不符合任何已定义case时的处理逻辑
			conversationID, parentMessageID, err := app.handleUserContext(message.UserID)
			//每句话清空上一句话的messageBuilder
			messageBuilder.Reset()
			fmt.Printf("conversationID: %s,parentMessageID%s\n", conversationID, parentMessageID)
			if err != nil {
				fmt.Printf("Error handling user context: %v\n", err)
				return
			}
			port := config.GetPort()
			// 构建并发送请求到conversation接口
			portStr := fmt.Sprintf(":%d", port)
			url := "http://127.0.0.1" + portStr + "/conversation"

			requestBody, err := json.Marshal(map[string]interface{}{
				"message":         message.Message,
				"conversationId":  conversationID,
				"parentMessageId": parentMessageID,
				"user_id":         message.UserID,
			})
			if err != nil {
				fmt.Printf("Error marshalling request: %v\n", err)
				return
			}

			resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
			if err != nil {
				fmt.Printf("Error sending request to conversation interface: %v\n", err)
				return
			}

			defer resp.Body.Close()

			var lastMessageID string

			if config.GetuseSse() {
				// 处理SSE流式响应
				reader := bufio.NewReader(resp.Body)
				for {
					line, err := reader.ReadBytes('\n')
					if err != nil {
						if err == io.EOF {
							break // 流结束
						}
						fmt.Printf("Error reading SSE response: %v\n", err)
						return
					}

					// 忽略空行
					if string(line) == "\n" {
						continue
					}

					// 处理接收到的数据
					fmt.Printf("Received SSE data: %s", string(line))

					// 去除"data: "前缀后进行JSON解析
					jsonData := strings.TrimPrefix(string(line), "data: ")
					var responseData map[string]interface{}
					if err := json.Unmarshal([]byte(jsonData), &responseData); err == nil {
						if id, ok := responseData["messageId"].(string); ok {
							lastMessageID = id // 更新lastMessageID
							// 检查是否有未发送的消息部分
							key := getKey(message.GroupID, message.UserID)
							accumulatedMessage, exists := groupUserMessages[key]

							// 提取response字段
							if response, ok := responseData["response"].(string); ok {
								// 如果accumulatedMessage是response的子串，则提取新的部分并发送
								if exists && strings.HasPrefix(response, accumulatedMessage) {
									newPart := response[len(accumulatedMessage):]
									if newPart != "" {
										fmt.Printf("A完整信息: %s,已发送信息:%s", response, accumulatedMessage)
										sendGroupMessage(message.GroupID, newPart)
									}
								} else if response != "" {
									// 如果accumulatedMessage不存在或不是子串，print
									fmt.Printf("B完整信息: %s,已发送信息:%s", response, accumulatedMessage)
								}

								// 清空映射中对应的累积消息
								groupUserMessages[key] = ""
							}
						} else {
							//发送信息
							fmt.Printf("发信息: %s", string(line))
							splitAndSendMessages(message.GroupID, message.UserID, string(line))
						}
					}

				}

				// 在SSE流结束后更新用户上下文
				if lastMessageID != "" {
					fmt.Printf("lastMessageID: %s\n", lastMessageID)
					err := app.updateUserContext(message.UserID, lastMessageID)
					if err != nil {
						fmt.Printf("Error updating user context: %v\n", err)
					}
				}
			} else {
				// 处理常规响应
				responseBody, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf("Error reading response body: %v\n", err)
					return
				}
				fmt.Printf("Response from conversation interface: %s\n", string(responseBody))

				// 使用map解析响应数据以获取response字段和messageId
				var responseData map[string]interface{}
				if err := json.Unmarshal(responseBody, &responseData); err != nil {
					fmt.Printf("Error unmarshalling response data: %v\n", err)
					return
				}

				// 使用提取的response内容发送消息
				if response, ok := responseData["response"].(string); ok && response != "" {
					sendGroupMessage(message.GroupID, response)
				}

				// 更新用户上下文
				if messageId, ok := responseData["messageId"].(string); ok {
					err := app.updateUserContext(message.UserID, messageId)
					if err != nil {
						fmt.Printf("Error updating user context: %v\n", err)
					}
				}
			}

			// 发送响应
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Request received and processed"))
		}

	case map[string]interface{}:
		// message.Message是一个map[string]interface{}
		fmt.Println("Received map message, handling not implemented yet")
		// 处理map类型消息的逻辑（TODO）

	default:
		// message.Message是一个未知类型
		fmt.Printf("Received message of unexpected type: %T\n", msg)
		return
	}

}

func (app *App) updateUserContext(userID int64, parentMessageID string) error {
	updateQuery := `UPDATE user_context SET parent_message_id = ? WHERE user_id = ?`
	_, err := app.DB.Exec(updateQuery, parentMessageID, userID)
	if err != nil {
		return err
	}
	return nil
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

func getKey(groupid int64, userid int64) string {
	return fmt.Sprintf("%d.%d", groupid, userid)
}

func processMessage(groupid int64, userid int64, message string) {
	key := getKey(groupid, userid)

	// 定义中文全角和英文标点符号
	punctuations := []rune{'。', '！', '？', '，', ',', '.', '!', '?'}

	for _, char := range message {
		messageBuilder.WriteRune(char)
		if containsRune(punctuations, char) {
			// 达到标点符号，发送累积的整个消息
			if messageBuilder.Len() > 0 {
				groupUserMessages[key] += messageBuilder.String()
				sendGroupMessage(groupid, messageBuilder.String())
				messageBuilder.Reset() // 重置消息构建器
			}
		}
	}
}

func containsRune(slice []rune, value rune) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func printChatProRequest(request *hunyuan.ChatProRequest) {

	// 打印Messages
	for i, msg := range request.Messages {
		fmt.Printf("Message %d:\n", i)
		fmt.Printf("Content: %s\n", *msg.Content)
		fmt.Printf("Role: %s\n", *msg.Role)
	}

}

func extractEventDetails(eventData map[string]interface{}) (string, UsageInfo) {
	var responseTextBuilder strings.Builder
	var totalUsage UsageInfo

	// 提取使用信息
	if usage, ok := eventData["Usage"].(map[string]interface{}); ok {
		var usageInfo UsageInfo
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

func sendGroupMessage(groupID int64, message string) error {
	// 获取基础URL
	baseURL := config.GetHttpPath() // 假设config.getHttpPath()返回基础URL

	// 构建完整的URL
	url := baseURL + "/send_group_msg"

	// 构造请求体
	requestBody, err := json.Marshal(map[string]interface{}{
		"group_id": groupID,
		"message":  message,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 发送POST请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	// TODO: 处理响应体（如果需要）

	return nil
}
