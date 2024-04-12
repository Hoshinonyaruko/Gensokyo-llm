package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/abadojack/whatlanggo"
	"github.com/google/uuid"
	"github.com/hoshinonyaruko/gensokyo-llm/acnode"
	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/hunyuan"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
)

// ResponseData 是用于解析HTTP响应的结构体
type ResponseData struct {
	Data struct {
		MessageID int64 `json:"message_id"`
	} `json:"data"`
}

// MessageIDInfo 代表消息ID及其到期时间
type MessageIDInfo struct {
	MessageID int64     // 消息ID
	Expires   time.Time // 到期时间
}

// UserIDMessageIDs 存储每个用户ID对应的消息ID数组及其有效期
var UserIDMessageIDs = make(map[int64][]MessageIDInfo)
var muUserIDMessageIDs sync.RWMutex // 用于UserIDMessageIDs的读写锁

func GenerateUUID() string {
	return uuid.New().String()
}

func PrintChatProRequest(request *hunyuan.ChatProRequest) {

	// 打印Messages
	for i, msg := range request.Messages {
		fmtf.Printf("Message %d:\n", i)
		fmtf.Printf("Content: %s\n", *msg.Content)
		fmtf.Printf("Role: %s\n", *msg.Role)
	}

}

func PrintChatStdRequest(request *hunyuan.ChatStdRequest) {

	// 打印Messages
	for i, msg := range request.Messages {
		fmtf.Printf("Message %d:\n", i)
		fmtf.Printf("Content: %s\n", *msg.Content)
		fmtf.Printf("Role: %s\n", *msg.Role)
	}

}

// contains 检查一个字符串切片是否包含一个特定的字符串
func Contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// 获取复合键
func GetKey(groupid int64, userid int64) string {
	return fmt.Sprintf("%d.%d", groupid, userid)
}

// 随机的分布发送
func ContainsRune(slice []rune, value rune, groupid int64) bool {
	var probability int
	if groupid == 0 {
		// 获取概率百分比
		probability = config.GetSplitByPuntuations()
	} else {
		// 获取概率百分比
		probability = config.GetSplitByPuntuationsGroup()
	}

	for _, item := range slice {
		if item == value {
			// 将概率转换为0到1之间的浮点数
			probabilityPercentage := float64(probability) / 100.0
			// 生成一个0到1之间的随机浮点数
			randomValue := rand.Float64()
			// 如果随机数小于或等于概率，则返回true
			return randomValue <= probabilityPercentage
		}
	}
	return false
}

// 取出ai回答
func ExtractEventDetails(eventData map[string]interface{}) (string, structs.UsageInfo) {
	var responseTextBuilder strings.Builder
	var totalUsage structs.UsageInfo

	// 提取使用信息
	if usage, ok := eventData["Usage"].(map[string]interface{}); ok {
		var usageInfo structs.UsageInfo
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

func SendGroupMessage(groupID int64, userID int64, message string) error {
	// 获取基础URL
	baseURL := config.GetHttpPath() // 假设config.getHttpPath()返回基础URL

	// 构建完整的URL
	url := baseURL + "/send_group_msg"

	if config.GetSensitiveModeType() == 1 {
		message = acnode.CheckWordOUT(message)
	}

	// 构造请求体
	requestBody, err := json.Marshal(map[string]interface{}{
		"group_id": groupID,
		"user_id":  userID,
		"message":  message,
	})
	fmtf.Printf("发群信息请求:%v", string(requestBody))
	if err != nil {
		return fmtf.Errorf("failed to marshal request body: %w", err)
	}

	// 发送POST请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmtf.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmtf.Errorf("received non-OK response status: %s", resp.Status)
	}

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应体以获取message_id
	var responseData ResponseData
	if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
		return fmt.Errorf("failed to unmarshal response data: %w", err)
	}
	messageID := responseData.Data.MessageID

	// 添加messageID到全局变量
	AddMessageID(userID, messageID)

	// 输出响应体，这一步是可选的
	fmt.Println("Response Body:", string(bodyBytes))

	return nil
}

func SendPrivateMessage(UserID int64, message string) error {
	// 获取基础URL
	baseURL := config.GetHttpPath() // 假设config.getHttpPath()返回基础URL

	// 构建完整的URL
	url := baseURL + "/send_private_msg"

	if config.GetSensitiveModeType() == 1 {
		message = acnode.CheckWordOUT(message)
	}

	// 构造请求体
	requestBody, err := json.Marshal(map[string]interface{}{
		"user_id": UserID,
		"message": message,
	})

	if err != nil {
		return fmtf.Errorf("failed to marshal request body: %w", err)
	}

	// 发送POST请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmtf.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmtf.Errorf("received non-OK response status: %s", resp.Status)
	}

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应体以获取message_id
	var responseData ResponseData
	if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
		return fmt.Errorf("failed to unmarshal response data: %w", err)
	}
	messageID := responseData.Data.MessageID

	// 添加messageID到全局变量
	AddMessageID(UserID, messageID)

	// 输出响应体，这一步是可选的
	fmt.Println("Response Body:", string(bodyBytes))

	return nil
}

func SendPrivateMessageSSE(UserID int64, message structs.InterfaceBody) error {
	// 获取基础URL
	baseURL := config.GetHttpPath() // 假设config.GetHttpPath()返回基础URL

	// 构建完整的URL
	url := baseURL + "/send_private_msg_sse"
	// 调试用的
	if config.GetPrintHanming() {
		fmtf.Printf("流式信息替换前:%v", message.Content)
	}

	// 检查是否需要启用敏感词过滤
	if config.GetSensitiveModeType() == 1 && message.Content != "" {
		message.Content = acnode.CheckWordOUT(message.Content)
	}

	// 调试用的
	if config.GetPrintHanming() {
		fmtf.Printf("流式信息替换后:%v", message.Content)
	}

	// 构造请求体，包括InterfaceBody
	requestBody, err := json.Marshal(map[string]interface{}{
		"user_id": UserID,
		"message": message,
	})
	if err != nil {
		return fmtf.Errorf("failed to marshal request body: %w", err)
	}

	// 发送POST请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmtf.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmtf.Errorf("received non-OK response status: %s", resp.Status)
	}

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应体以获取message_id
	var responseData ResponseData
	if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
		return fmt.Errorf("failed to unmarshal response data: %w", err)
	}
	messageID := responseData.Data.MessageID

	// 添加messageID到全局变量
	AddMessageID(UserID, messageID)

	// 输出响应体，这一步是可选的
	fmt.Println("Response Body:", string(bodyBytes))

	return nil
}

// ReverseString 颠倒给定字符串中的字符顺序
func ReverseString(s string) string {
	// // 将字符串转换为rune切片，以便处理多字节字符
	// runes := []rune(s)
	// for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
	// 	// 交换前后对应的字符
	// 	runes[i], runes[j] = runes[j], runes[i]
	// }
	// // 将颠倒后的rune切片转换回字符串
	// return string(runes)
	return "####" + s + "####"
}

// RemoveBracketsContent 接收一个字符串，并移除所有[[...]]的内容
func RemoveBracketsContent(input string) string {
	// 编译一个正则表达式，用于匹配[[任意字符]]的模式
	re := regexp.MustCompile(`\[\[.*?\]\]`)
	// 使用正则表达式的ReplaceAllString方法删除匹配的部分
	return re.ReplaceAllString(input, "")
}

func PostSensitiveMessages() error {
	port := config.GetPort()                                     // 从config包获取端口号
	portStr := fmt.Sprintf("http://127.0.0.1:%d/gensokyo", port) // 根据端口号构建URL

	file, err := os.Open("test.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var results []string
	for scanner.Scan() {
		text := scanner.Text()

		// 使用读取的文本填充message和raw_message字段
		data := structs.OnebotGroupMessage{
			Font:            0,
			Message:         text,
			MessageID:       0,
			MessageSeq:      0,
			MessageType:     "private",
			PostType:        "message",
			RawMessage:      text,
			RealMessageType: "group_private",
			SelfID:          100000000,
			Sender: structs.Sender{
				Nickname: "测试脚本",
				UserID:   100000000,
			},
			SubType: "friend",
			Time:    1700000000,
			UserID:  100000000,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}

		response, err := http.Post(portStr, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		defer response.Body.Close()

		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		fmtf.Printf("测试脚本运行中:%v", results)
		results = append(results, string(responseBody))
	}

	// 使用当前时间戳生成文件名
	currentTime := time.Now()
	fileName := "test_result_" + currentTime.Format("20060102_150405") + ".txt"

	// 将HTTP响应结果保存到指定的文件中
	return os.WriteFile(fileName, []byte(strings.Join(results, "\n")), 0644)
}

// SendSSEPrivateMessage 分割并发送消息的核心逻辑，直接遍历字符串
func SendSSEPrivateMessage(userID int64, content string) {
	punctuations := []rune{'。', '！', '？', '，', ',', '.', '!', '?'}
	splitProbability := config.GetSplitByPuntuations()

	var parts []string
	var currentPart strings.Builder

	for _, runeValue := range content {
		currentPart.WriteRune(runeValue)
		if strings.ContainsRune(string(punctuations), runeValue) {
			// 根据概率决定是否分割
			if rand.Intn(100) < splitProbability {
				parts = append(parts, currentPart.String())
				currentPart.Reset()
			}
		}
	}
	// 添加最后一部分（如果有的话）
	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	// 根据parts长度处理状态
	for i, part := range parts {
		state := 1
		if i == len(parts)-2 { // 倒数第二部分
			state = 11
		} else if i == len(parts)-1 { // 最后一部分
			state = 20
		}

		// 构造消息体并发送
		messageSSE := structs.InterfaceBody{
			Content: part,
			State:   state,
		}

		if state == 20 { // 对最后一部分特殊处理
			RestoreResponses := config.GetRestoreCommand()
			promptKeyboard := config.GetPromptkeyboard()

			if len(RestoreResponses) > 0 {
				selectedRestoreResponse := RestoreResponses[rand.Intn(len(RestoreResponses))]
				if len(promptKeyboard) > 0 {
					promptKeyboard[0] = selectedRestoreResponse
				}
			}

			messageSSE.PromptKeyboard = promptKeyboard
		}

		// 发送SSE消息函数
		SendPrivateMessageSSE(userID, messageSSE)
	}
}

// SendSSEPrivateSafeMessage 分割并发送安全消息的核心逻辑，直接遍历字符串
func SendSSEPrivateSafeMessage(userID int64, saveresponse string) {
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

	SendPrivateMessageSSE(userID, messageSSE)

	// 中间
	messageSSE = structs.InterfaceBody{
		Content: parts[1],
		State:   11,
	}
	SendPrivateMessageSSE(userID, messageSSE)

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
	SendPrivateMessageSSE(userID, messageSSE)
}

// SendSSEPrivateRestoreMessage 分割并发送重置消息的核心逻辑，直接遍历字符串
func SendSSEPrivateRestoreMessage(userID int64, RestoreResponse string) {
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

	SendPrivateMessageSSE(userID, messageSSE)

	//中间
	messageSSE = structs.InterfaceBody{
		Content: parts[1],
		State:   11,
	}
	SendPrivateMessageSSE(userID, messageSSE)

	// 从配置中获取promptkeyboard
	promptkeyboard := config.GetPromptkeyboard()

	// 创建InterfaceBody结构体实例
	messageSSE = structs.InterfaceBody{
		Content:        parts[2],       // 假设空格字符串是期望的内容
		State:          20,             // 假设的状态码
		PromptKeyboard: promptkeyboard, // 使用更新后的promptkeyboard
	}

	// 发送SSE私人消息
	SendPrivateMessageSSE(userID, messageSSE)
}

// LanguageIntercept 检查文本语言，如果不在允许列表中，则返回 true 并发送消息
func LanguageIntercept(text string, message structs.OnebotGroupMessage) bool {
	info := whatlanggo.Detect(text)
	lang := whatlanggo.LangToString(info.Lang)
	fmtf.Printf("LanguageIntercept:%v\n", lang)

	allowedLanguages := config.GetAllowedLanguages()
	for _, allowed := range allowedLanguages {
		if strings.Contains(allowed, lang) {
			return false // 语言允许，不拦截
		}
	}

	// 语言不允许，进行拦截
	responseMessage := config.GetLanguagesResponseMessages()
	friendlyName := FriendlyLanguageNameCN(info.Lang)
	responseMessage = strings.Replace(responseMessage, "**", friendlyName, -1)

	// 发送响应消息
	if message.RealMessageType == "group_private" || message.MessageType == "private" {
		if !config.GetUsePrivateSSE() {
			SendPrivateMessage(message.UserID, responseMessage)
		} else {
			SendSSEPrivateMessage(message.UserID, responseMessage)
		}
	} else {
		SendGroupMessage(message.GroupID, message.UserID, responseMessage)
	}

	return true // 拦截
}

// FriendlyLanguageNameCN 将语言代码映射为中文名称
func FriendlyLanguageNameCN(lang whatlanggo.Lang) string {
	langMapCN := map[whatlanggo.Lang]string{
		whatlanggo.Eng: "英文",
		whatlanggo.Cmn: "中文",
		whatlanggo.Spa: "西班牙文",
		whatlanggo.Por: "葡萄牙文",
		whatlanggo.Rus: "俄文",
		whatlanggo.Jpn: "日文",
		whatlanggo.Deu: "德文",
		whatlanggo.Kor: "韩文",
		whatlanggo.Fra: "法文",
		whatlanggo.Ita: "意大利文",
		whatlanggo.Tur: "土耳其文",
		whatlanggo.Pol: "波兰文",
		whatlanggo.Nld: "荷兰文",
		whatlanggo.Hin: "印地文",
		whatlanggo.Ben: "孟加拉文",
		whatlanggo.Vie: "越南文",
		whatlanggo.Ukr: "乌克兰文",
		whatlanggo.Swe: "瑞典文",
		whatlanggo.Fin: "芬兰文",
		whatlanggo.Dan: "丹麦文",
		whatlanggo.Heb: "希伯来文",
		whatlanggo.Tha: "泰文",
		// 根据需要添加更多语言
	}

	// 获取中文的语言名称，如果没有找到，则返回"未知语言"
	name, ok := langMapCN[lang]
	if !ok {
		return "未知语言"
	}
	return name
}

// LengthIntercept 检查文本长度，如果超过最大长度，则返回 true 并发送消息
func LengthIntercept(text string, message structs.OnebotGroupMessage) bool {
	maxLen := config.GetQuestionMaxLenth()
	if len(text) > maxLen {
		// 长度超出限制，获取并发送响应消息
		responseMessage := config.GetQmlResponseMessages()

		// 根据消息类型发送响应
		if message.RealMessageType == "group_private" || message.MessageType == "private" {
			if !config.GetUsePrivateSSE() {
				SendPrivateMessage(message.UserID, responseMessage)
			} else {
				SendSSEPrivateMessage(message.UserID, responseMessage)
			}
		} else {
			SendGroupMessage(message.GroupID, message.UserID, responseMessage)
		}

		return true // 拦截
	}
	return false // 长度符合要求，不拦截
}

// AddMessageID 为指定user_id添加新的消息ID
func AddMessageID(userID int64, messageID int64) {
	muUserIDMessageIDs.Lock()
	defer muUserIDMessageIDs.Unlock()

	// 消息ID的有效期是120秒
	expiration := time.Now().Add(120 * time.Second)
	messageInfo := MessageIDInfo{MessageID: messageID, Expires: expiration}

	// 清理已过期的消息ID
	cleanExpiredMessageIDs(userID)

	// 添加新的消息ID
	UserIDMessageIDs[userID] = append(UserIDMessageIDs[userID], messageInfo)
}

// cleanExpiredMessageIDs 清理指定user_id的已过期消息ID
func cleanExpiredMessageIDs(userID int64) {
	validMessageIDs := []MessageIDInfo{}
	for _, messageInfo := range UserIDMessageIDs[userID] {
		if messageInfo.Expires.After(time.Now()) {
			validMessageIDs = append(validMessageIDs, messageInfo)
		}
	}
	UserIDMessageIDs[userID] = validMessageIDs
}

// GetLatestValidMessageID 获取指定user_id当前有效的最新消息ID
func GetLatestValidMessageID(userID int64) (int64, bool) {
	muUserIDMessageIDs.RLock()
	defer muUserIDMessageIDs.RUnlock()

	// 确保已过期的消息ID被清理
	cleanExpiredMessageIDs(userID)

	// 获取最新的消息ID
	if len(UserIDMessageIDs[userID]) > 0 {
		latestMessageInfo := UserIDMessageIDs[userID][len(UserIDMessageIDs[userID])-1]
		return latestMessageInfo.MessageID, true
	}
	return 0, false
}

// sendDeleteRequest 发送删除消息的请求，并输出响应内容
func sendDeleteRequest(url string, requestBody []byte) error {
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

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 将响应体转换为字符串，并输出
	bodyString := string(bodyBytes)
	fmt.Println("Response Body:", bodyString)

	return nil
}

func DeleteLatestMessage(messageType string, id int64, userid int64) error {
	// 获取基础URL
	baseURL := config.GetHttpPath() // 假设config.GetHttpPath()返回基础URL

	// 构建完整的URL
	url := baseURL + "/delete_msg"

	// 获取最新的有效消息ID
	messageID, valid := GetLatestValidMessageID(userid)
	if !valid {
		return fmt.Errorf("no valid message ID found for user/group/guild ID: %d", id)
	}

	// 构造请求体
	requestBody := make(map[string]interface{})
	requestBody["message_id"] = strconv.FormatInt(messageID, 10)

	// 根据type填充相应的ID字段
	switch messageType {
	case "group_private":
		requestBody["user_id"] = id
	case "group":
		requestBody["group_id"] = id
	case "guild":
		requestBody["channel_id"] = id
	case "guild_private":
		requestBody["guild_id"] = id
	default:
		return fmt.Errorf("unsupported message type: %s", messageType)
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	fmtf.Printf("发送撤回请求:%v", string(requestBodyBytes))

	// 发送删除消息请求
	return sendDeleteRequest(url, requestBodyBytes)
}
