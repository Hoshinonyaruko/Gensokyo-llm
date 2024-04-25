package applogic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
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

	// 从数据库读取用户的剧情存档
	CustomRecord, err := app.FetchCustomRecord(message.UserID)
	if err != nil {
		fmt.Printf("app.FetchCustomRecord 出错: %s\n", err)
	}

	var promptstr string
	if CustomRecord != nil {
		// 提示词参数
		if CustomRecord.PromptStr == "" {
			// 读取URL参数 "prompt"
			promptstr = r.URL.Query().Get("prompt")
			if promptstr != "" {
				// 使用 prompt 变量进行后续处理
				fmt.Printf("收到prompt参数: %s\n", promptstr)
			}
		} else {
			promptstr = CustomRecord.PromptStr
			fmt.Printf("刷新prompt参数: %s,newPromptStrStat:%d\n", promptstr, CustomRecord.PromptStrStat-1)
			newPromptStrStat := CustomRecord.PromptStrStat - 1
			err = app.InsertCustomTableRecord(message.UserID, promptstr, newPromptStrStat)
			if err != nil {
				fmt.Printf("app.InsertCustomTableRecord 出错: %s\n", err)
			}
		}

		// 提示词之间流转 达到信号量
		markType := config.GetPromptMarkType(promptstr)
		if (markType == 0 || markType == 1) && (CustomRecord.PromptStrStat-1 <= 0) {
			PromptMarks := config.GetPromptMarks(promptstr)
			if len(PromptMarks) != 0 {
				randomIndex := rand.Intn(len(PromptMarks))
				newPromptStr := PromptMarks[randomIndex]

				// 如果 markType 是 1，提取 "aaa" 部分
				if markType == 1 {
					parts := strings.Split(newPromptStr, ":")
					if len(parts) > 0 {
						newPromptStr = parts[0] // 取冒号前的部分作为新的提示词
					}
				}

				// 刷新新的提示词给用户目前的状态
				// 获取新的信号长度
				PromptMarksLength := config.GetPromptMarksLength(newPromptStr)

				app.InsertCustomTableRecord(message.UserID, newPromptStr, PromptMarksLength)
				fmt.Printf("流转prompt参数: %s,newPromptStrStat:%d\n", newPromptStr, PromptMarksLength)
			}
		}
	} else {
		// 读取URL参数 "prompt"
		promptstr = r.URL.Query().Get("prompt")
		if promptstr != "" {
			// 使用 prompt 变量进行后续处理
			fmt.Printf("收到prompt参数: %s\n", promptstr)
		}
		PromptMarksLength := config.GetPromptMarksLength(promptstr)
		err = app.InsertCustomTableRecord(message.UserID, promptstr, PromptMarksLength)
		if err != nil {
			fmt.Printf("app.InsertCustomTableRecord 出错: %s\n", err)
		}
	}

	// 读取URL参数 "selfid"
	selfid := r.URL.Query().Get("selfid")
	if selfid != "" {
		// 使用 prompt 变量进行后续处理
		fmt.Printf("收到selfid参数: %s\n", selfid)
	}

	// 读取URL参数 "api"
	api := r.URL.Query().Get("api")
	if selfid != "" {
		// 使用 prompt 变量进行后续处理
		fmt.Printf("收到api参数: %s\n", selfid)
	}

	// 打印日志信息，包括prompt参数
	fmtf.Printf("收到onebotv11信息: %+v\n", string(body))

	// 打印消息和其他相关信息
	fmtf.Printf("Received message: %v\n", message.Message)
	fmtf.Printf("Full message details: %+v\n", message)

	// 判断message.Message的类型
	switch msg := message.Message.(type) {
	case string:
		// message.Message是一个string
		fmtf.Printf("userid:[%v]Received string message: %s\n", message.UserID, msg)

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

		if utils.BlacklistIntercept(message, selfid) {
			fmtf.Printf("userid:[%v]这位用户在黑名单中,被拦截", message.UserID)
			return
		}

		//处理重置指令
		if isResetCommand {
			fmtf.Println("处理重置操作")
			app.migrateUserToNewContext(message.UserID)
			RestoreResponse := config.GetRandomRestoreResponses()
			if message.RealMessageType == "group_private" || message.MessageType == "private" {
				if !config.GetUsePrivateSSE() {
					utils.SendPrivateMessage(message.UserID, RestoreResponse, selfid)
				} else {
					utils.SendSSEPrivateRestoreMessage(message.UserID, RestoreResponse)
				}
			} else {
				utils.SendGroupMessage(message.GroupID, message.UserID, RestoreResponse, selfid)
			}
			// 处理故事情节的重置
			app.deleteCustomRecord(message.UserID)
			return
		}

		withdrawCommand := config.GetWithdrawCommand()

		// 检查checkResetCommand是否在WithdrawCommand列表中
		iswithdrawCommand := false
		for _, command := range withdrawCommand {
			if checkResetCommand == command {
				iswithdrawCommand = true
				break
			}
		}

		// 处理撤回信息
		if iswithdrawCommand {
			handleWithdrawMessage(message)
			return
		}

		// newmsg 是一个用于缓存和安全判断的临时量
		newmsg := message.Message.(string)
		// 去除注入的提示词
		if config.GetIgnoreExtraTips() {
			newmsg = utils.RemoveBracketsContent(newmsg)
		}

		var (
			vector               []float64
			lastSelectedVectorID int // 用于存储最后选取的相似文本的ID
		)

		// 进行字数拦截
		if config.GetQuestionMaxLenth() != 0 {
			if utils.LengthIntercept(newmsg, message, selfid) {
				fmtf.Printf("字数过长,可在questionMaxLenth配置项修改,Q: %v", newmsg)
				// 发送响应
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("question too long"))
				return
			}
		}

		// 进行语言判断拦截
		if len(config.GetAllowedLanguages()) > 0 {
			if utils.LanguageIntercept(newmsg, message, selfid) {
				fmtf.Printf("不安全!不支持的语言,可在config.yml设置允许的语言,allowedLanguages配置项,Q: %v", newmsg)
				// 发送响应
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("language not support"))
				return
			}
		}

		// 如果使用向量缓存 或者使用 向量安全词
		if config.GetUseCache(promptstr) || config.GetVectorSensitiveFilter() {
			if config.GetPrintHanming() {
				fmtf.Printf("计算向量的文本: %v", newmsg)
			}
			// 计算文本向量
			vector, err = app.CalculateTextEmbedding(newmsg)
			if err != nil {
				fmtf.Printf("Error calculating text embedding: %v", err)
				// 发送响应
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Error calculating text embedding"))
				return
			}
		}

		// 向量安全词部分,机器人向量安全屏障
		if config.GetVectorSensitiveFilter() {
			ret, retstr, err := app.InterceptSensitiveContent(vector, message, selfid)
			if err != nil {
				fmtf.Printf("Error in InterceptSensitiveContent: %v", err)
				// 发送响应
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Error in InterceptSensitiveContent"))
				return
			}
			if ret != 0 {
				fmtf.Printf("sensitive content detected!%v\n", message)
				// 发送响应
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("sensitive content detected![" + retstr + "]"))
				return
			}
		}

		// 缓存省钱部分
		if config.GetUseCache(promptstr) {
			//fmtf.Printf("计算向量: %v", vector)
			cacheThreshold := config.GetCacheThreshold()
			// 搜索相似文本和对应的ID
			similarTexts, ids, err := app.searchForSingleVector(vector, cacheThreshold)
			if err != nil {
				fmtf.Printf("Error searching for similar texts: %v", err)
				// 发送响应
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Error searching for similar texts"))
				return
			}

			if len(similarTexts) > 0 {
				// 总是获取最相似的文本的ID，不管是否最终使用
				lastSelectedVectorID = ids[0]

				chance := rand.Intn(100)
				// 检查是否满足设定的概率
				if chance < config.GetCacheChance() {
					// 使用最相似的文本的答案
					fmtf.Printf("读取表:%v\n", similarTexts[0])
					responseText, err := app.GetRandomAnswer(similarTexts[0])
					if err == nil {
						fmtf.Printf("缓存命中,Q:%v,A:%v\n", newmsg, responseText)
						//加入上下文
						if app.AddSingleContext(message, responseText) {
							fmtf.Printf("缓存加入上下文成功")
						}
						// 发送响应消息
						if message.RealMessageType == "group_private" || message.MessageType == "private" {
							if !config.GetUsePrivateSSE() {
								utils.SendPrivateMessage(message.UserID, responseText, selfid)
							} else {
								utils.SendSSEPrivateMessage(message.UserID, responseText)
							}
						} else {
							utils.SendGroupMessage(message.GroupID, message.UserID, responseText, selfid)
						}
						// 发送响应
						w.WriteHeader(http.StatusOK)
						w.Write([]byte("Request received and use cache"))
						return // 成功使用缓存答案，提前退出
					} else {
						fmtf.Printf("Error getting random answer: %v", err)

					}
				} else {
					fmtf.Printf("缓存命中，但没有符合概率，继续执行后续代码\n")
					// 注意：这里不需要再生成 lastSelectedVectorID，因为上面已经生成
				}
			} else {
				// 没有找到相似文本，存储新的文本及其向量
				newVectorID, err := app.insertVectorData(newmsg, vector)
				if err != nil {
					fmtf.Printf("Error inserting new vector data: %v", err)
					// 发送响应
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Error inserting new vector data"))
					return
				}
				lastSelectedVectorID = int(newVectorID) // 存储新插入向量的ID
				fmtf.Printf("没找到缓存,准备储存了lastSelectedVectorID: %v\n", lastSelectedVectorID)
			}

			// 这里继续执行您的逻辑，比如生成新的答案等
			// 注意：根据实际情况调整后续逻辑
		}

		//提示词安全部分
		if config.GetAntiPromptAttackPath() != "" {
			if checkResponseThreshold(newmsg) {
				fmtf.Printf("提示词不安全,过滤:%v", message)
				saveresponse := config.GetRandomSaveResponse()
				if saveresponse != "" {
					if message.RealMessageType == "group_private" || message.MessageType == "private" {
						if !config.GetUsePrivateSSE() {
							utils.SendPrivateMessage(message.UserID, saveresponse, selfid)
						} else {
							utils.SendSSEPrivateSafeMessage(message.UserID, saveresponse)
						}
					} else {
						utils.SendGroupMessage(message.GroupID, message.UserID, saveresponse, selfid)
					}
				}
				// 发送响应
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Request received and not safe"))
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
			// 发送响应
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Error handling user context"))
			return
		}

		// 构建并发送请求到conversation接口
		port := config.GetPort()
		portStr := fmt.Sprintf(":%d", port)

		// 初始化URL，根据api参数动态调整路径
		basePath := "/conversation"
		if api != "" {
			fmtf.Printf("收到api参数: %s\n", api)
			basePath = "/" + api // 动态替换conversation部分为api参数值
		}
		var baseURL string
		if config.GetLotus(promptstr) == "" {
			baseURL = "http://127.0.0.1" + portStr + basePath
		} else {
			baseURL = config.GetLotus(promptstr) + basePath
		}

		// 使用net/url包来构建和编码URL
		urlParams := url.Values{}
		if promptstr != "" {
			urlParams.Add("prompt", promptstr)
		}

		// 将查询参数编码后附加到基本URL上
		fullURL := baseURL
		if len(urlParams) > 0 {
			fullURL += "?" + urlParams.Encode()
		}

		fmtf.Printf("Generated URL:%v\n", fullURL)

		// 请求模型还是使用原文请求
		requestmsg := message.Message.(string)

		if config.GetPrintHanming() {
			fmtf.Printf("消息进入替换前:%v", requestmsg)
		}

		// 繁体转换简体 安全策略
		requestmsg, err = utils.ConvertTraditionalToSimplified(requestmsg)
		if err != nil {
			fmtf.Printf("繁体转换简体失败:%v", err)
		}

		// 替换in替换词规则
		if config.GetSensitiveMode() {
			requestmsg = acnode.CheckWordIN(requestmsg)
		}

		fmtf.Printf("实际请求conversation端点内容:[%v]%v\n", message.UserID, requestmsg)

		requestBody, err := json.Marshal(map[string]interface{}{
			"message":         requestmsg,
			"conversationId":  conversationID,
			"parentMessageId": parentMessageID,
			"user_id":         message.UserID,
		})

		if err != nil {
			fmtf.Printf("Error marshalling request: %v\n", err)
			return
		}

		resp, err := http.Post(fullURL, "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			fmtf.Printf("Error sending request to conversation interface: %v\n", err)
			return
		}

		defer resp.Body.Close()

		var lastMessageID string
		var response string

		if config.GetuseSse(promptstr) {
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
				if !config.GetHideExtraLogs() {
					fmtf.Printf("Received SSE data: %s", string(line))
				}

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
						if response, ok = responseData["response"].(string); ok {
							// 如果accumulatedMessage是response的子串，则提取新的部分并发送
							if exists && strings.HasPrefix(response, accumulatedMessage) {
								newPart := response[len(accumulatedMessage):]
								if newPart != "" {
									fmtf.Printf("A完整信息: %s,已发送信息:%s 新部分:%s\n", response, accumulatedMessage, newPart)
									// 判断消息类型，如果是私人消息或私有群消息，发送私人消息；否则，根据配置决定是否发送群消息
									if message.RealMessageType == "group_private" || message.MessageType == "private" {
										if !config.GetUsePrivateSSE() {
											utils.SendPrivateMessage(message.UserID, newPart, selfid)
										} else {
											//最后一条了
											messageSSE := structs.InterfaceBody{
												Content: newPart,
												State:   11,
											}
											utils.SendPrivateMessageSSE(message.UserID, messageSSE)
										}
									} else {
										utils.SendGroupMessage(message.GroupID, message.UserID, newPart, selfid)
									}
								} else {
									//流的最后一次是完整结束的
									fmtf.Printf("A完整信息: %s(sse完整结束)\n", response)
								}

							} else if response != "" {
								// 如果accumulatedMessage不存在或不是子串，print
								fmtf.Printf("B完整信息: %s,已发送信息:%s", response, accumulatedMessage)
								if accumulatedMessage == "" {
									// 判断消息类型，如果是私人消息或私有群消息，发送私人消息；否则，根据配置决定是否发送群消息
									if message.RealMessageType == "group_private" || message.MessageType == "private" {
										if !config.GetUsePrivateSSE() {
											utils.SendPrivateMessage(message.UserID, response, selfid)
										} else {
											//最后一条了
											messageSSE := structs.InterfaceBody{
												Content: response,
												State:   11,
											}
											utils.SendPrivateMessageSSE(message.UserID, messageSSE)
										}
									} else {
										utils.SendGroupMessage(message.GroupID, message.UserID, response, selfid)
									}
								}
							}
							// 处理故事模式
							app.ProcessAnswer(message.UserID, response, promptstr)
							// 清空之前加入缓存
							// 缓存省钱部分 这里默认不被覆盖,如果主配置开了缓存,始终缓存.
							if config.GetUseCache() {
								if response != "" {
									fmtf.Printf("缓存了Q:%v,A:%v,向量ID:%v", newmsg, response, lastSelectedVectorID)
									app.InsertQAEntry(newmsg, response, lastSelectedVectorID)
								} else {
									fmtf.Printf("缓存Q:%v时遇到问题,A为空,检查api是否存在问题", newmsg)
								}

							}

							// 清空映射中对应的累积消息
							groupUserMessages[key] = ""
						}
					} else {
						//发送信息
						if !config.GetHideExtraLogs() {
							fmtf.Printf("收到流数据,切割并发送信息: %s", string(line))
						}
						splitAndSendMessages(message, string(line), newmsg, selfid)
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
						var promptkeyboard []string
						if !config.GetUseAIPromptkeyboard() {
							promptkeyboard = config.GetPromptkeyboard()
						} else {
							fmtf.Printf("ai生成气泡:%v", "Q"+newmsg+"A"+response)
							promptkeyboard = GetPromptKeyboardAI("Q"+newmsg+"A"+response, promptstr)
						}

						// 使用acnode.CheckWordOUT()过滤promptkeyboard中的每个字符串
						for i, item := range promptkeyboard {
							promptkeyboard[i] = acnode.CheckWordOUT(item)
						}

						//最后一条了
						messageSSE := structs.InterfaceBody{
							Content:        " ",
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
			var ok bool
			// 使用提取的response内容发送消息
			if response, ok = responseData["response"].(string); ok && response != "" {
				// 判断消息类型，如果是私人消息或私有群消息，发送私人消息；否则，根据配置决定是否发送群消息
				if message.RealMessageType == "group_private" || message.MessageType == "private" {
					utils.SendPrivateMessage(message.UserID, response, selfid)
				} else {
					utils.SendGroupMessage(message.GroupID, message.UserID, response, selfid)
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

		// OUT规则不仅对实际发送api生效,也对http结果生效
		if config.GetSensitiveModeType() == 1 {
			response = acnode.CheckWordOUT(response)
		}

		// 发送响应
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Request received and processed Q:" + newmsg + " A:" + response))

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

func splitAndSendMessages(message structs.OnebotGroupMessage, line string, newmesssage string, selfid string) {
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

	if sseData.Response != "\n\n" {
		// 处理提取出的信息
		processMessage(sseData.Response, message, newmesssage, selfid)
	} else {
		fmtf.Printf("忽略llm末尾的换行符")
	}
}

func processMessage(response string, msg structs.OnebotGroupMessage, newmesssage string, selfid string) {
	key := utils.GetKey(msg.GroupID, msg.UserID)

	// 定义中文全角和英文标点符号
	punctuations := []rune{'。', '！', '？', '，', ',', '.', '!', '?', '~'}

	for _, char := range response {
		messageBuilder.WriteRune(char)
		if utils.ContainsRune(punctuations, char, msg.GroupID) {
			// 达到标点符号，发送累积的整个消息
			if messageBuilder.Len() > 0 {
				accumulatedMessage := messageBuilder.String()
				groupUserMessages[key] += accumulatedMessage

				// 判断消息类型，如果是私人消息或私有群消息，发送私人消息；否则，根据配置决定是否发送群消息
				if msg.RealMessageType == "group_private" || msg.MessageType == "private" {
					if !config.GetUsePrivateSSE() {
						utils.SendPrivateMessage(msg.UserID, accumulatedMessage, selfid)
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
					utils.SendGroupMessage(msg.GroupID, msg.UserID, accumulatedMessage, selfid)
				}

				messageBuilder.Reset() // 重置消息构建器
			}
		}
	}
}

// 处理撤回信息的函数
func handleWithdrawMessage(message structs.OnebotGroupMessage) {
	fmt.Println("处理撤回操作")
	var id int64

	// 根据消息类型决定使用哪个ID
	switch message.RealMessageType {
	case "group_private", "guild_private":
		id = message.UserID
	case "group", "guild":
		id = message.GroupID
	default:
		fmt.Println("Unsupported message type for withdrawal:", message.RealMessageType)
		return
	}

	// 调用DeleteLatestMessage函数
	err := utils.DeleteLatestMessage(message.RealMessageType, id, message.UserID)
	if err != nil {
		fmt.Println("Error deleting latest message:", err)
		return
	}
}
