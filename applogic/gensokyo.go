package applogic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
	"github.com/hoshinonyaruko/gensokyo-llm/utils"
)

func (app *App) GensokyoHandler(w http.ResponseWriter, r *http.Request) {
	// 只处理POST请求
	if r.Method != http.MethodPost {
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

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// 解析请求体到OnebotGroupMessage结构体
	var message structs.OnebotGroupMessage
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
			utils.SendGroupMessage(message.GroupID, "重置成功")

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
							key := utils.GetKey(message.GroupID, message.UserID)
							accumulatedMessage, exists := groupUserMessages[key]

							// 提取response字段
							if response, ok := responseData["response"].(string); ok {
								// 如果accumulatedMessage是response的子串，则提取新的部分并发送
								if exists && strings.HasPrefix(response, accumulatedMessage) {
									newPart := response[len(accumulatedMessage):]
									if newPart != "" {
										fmt.Printf("A完整信息: %s,已发送信息:%s", response, accumulatedMessage)
										utils.SendGroupMessage(message.GroupID, newPart)
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
					utils.SendGroupMessage(message.GroupID, response)
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
