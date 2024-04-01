package applogic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/hoshinonyaruko/gensokyo-llm/acnode"
	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
	"github.com/hoshinonyaruko/gensokyo-llm/utils"
)

var newmsgToStringMap = make(map[string]string)
var stringToIndexMap = make(map[string]int)

// RecordStringById 根据id记录一个string
func RecordStringByNewmsg(id, value string) {
	newmsgToStringMap[id] = value
}

// GetStringById 根据newmsg取出对应的string
func GetStringByNewmsg(newmsg string) string {
	if value, exists := newmsgToStringMap[newmsg]; exists {
		return value
	}
	// 如果id不存在，返回空字符串
	return ""
}

// IncrementIndex 为给定的字符串递增索引
func IncrementIndex(s string) int {
	// 检查map中是否已经有这个字符串的索引
	if _, exists := stringToIndexMap[s]; !exists {
		// 如果不存在，初始化为0
		stringToIndexMap[s] = 0
	}
	// 递增索引
	stringToIndexMap[s]++
	// 返回新的索引值
	return stringToIndexMap[s]
}

// ResetIndex 将给定字符串的索引归零
func ResetIndex(s string) {
	stringToIndexMap[s] = 0
}

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
		fmtf.Printf("Error parsing request body: %+v\n", string(body))
		http.Error(w, "Error parsing request body", http.StatusInternalServerError)
		return
	}

	// 打印消息和其他相关信息
	fmtf.Printf("Received message: %v\n", message.Message)
	fmtf.Printf("Full message details: %+v\n", message)

	// 判断message.Message的类型
	switch msg := message.Message.(type) {
	case string:
		// message.Message是一个string
		fmtf.Printf("Received string message: %s\n", msg)

		//是否过滤群信息
		if !config.GetGroupmessage() {
			fmtf.Printf("你设置了不响应群信息：%v", message)
			return
		}

		// 从GetRestoreCommand获取重置指令的列表
		restoreCommands := config.GetRestoreCommand()

		checkResetCommand := msg
		if config.GetIgnoreExtraTips() {
			checkResetCommand = utils.RemoveBracketsContent(checkResetCommand)
		}

		// 检查checkResetCommand是否在restoreCommands列表中
		isResetCommand := false
		for _, command := range restoreCommands {
			if checkResetCommand == command {
				isResetCommand = true
				break
			}
		}

		//处理重置指令
		if isResetCommand {
			fmtf.Println("处理重置操作")
			app.migrateUserToNewContext(message.UserID)
			RestoreResponse := config.GetRandomRestoreResponses()
			if message.RealMessageType == "group_private" || message.MessageType == "private" {
				if !config.GetUsePrivateSSE() {
					utils.SendPrivateMessage(message.UserID, RestoreResponse)
				} else {

					// 将字符串转换为rune切片，以正确处理多字节字符
					runes := []rune(RestoreResponse)

					// 计算每部分应该包含的rune数量
					partLength := len(runes) / 3

					// 初始化用于存储分割结果的切片
					parts := make([]string, 3)

					// 按字符分割字符串
					for i := 0; i < 3; i++ {
						if i < 2 { // 前两部分
							start := i * partLength
							end := start + partLength
							parts[i] = string(runes[start:end])
						} else { // 最后一部分，包含所有剩余的字符
							start := i * partLength
							parts[i] = string(runes[start:])
						}
					}

					// 开头
					messageSSE := structs.InterfaceBody{
						Content: parts[0],
						State:   1,
					}

					utils.SendPrivateMessageSSE(message.UserID, messageSSE)

					//中间
					messageSSE = structs.InterfaceBody{
						Content: parts[1],
						State:   11,
					}
					utils.SendPrivateMessageSSE(message.UserID, messageSSE)

					// 从配置中获取promptkeyboard
					promptkeyboard := config.GetPromptkeyboard()

					// 创建InterfaceBody结构体实例
					messageSSE = structs.InterfaceBody{
						Content:        parts[2],       // 假设空格字符串是期望的内容
						State:          20,             // 假设的状态码
						PromptKeyboard: promptkeyboard, // 使用更新后的promptkeyboard
					}

					// 发送SSE私人消息
					utils.SendPrivateMessageSSE(message.UserID, messageSSE)
				}
			} else {
				utils.SendGroupMessage(message.GroupID, RestoreResponse)
			}
			return
		}

		//审核部分 文本替换规则
		newmsg := message.Message.(string)
		if config.GetSensitiveMode() {
			newmsg = acnode.CheckWord(newmsg)
		}

		//提示词安全部分
		if config.GetAntiPromptAttackPath() != "" {
			if config.GetIgnoreExtraTips() {
				newmsg = utils.RemoveBracketsContent(newmsg)
			}

			if checkResponseThreshold(newmsg) {
				fmtf.Printf("提示词不安全,过滤:%v", message)
				saveresponse := config.GetRandomSaveResponse()
				if saveresponse != "" {
					if message.RealMessageType == "group_private" || message.MessageType == "private" {
						if !config.GetUsePrivateSSE() {
							utils.SendPrivateMessage(message.UserID, saveresponse)
						} else {
							// 将字符串转换为rune切片，以正确处理多字节字符
							runes := []rune(saveresponse)

							// 计算每部分应该包含的rune数量
							partLength := len(runes) / 3

							// 初始化用于存储分割结果的切片
							parts := make([]string, 3)

							// 按字符分割字符串
							for i := 0; i < 3; i++ {
								if i < 2 { // 前两部分
									start := i * partLength
									end := start + partLength
									parts[i] = string(runes[start:end])
								} else { // 最后一部分，包含所有剩余的字符
									start := i * partLength
									parts[i] = string(runes[start:])
								}
							}
							// 开头
							messageSSE := structs.InterfaceBody{
								Content: parts[0],
								State:   1,
							}

							utils.SendPrivateMessageSSE(message.UserID, messageSSE)

							//中间
							messageSSE = structs.InterfaceBody{
								Content: parts[1],
								State:   11,
							}
							utils.SendPrivateMessageSSE(message.UserID, messageSSE)

							// 从配置中获取恢复响应数组
							RestoreResponses := config.GetRestoreCommand()

							var selectedRestoreResponse string
							// 如果RestoreResponses至少有一个成员，则随机选择一个
							if len(RestoreResponses) > 0 {
								selectedRestoreResponse = RestoreResponses[rand.Intn(len(RestoreResponses))]
							}

							// 从配置中获取promptkeyboard
							promptkeyboard := config.GetPromptkeyboard()

							// 确保promptkeyboard至少有一个成员
							if len(promptkeyboard) > 0 {
								// 使用随机选中的RestoreResponse替换promptkeyboard的第一个成员
								promptkeyboard[0] = selectedRestoreResponse
							}

							// 创建InterfaceBody结构体实例
							messageSSE = structs.InterfaceBody{
								Content:        parts[2],       // 假设空格字符串是期望的内容
								State:          20,             // 假设的状态码
								PromptKeyboard: promptkeyboard, // 使用更新后的promptkeyboard
							}

							// 发送SSE私人消息
							utils.SendPrivateMessageSSE(message.UserID, messageSSE)
						}
					} else {
						utils.SendGroupMessage(message.GroupID, saveresponse)
					}
				}
				return
			}
		}

		// 请求conversation api 增加当前用户上下文
		conversationID, parentMessageID, err := app.handleUserContext(message.UserID)
		//每句话清空上一句话的messageBuilder
		messageBuilder.Reset()
		fmtf.Printf("conversationID: %s,parentMessageID%s\n", conversationID, parentMessageID)
		if err != nil {
			fmtf.Printf("Error handling user context: %v\n", err)
			return
		}
		// 构建并发送请求到conversation接口
		port := config.GetPort()
		portStr := fmtf.Sprintf(":%d", port)
		url := "http://127.0.0.1" + portStr + "/conversation"

		requestBody, err := json.Marshal(map[string]interface{}{
			"message":         newmsg,
			"conversationId":  conversationID,
			"parentMessageId": parentMessageID,
			"user_id":         message.UserID,
		})
		if err != nil {
			fmtf.Printf("Error marshalling request: %v\n", err)
			return
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			fmtf.Printf("Error sending request to conversation interface: %v\n", err)
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
					fmtf.Printf("Error reading SSE response: %v\n", err)
					return
				}

				// 忽略空行
				if string(line) == "\n" {
					continue
				}

				// 处理接收到的数据
				fmtf.Printf("Received SSE data: %s", string(line))

				// 去除"data: "前缀后进行JSON解析
				jsonData := strings.TrimPrefix(string(line), "data: ")
				var responseData map[string]interface{}
				if err := json.Unmarshal([]byte(jsonData), &responseData); err == nil {
					//接收到最后一条信息
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
									fmtf.Printf("A完整信息: %s,已发送信息:%s 新部分:%s\n", response, accumulatedMessage, newPart)
									//这里记录完整的信息
									//RecordStringByNewmsg(newmsg, response)
									// 判断消息类型，如果是私人消息或私有群消息，发送私人消息；否则，根据配置决定是否发送群消息
									if message.RealMessageType == "group_private" || message.MessageType == "private" {
										if !config.GetUsePrivateSSE() {
											utils.SendPrivateMessage(message.UserID, newPart)
										} else {
											//最后一条了
											messageSSE := structs.InterfaceBody{
												Content: newPart,
												State:   11,
											}
											utils.SendPrivateMessageSSE(message.UserID, messageSSE)
										}
									} else {
										utils.SendGroupMessage(message.GroupID, newPart)
									}
								}

							} else if response != "" {
								// 如果accumulatedMessage不存在或不是子串，print
								fmtf.Printf("B完整信息: %s,已发送信息:%s", response, accumulatedMessage)
								if accumulatedMessage == "" {
									// 判断消息类型，如果是私人消息或私有群消息，发送私人消息；否则，根据配置决定是否发送群消息
									if message.RealMessageType == "group_private" || message.MessageType == "private" {
										utils.SendPrivateMessage(message.UserID, response)
									} else {
										utils.SendGroupMessage(message.GroupID, response)
									}
								}
							}

							// 清空映射中对应的累积消息
							groupUserMessages[key] = ""
						}
					} else {
						//发送信息
						fmtf.Printf("发信息: %s", string(line))
						splitAndSendMessages(message, string(line), newmsg)
					}
				}

			}

			// 在SSE流结束后更新用户上下文 在这里调用gensokyo流式接口的最后一步 插推荐气泡
			if lastMessageID != "" {
				fmtf.Printf("lastMessageID: %s\n", lastMessageID)
				err := app.updateUserContext(message.UserID, lastMessageID)
				if err != nil {
					fmtf.Printf("Error updating user context: %v\n", err)
				}
				if message.RealMessageType == "group_private" || message.MessageType == "private" {
					if config.GetUsePrivateSSE() {
						//发气泡和按钮
						promptkeyboard := config.GetPromptkeyboard()
						//最后一条了
						messageSSE := structs.InterfaceBody{
							Content:        " ",
							State:          20,
							PromptKeyboard: promptkeyboard,
						}
						utils.SendPrivateMessageSSE(message.UserID, messageSSE)
						ResetIndex(newmsg)
					}
				}

			}
		} else {
			// 处理常规响应
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				fmtf.Printf("Error reading response body: %v\n", err)
				return
			}
			fmtf.Printf("Response from conversation interface: %s\n", string(responseBody))

			// 使用map解析响应数据以获取response字段和messageId
			var responseData map[string]interface{}
			if err := json.Unmarshal(responseBody, &responseData); err != nil {
				fmtf.Printf("Error unmarshalling response data: %v\n", err)
				return
			}

			// 使用提取的response内容发送消息
			if response, ok := responseData["response"].(string); ok && response != "" {
				// 判断消息类型，如果是私人消息或私有群消息，发送私人消息；否则，根据配置决定是否发送群消息
				if message.RealMessageType == "group_private" || message.MessageType == "private" {
					utils.SendPrivateMessage(message.UserID, response)
				} else {
					utils.SendGroupMessage(message.GroupID, response)
				}
			}

			// 更新用户上下文
			if messageId, ok := responseData["messageId"].(string); ok {
				err := app.updateUserContext(message.UserID, messageId)
				if err != nil {
					fmtf.Printf("Error updating user context: %v\n", err)
				}
			}
		}

		// 发送响应
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Request received and processed"))

	case map[string]interface{}:
		// message.Message是一个map[string]interface{}
		fmtf.Println("Received map message, handling not implemented yet")
		// 处理map类型消息的逻辑（TODO）

	default:
		// message.Message是一个未知类型
		fmtf.Printf("Received message of unexpected type: %T\n", msg)
		return
	}

}

func splitAndSendMessages(message structs.OnebotGroupMessage, line string, newmesssage string) {
	// 提取JSON部分
	dataPrefix := "data: "
	jsonStr := strings.TrimPrefix(line, dataPrefix)

	// 解析JSON数据
	var sseData struct {
		Response string `json:"response"`
	}
	err := json.Unmarshal([]byte(jsonStr), &sseData)
	if err != nil {
		fmtf.Printf("Error unmarshalling SSE data: %v\n", err)
		return
	}

	// 处理提取出的信息
	processMessage(sseData.Response, message, newmesssage)
}

func processMessage(response string, msg structs.OnebotGroupMessage, newmesssage string) {
	key := utils.GetKey(msg.GroupID, msg.UserID)

	// 定义中文全角和英文标点符号
	punctuations := []rune{'。', '！', '？', '，', ',', '.', '!', '?'}

	for _, char := range response {
		messageBuilder.WriteRune(char)
		if utils.ContainsRune(punctuations, char) {
			// 达到标点符号，发送累积的整个消息
			if messageBuilder.Len() > 0 {
				accumulatedMessage := messageBuilder.String()
				groupUserMessages[key] += accumulatedMessage

				// 判断消息类型，如果是私人消息或私有群消息，发送私人消息；否则，根据配置决定是否发送群消息
				if msg.RealMessageType == "group_private" || msg.MessageType == "private" {
					if !config.GetUsePrivateSSE() {
						utils.SendPrivateMessage(msg.UserID, accumulatedMessage)
					} else {
						if IncrementIndex(newmesssage) == 1 {
							//第一条信息
							//取出当前信息作为按钮回调
							//CallbackData := GetStringById(lastMessageID)
							uerid := strconv.FormatInt(msg.UserID, 10)
							messageSSE := structs.InterfaceBody{
								Content:      accumulatedMessage,
								State:        1,
								ActionButton: 10,
								CallbackData: uerid,
							}
							utils.SendPrivateMessageSSE(msg.UserID, messageSSE)
						} else {
							//SSE的前半部分
							messageSSE := structs.InterfaceBody{
								Content: accumulatedMessage,
								State:   1,
							}
							utils.SendPrivateMessageSSE(msg.UserID, messageSSE)
						}
					}
				} else {
					utils.SendGroupMessage(msg.GroupID, accumulatedMessage)
				}

				messageBuilder.Reset() // 重置消息构建器
			}
		}
	}
}
