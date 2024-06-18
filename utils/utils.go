package utils

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
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
	"github.com/hoshinonyaruko/gensokyo-llm/promptkb"
	"github.com/hoshinonyaruko/gensokyo-llm/server"
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

var (
	baseURLMap   = make(map[string]string)
	baseURLMapMu sync.Mutex
)

// 结构体用于解析 JSON 响应
type loginInfoResponse struct {
	Status  string `json:"status"`
	Retcode int    `json:"retcode"`
	Data    struct {
		UserID   int64  `json:"user_id"`
		Nickname string `json:"nickname"`
	} `json:"data"`
}

// 构造 URL 并请求数据
func FetchAndStoreUserIDs() {
	httpPaths := config.GetHttpPaths()
	for _, baseURL := range httpPaths {
		url := baseURL + "/get_login_info"
		resp, err := http.Get(url)
		if err != nil {
			fmtf.Printf("Error fetching login info from %s: %v", url, err)
			continue
		}
		defer resp.Body.Close()

		var result loginInfoResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmtf.Printf("Error decoding response from %s: %v", url, err)
			continue
		}

		if result.Retcode == 0 && result.Status == "ok" {
			fmtf.Printf("成功绑定机器人selfid[%v]onebot api baseURL[%v]", result.Data.UserID, baseURL)
			baseURLMapMu.Lock()
			useridstr := strconv.FormatInt(result.Data.UserID, 10)
			baseURLMap[useridstr] = baseURL
			baseURLMapMu.Unlock()
		}
	}
}

// GetBaseURLByUserID 通过 user_id 获取 baseURL
func GetBaseURLByUserID(userID string) (string, bool) {
	baseURLMapMu.Lock()
	defer baseURLMapMu.Unlock()
	url, exists := baseURLMap[userID]
	return url, exists
}

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

func PrintChatCompletionsRequest(request *hunyuan.ChatCompletionsRequest) {

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
func ContainsRune(slice []rune, value rune, groupid int64, promptstr string) bool {
	var probability int
	if groupid == 0 {
		// 获取概率百分比
		probability = config.GetSplitByPuntuations(promptstr)
	} else {
		// 获取概率百分比
		probability = config.GetSplitByPuntuationsGroup(promptstr)
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

func SendGroupMessage(groupID int64, userID int64, message string, selfid string, promptstr string) error {
	//TODO: 用userid作为了echo,在ws收到回调信息的时候,加入到全局撤回数组,AddMessageID,实现撤回
	if server.IsSelfIDExists(selfid) {
		// 创建消息结构体
		msg := map[string]interface{}{
			"action": "send_group_msg",
			"params": map[string]interface{}{
				"group_id": groupID,
				"user_id":  userID,
				"message":  message,
			},
			"echo": userID,
		}

		// 发送消息
		return server.SendMessageBySelfID(selfid, msg)
	}
	var baseURL string
	if len(config.GetHttpPaths()) > 0 {
		baseURL, _ = GetBaseURLByUserID(selfid)
	} else {
		// 获取基础URL
		baseURL = config.GetHttpPath() // 假设config.getHttpPath()返回基础URL
	}

	// 构建完整的URL
	baseURL = baseURL + "/send_group_msg"

	// 获取PathToken并检查其是否为空
	pathToken := config.GetPathToken()
	// 使用net/url包构建URL
	u, err := url.Parse(baseURL)
	if err != nil {
		panic("URL parsing failed: " + err.Error())
	}

	// 添加access_token参数
	query := u.Query()
	if pathToken != "" {
		query.Set("access_token", pathToken)
	}
	u.RawQuery = query.Encode()

	if config.GetSensitiveModeType() == 1 {
		message = acnode.CheckWordOUT(message)
	}

	//精细化替换 每个yml配置文件都可以具有一个非全局的文本替换规则
	message = ReplaceTextOut(message, promptstr)

	// 去除末尾的换行符 不去除会导致不好看
	message = removeTrailingCRLFs(message)

	// 繁体转换简体 安全策略 防止用户诱导ai发繁体绕过替换规则
	message, err = ConvertTraditionalToSimplified(message)
	if err != nil {
		fmtf.Printf("繁体转换简体失败:%v", err)
	}

	// 构造请求体
	requestBody, err := json.Marshal(map[string]interface{}{
		"group_id": groupID,
		"user_id":  userID,
		"message":  message,
	})
	fmtf.Printf("发群信息请求:%v", string(requestBody))
	fmtf.Printf("实际发送信息:%v", message)
	if err != nil {
		return fmtf.Errorf("failed to marshal request body: %w", err)
	}

	// 发送POST请求
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(requestBody))
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

func SendGroupMessageMdPromptKeyboard(groupID int64, userID int64, message string, selfid string, newmsg string, response string, promptstr string) error {
	//TODO: 用userid作为了echo,在ws收到回调信息的时候,加入到全局撤回数组,AddMessageID,实现反向ws连接时候的撤回
	if server.IsSelfIDExists(selfid) {
		// 创建消息结构体
		msg := map[string]interface{}{
			"action": "send_group_msg",
			"params": map[string]interface{}{
				"group_id": groupID,
				"user_id":  userID,
				"message":  message,
			},
			"echo": userID,
		}

		// 发送消息
		return server.SendMessageBySelfID(selfid, msg)
	}
	// 获取基础URL
	baseURL := config.GetHttpPath() // 假设config.getHttpPath()返回基础URL

	// 构建完整的URL
	baseURL = baseURL + "/send_group_msg"

	// 获取PathToken并检查其是否为空
	pathToken := config.GetPathToken()
	// 使用net/url包构建URL
	u, err := url.Parse(baseURL)
	if err != nil {
		panic("URL parsing failed: " + err.Error())
	}

	// 添加access_token参数
	query := u.Query()
	if pathToken != "" {
		query.Set("access_token", pathToken)
	}
	u.RawQuery = query.Encode()

	if config.GetSensitiveModeType() == 1 {
		message = acnode.CheckWordOUT(message)
	}

	//精细化替换 每个yml配置文件都可以具有一个非全局的文本替换规则
	message = ReplaceTextOut(message, promptstr)

	// 去除末尾的换行符 不去除会导致不好看
	message = removeTrailingCRLFs(message)

	// 繁体转换简体 安全策略 防止用户诱导ai发繁体绕过替换规则
	message, err = ConvertTraditionalToSimplified(message)
	if err != nil {
		fmtf.Printf("繁体转换简体失败:%v", err)
	}

	// 首先获取当前的keyboard
	var promptkeyboard []string
	if !config.GetUseAIPromptkeyboard() {
		promptkeyboard = config.GetPromptkeyboard()
	} else {
		fmtf.Printf("ai生成气泡:%v", "Q"+newmsg+"A"+response)
		promptkeyboard = promptkb.GetPromptKeyboardAI("Q"+newmsg+"A"+response, promptstr)
	}

	// 使用acnode.CheckWordOUT()过滤promptkeyboard中的每个字符串
	for i, item := range promptkeyboard {
		promptkeyboard[i] = acnode.CheckWordOUT(item)
	}

	var mdContent string
	// 这里把message构造成一个cq,md码
	if config.GetMemoryListMD() == 2 {
		//构建Markdown内容，对promptkeyboard的内容进行URL编码
		var sb strings.Builder
		// 添加初始消息
		sb.WriteString(message)
		sb.WriteString("\r")
		lastIndex := len(promptkeyboard) - 1 // 获取最后一个索引
		// 遍历promptkeyboard数组，为每个元素生成一个标签
		for i, cmd := range promptkeyboard {
			// 对每个命令进行URL编码
			encodedCmd := url.QueryEscape(cmd)
			// 构建并添加qqbot-cmd-input标签
			if i == lastIndex {
				// 如果是最后一个元素，则不添加 \r
				sb.WriteString(fmt.Sprintf("<qqbot-cmd-input text=\"%s\" show=\"%s\" reference=\"false\" />", encodedCmd, encodedCmd))
			} else {
				sb.WriteString(fmt.Sprintf("<qqbot-cmd-input text=\"%s\" show=\"%s\" reference=\"false\" />\r", encodedCmd, encodedCmd))
			}
		}
		mdContent = sb.String()
	} else {
		mdContent = message
	}

	fmt.Println(mdContent)

	// 构建Buttons
	buttons := []structs.Button{}
	// 添加promptkeyboard的按钮，每个按钮一行
	for i, label := range promptkeyboard {
		buttons = append(buttons, structs.Button{
			ID: fmt.Sprintf("%d", i+1),
			RenderData: structs.RenderData{
				Label:        label,
				VisitedLabel: label,
				Style:        1,
			},
			Action: structs.Action{
				Type: 2,
				Permission: structs.Permission{
					Type:           2,
					SpecifyRoleIDs: []string{"1", "2", "3"},
				},
				Data:          label,
				UnsupportTips: "请升级新版手机QQ",
				Enter:         true,
				Reply:         true,
			},
		})
	}

	// 添加"重置", "撤回", "重发"按钮，它们在一个单独的行
	rowWithThreeButtons := []structs.Button{}
	labels := []string{"重置", "忽略", "记忆", "载入"}

	for i, label := range labels {
		actionType := 1
		if label == "载入" {
			actionType = 2 // 设置特定的 ActionType
		}

		button := structs.Button{
			ID: fmt.Sprintf("%d", i+4), // 确保ID不重复
			RenderData: structs.RenderData{
				Label:        label,
				VisitedLabel: label,
				Style:        1,
			},
			Action: structs.Action{
				Type: actionType, // 使用条件变量设置的 actionType
				Permission: structs.Permission{
					Type:           2,
					SpecifyRoleIDs: []string{"1", "2", "3"},
				},
				Data:          label,
				UnsupportTips: "请升级新版手机QQ",
			},
		}

		rowWithThreeButtons = append(rowWithThreeButtons, button)
	}

	// 构建完整的PromptKeyboardMarkdown对象
	var rows []structs.Row // 初始化一个空切片来存放行

	// GetMemoryListMD==1 将buttons添加到rows
	if config.GetMemoryListMD() == 1 {
		// 遍历所有按钮，并每个按钮创建一行
		for _, button := range buttons {
			row := structs.Row{
				Buttons: []structs.Button{button}, // 将当前按钮加入到新行中
			}
			rows = append(rows, row) // 将新行添加到行切片中
		}
	}

	// 添加特定的 rowWithThreeButtons 至 rows 数组的末尾
	row := structs.Row{
		Buttons: rowWithThreeButtons, // 将当前三个按钮放入
	}
	rows = append(rows, row)

	// 构建 PromptKeyboardMarkdown 结构体
	promptKeyboardMd := structs.PromptKeyboardMarkdown{
		Markdown: structs.Markdown{
			Content: mdContent,
		},
		Keyboard: structs.Keyboard{
			Content: structs.KeyboardContent{
				Rows: rows, // 使用动态创建的行数组
			},
		},
		Content:   "keyboard",
		MsgID:     "123",
		Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
		MsgType:   2,
	}

	// 序列化成JSON
	mdContentBytes, err := json.Marshal(promptKeyboardMd)
	if err != nil {
		fmt.Printf("Error marshaling to JSON: %v", err)
		return nil
	}

	// 编码成Base64
	encoded := base64.StdEncoding.EncodeToString(mdContentBytes)
	segmentContent := "[CQ:markdown,data=base64://" + encoded + "]"

	// 构造请求体
	requestBody, err := json.Marshal(map[string]interface{}{
		"group_id": groupID,
		"user_id":  userID,
		"message":  segmentContent,
	})
	fmtf.Printf("发群信息请求:%v", string(requestBody))
	fmtf.Printf("实际发送信息:%v", message)
	if err != nil {
		return fmtf.Errorf("failed to marshal request body: %w", err)
	}

	// 发送POST请求
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(requestBody))
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

func SendGroupMessageMdPromptKeyboardV2(groupID int64, userID int64, message string, selfid string, promptstr string, promptkeyboard []string) error {
	//TODO: 用userid作为了echo,在ws收到回调信息的时候,加入到全局撤回数组,AddMessageID,实现反向ws连接时候的撤回
	if server.IsSelfIDExists(selfid) {
		// 创建消息结构体
		msg := map[string]interface{}{
			"action": "send_group_msg",
			"params": map[string]interface{}{
				"group_id": groupID,
				"user_id":  userID,
				"message":  message,
			},
			"echo": userID,
		}

		// 发送消息
		return server.SendMessageBySelfID(selfid, msg)
	}
	// 获取基础URL
	baseURL := config.GetHttpPath() // 假设config.getHttpPath()返回基础URL

	// 构建完整的URL
	baseURL = baseURL + "/send_group_msg"

	// 获取PathToken并检查其是否为空
	pathToken := config.GetPathToken()
	// 使用net/url包构建URL
	u, err := url.Parse(baseURL)
	if err != nil {
		panic("URL parsing failed: " + err.Error())
	}

	// 添加access_token参数
	query := u.Query()
	if pathToken != "" {
		query.Set("access_token", pathToken)
	}
	u.RawQuery = query.Encode()

	if config.GetSensitiveModeType() == 1 {
		message = acnode.CheckWordOUT(message)
	}

	//精细化替换 每个yml配置文件都可以具有一个非全局的文本替换规则
	message = ReplaceTextOut(message, promptstr)

	// 去除末尾的换行符 不去除会导致不好看
	message = removeTrailingCRLFs(message)

	// 繁体转换简体 安全策略 防止用户诱导ai发繁体绕过替换规则
	message, err = ConvertTraditionalToSimplified(message)
	if err != nil {
		fmtf.Printf("繁体转换简体失败:%v", err)
	}
	var mdContent string
	// 这里把message构造成一个cq,md码
	if config.GetMemoryListMD() == 2 {
		//构建Markdown内容，对promptkeyboard的内容进行URL编码
		var sb strings.Builder
		// 添加初始消息
		sb.WriteString(message)
		sb.WriteString("\r")
		lastIndex := len(promptkeyboard) - 1 // 获取最后一个索引
		// 遍历promptkeyboard数组，为每个元素生成一个标签
		for i, cmd := range promptkeyboard {
			// 对每个命令进行URL编码
			encodedCmd := url.QueryEscape(cmd)
			// 构建并添加qqbot-cmd-input标签
			if i == lastIndex {
				// 如果是最后一个元素，则不添加 \r
				sb.WriteString(fmt.Sprintf("<qqbot-cmd-input text=\"%s\" show=\"%s\" reference=\"false\" />", encodedCmd, encodedCmd))
			} else {
				sb.WriteString(fmt.Sprintf("<qqbot-cmd-input text=\"%s\" show=\"%s\" reference=\"false\" />\r", encodedCmd, encodedCmd))
			}
		}
		mdContent = sb.String()
	} else {
		mdContent = message
	}

	fmt.Println(mdContent)
	var promptKeyboardMd structs.PromptKeyboardMarkdown
	if config.GetMemoryListMD() == 1 {
		// 构建Buttons
		buttons := []structs.Button{}
		// 添加promptkeyboard的按钮，每个按钮一行
		for i, label := range promptkeyboard {
			buttons = append(buttons, structs.Button{
				ID: fmt.Sprintf("%d", i+1),
				RenderData: structs.RenderData{
					Label:        label,
					VisitedLabel: "已载入",
					Style:        1,
				},
				Action: structs.Action{
					Type: 2,
					Permission: structs.Permission{
						Type:           2,
						SpecifyRoleIDs: []string{"1", "2", "3"},
					},
					Data:          label,
					UnsupportTips: "请升级新版手机QQ",
					Enter:         true,
					Reply:         true,
				},
			})
		}

		// 构建完整的PromptKeyboardMarkdown对象
		var rows []structs.Row // 初始化一个空切片来存放行

		// 遍历所有按钮，并每个按钮创建一行
		for _, button := range buttons {
			row := structs.Row{
				Buttons: []structs.Button{button}, // 将当前按钮加入到新行中
			}
			rows = append(rows, row) // 将新行添加到行切片中
		}

		// 构建 PromptKeyboardMarkdown 结构体
		promptKeyboardMd = structs.PromptKeyboardMarkdown{
			Markdown: structs.Markdown{
				Content: mdContent,
			},
			Keyboard: structs.Keyboard{
				Content: structs.KeyboardContent{
					Rows: rows, // 使用动态创建的行数组
				},
			},
			Content:   "keyboard",
			MsgID:     "123",
			Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
			MsgType:   2,
		}
	} else {
		// 构建 PromptKeyboardMarkdown 结构体
		promptKeyboardMd = structs.PromptKeyboardMarkdown{
			Markdown: structs.Markdown{
				Content: mdContent,
			},
			Content:   "keyboard",
			MsgID:     "123",
			Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
			MsgType:   2,
		}
	}

	// 序列化成JSON
	mdContentBytes, err := json.Marshal(promptKeyboardMd)
	if err != nil {
		fmt.Printf("Error marshaling to JSON: %v", err)
		return nil
	}

	// 编码成Base64
	encoded := base64.StdEncoding.EncodeToString(mdContentBytes)
	segmentContent := "[CQ:markdown,data=base64://" + encoded + "]"

	// 构造请求体
	requestBody, err := json.Marshal(map[string]interface{}{
		"group_id": groupID,
		"user_id":  userID,
		"message":  segmentContent,
	})
	fmtf.Printf("发群信息请求:%v", string(requestBody))
	fmtf.Printf("实际发送信息:%v", message)
	if err != nil {
		return fmtf.Errorf("failed to marshal request body: %w", err)
	}

	// 发送POST请求
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(requestBody))
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

func SendPrivateMessage(UserID int64, message string, selfid string, promptstr string) error {
	if server.IsSelfIDExists(selfid) {
		// 创建消息结构体
		msg := map[string]interface{}{
			"action": "send_private_msg",
			"params": map[string]interface{}{
				"user_id": UserID,
				"message": message,
			},
			"echo": UserID,
		}

		// 发送消息
		return server.SendMessageBySelfID(selfid, msg)
	}
	var baseURL string
	if len(config.GetHttpPaths()) > 0 {
		baseURL, _ = GetBaseURLByUserID(selfid)
	} else {
		// 获取基础URL
		baseURL = config.GetHttpPath() // 假设config.getHttpPath()返回基础URL
	}

	// 构建完整的URL
	baseURL = baseURL + "/send_private_msg"

	// 获取PathToken并检查其是否为空
	pathToken := config.GetPathToken()
	// 使用net/url包构建URL
	u, err := url.Parse(baseURL)
	if err != nil {
		panic("URL parsing failed: " + err.Error())
	}

	// 添加access_token参数
	query := u.Query()
	if pathToken != "" {
		query.Set("access_token", pathToken)
	}
	u.RawQuery = query.Encode()

	if config.GetSensitiveModeType() == 1 {
		message = acnode.CheckWordOUT(message)
	}

	//精细化替换 每个yml配置文件都可以具有一个非全局的文本替换规则
	message = ReplaceTextOut(message, promptstr)

	// 去除末尾的换行符 不去除会导致不好看
	message = removeTrailingCRLFs(message)

	// 繁体转换简体 安全策略 防止用户诱导ai发繁体绕过替换规则
	message, err = ConvertTraditionalToSimplified(message)
	if err != nil {
		fmtf.Printf("繁体转换简体失败:%v", err)
	}

	// 构造请求体
	requestBody, err := json.Marshal(map[string]interface{}{
		"user_id": UserID,
		"message": message,
	})

	if err != nil {
		return fmtf.Errorf("failed to marshal request body: %w", err)
	}
	fmtf.Printf("实际发送信息:%v", message)

	// 发送POST请求
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(requestBody))
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

func SendPrivateMessageRaw(UserID int64, message string, selfid string) error {
	if server.IsSelfIDExists(selfid) {
		// 创建消息结构体
		msg := map[string]interface{}{
			"action": "send_private_msg",
			"params": map[string]interface{}{
				"user_id": UserID,
				"message": message,
			},
			"echo": UserID,
		}

		// 发送消息
		return server.SendMessageBySelfID(selfid, msg)
	}
	var baseURL string
	if len(config.GetHttpPaths()) > 0 {
		baseURL, _ = GetBaseURLByUserID(selfid)
	} else {
		// 获取基础URL
		baseURL = config.GetHttpPath() // 假设config.getHttpPath()返回基础URL
	}

	// 构建完整的URL
	baseURL = baseURL + "/send_private_msg"

	// 获取PathToken并检查其是否为空
	pathToken := config.GetPathToken()
	// 使用net/url包构建URL
	u, err := url.Parse(baseURL)
	if err != nil {
		panic("URL parsing failed: " + err.Error())
	}

	// 添加access_token参数
	query := u.Query()
	if pathToken != "" {
		query.Set("access_token", pathToken)
	}
	u.RawQuery = query.Encode()

	// 构造请求体
	requestBody, err := json.Marshal(map[string]interface{}{
		"user_id": UserID,
		"message": message,
	})

	if err != nil {
		return fmtf.Errorf("failed to marshal request body: %w", err)
	}

	// 发送POST请求
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(requestBody))
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

func SendPrivateMessageSSE(UserID int64, message structs.InterfaceBody, promptstr string) error {
	// 获取基础URL
	baseURL := config.GetHttpPath() // 假设config.GetHttpPath()返回基础URL

	// 构建完整的URL
	baseURL = baseURL + "/send_private_msg_sse"

	// 获取PathToken并检查其是否为空
	pathToken := config.GetPathToken()
	// 使用net/url包构建URL
	u, err := url.Parse(baseURL)
	if err != nil {
		panic("URL parsing failed: " + err.Error())
	}

	// 添加access_token参数
	query := u.Query()
	if pathToken != "" {
		query.Set("access_token", pathToken)
	}
	u.RawQuery = query.Encode()

	// 调试用的
	if config.GetPrintHanming() {
		fmtf.Printf("流式信息替换前:%v", message.Content)
	}

	// 检查是否需要启用敏感词过滤
	if config.GetSensitiveModeType() == 1 && message.Content != "" {
		message.Content = acnode.CheckWordOUT(message.Content)
	}

	//精细化替换 每个yml配置文件都可以具有一个非全局的文本替换规则
	message.Content = ReplaceTextOut(message.Content, promptstr)

	// 调试用的
	if config.GetPrintHanming() {
		fmtf.Printf("流式信息替换后:%v", message.Content)
	}

	// 去除末尾的换行符 不去除会导致sse接口始终等待
	message.Content = removeTrailingCRLFs(message.Content)

	// 繁体转换简体 安全策略 防止用户诱导ai发繁体绕过替换规则
	message.Content, err = ConvertTraditionalToSimplified(message.Content)
	if err != nil {
		fmtf.Printf("繁体转换简体失败:%v", err)
	}

	if message.Content == "" {
		message.Content = " "
		fmtf.Printf("过滤空SendPrivateMessageSSE,可能是llm api只发了换行符.")
		return nil
	}

	fmtf.Printf("实际发送信息:%v", message.Content)

	// 构造请求体，包括InterfaceBody
	requestBody, err := json.Marshal(map[string]interface{}{
		"user_id": UserID,
		"message": message,
	})
	if err != nil {
		return fmtf.Errorf("failed to marshal request body: %w", err)
	}

	// 发送POST请求
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(requestBody))
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

// removeTrailingCRLFs 移除字符串末尾的所有CRLF换行符
func removeTrailingCRLFs(input string) string {
	// 将字符串转换为字节切片
	byteMessage := []byte(input)

	// CRLF的字节表示
	crlf := []byte{'\r', '\n'}

	// 循环移除末尾的CRLF
	for bytes.HasSuffix(byteMessage, crlf) {
		byteMessage = bytes.TrimSuffix(byteMessage, crlf)
	}

	// LFLF的字节表示
	lflf := []byte{'\n', '\n'}

	// 循环移除末尾的LFLF
	for bytes.HasSuffix(byteMessage, lflf) {
		byteMessage = bytes.TrimSuffix(byteMessage, lflf)
	}

	// 将处理后的字节切片转换回字符串
	return string(byteMessage)
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

// RemoveAtTagContentConditional 接收一个字符串和一个 int64 类型的 selfid，
// 并根据条件移除[@selfowd]格式的内容，然后去除前后的空格。
// 只有当标签中的ID与传入的selfid相匹配时才进行移除，
// 如果没有任何[@xxx]出现或者所有出现的[@xxx]中的xxx都不等于selfid时，返回原字符串。
func RemoveAtTagContentConditionalWithoutAddNick(input string, message structs.OnebotGroupMessage) string {
	// 将 int64 类型的 selfid 转换为字符串
	selfidStr := strconv.FormatInt(message.SelfID, 10)

	// 编译一个正则表达式，用于匹配并捕获[@任意字符]的模式
	re := regexp.MustCompile(`\[@(.*?)\]`)
	matches := re.FindAllStringSubmatch(input, -1)

	// 如果没有找到任何匹配项,直接返回原输入,代表不带at的信息,会在更上方判断是否处理.同时根据配置项为请求增加名字.
	if len(matches) == 0 { //私聊无法at 只会走这里,只会有nick生效
		return input
	}

	foundSelfId := false // 用于跟踪是否找到与selfid相匹配的标签

	// 遍历所有匹配项
	for _, match := range matches {
		if match[1] == selfidStr {
			// 如果找到与selfid相匹配的标签，替换该标签
			input = strings.Replace(input, match[0], "", -1)
			foundSelfId = true // 标记已找到至少一个与selfid相匹配的标签
		} else {
			input = strings.Replace(input, match[0], "", -1)
		}
	}

	// 只有在包含了at 但是at不包含自己,才忽略信息
	if !foundSelfId {
		// 如果没有找到任何与selfid相匹配的标签，将输入置为空,代表不响应这一条信息
		input = ""
	}

	// 去除前后的空格
	cleaned := strings.TrimSpace(input)
	return cleaned
}

// RemoveAtTagContentConditional 接收一个字符串和一个 int64 类型的 selfid，
// 并根据条件移除[@selfowd]格式的内容，然后去除前后的空格。
// 只有当标签中的ID与传入的selfid相匹配时才进行移除，
// 如果没有任何[@xxx]出现或者所有出现的[@xxx]中的xxx都不等于selfid时，返回原字符串。
func RemoveAtTagContentConditional(input string, message structs.OnebotGroupMessage, promptstr string) string {

	// 获取特殊名称替换对数组
	specialNames := config.GetSpecialNameToQ(promptstr)

	// 将 int64 类型的 selfid 转换为字符串
	selfidStr := strconv.FormatInt(message.SelfID, 10)

	// 编译一个正则表达式，用于匹配并捕获[@任意字符]的模式
	re := regexp.MustCompile(`\[@(.*?)\]`)
	matches := re.FindAllStringSubmatch(input, -1)

	// 如果没有找到任何匹配项,直接返回原输入,代表不带at的信息,会在更上方判断是否处理.同时根据配置项为请求增加名字.
	if len(matches) == 0 { //私聊无法at 只会走这里,只会有nick生效
		var name string
		name = ""
		// 可以都是2 打开 呈现覆盖关系
		if config.GetGroupAddCardToQ(promptstr) == 2 {
			if message.Sender.Card != "" {
				name = "[username:" + message.Sender.Card + "]"
			}
		}

		if config.GetGroupAddNicknameToQ(promptstr) == 2 && name == "" {
			if message.Sender.Nickname != "" {
				name = "[username:" + message.Sender.Nickname + "]"
			}
		}

		useridstr := strconv.FormatInt(message.UserID, 10)
		// 遍历特殊名称数组，检查是否需要进行进一步替换
		for _, replacement := range specialNames {
			if useridstr == replacement.ID {
				name = "[username:" + replacement.Name + "]"
				break // 找到匹配，跳出循环
			}
		}

		if name != "" {
			input = name + input
		}

		return input
	}

	foundSelfId := false // 用于跟踪是否找到与selfid相匹配的标签

	// 遍历所有匹配项
	for _, match := range matches {
		if match[1] == selfidStr {
			// 把at自己的信息替换为当前at自己的人的昵称或者群名片
			var name string
			name = "" // 初始状态

			// 可以都是2 打开 呈现覆盖关系
			if config.GetGroupAddCardToQ(promptstr) == 2 {
				if message.Sender.Card != "" {
					name = "[username:" + message.Sender.Card + "]"
				}
			}

			if config.GetGroupAddNicknameToQ(promptstr) == 2 && name == "" {
				if message.Sender.Nickname != "" {
					name = "[username:" + message.Sender.Nickname + "]"
				}
			}

			useridstr := strconv.FormatInt(message.UserID, 10)
			// 遍历特殊名称数组，检查是否需要进行进一步替换
			for _, replacement := range specialNames {
				if useridstr == replacement.ID {
					name = "[username:" + replacement.Name + "]"
					break // 找到匹配，跳出循环
				}
			}

			// 如果找到与selfid相匹配的标签，替换该标签
			input = strings.Replace(input, match[0], name, -1)
			foundSelfId = true // 标记已找到至少一个与selfid相匹配的标签
		} else {
			var name string
			name = ""

			// 可以都是2 打开 呈现覆盖关系
			if config.GetGroupAddCardToQ(promptstr) == 2 {
				if message.Sender.Card != "" {
					name = "[username:" + message.Sender.Card + "]"
				}
			}

			// 将CQat标签替换为名字
			if config.GetGroupAddNicknameToQ(promptstr) == 2 && name == "" {
				if message.Sender.Nickname != "" {
					name = "[username:" + message.Sender.Nickname + "]"
				}
			}

			// 遍历特殊名称数组，检查是否需要进行进一步替换
			for _, replacement := range specialNames {
				if match[1] == replacement.ID {
					name = "[username:" + replacement.Name + "]"
					break // 找到匹配，跳出循环
				}
			}

			input = strings.Replace(input, match[0], name, -1)
		}
	}

	// 只有在包含了at 但是at不包含自己,才忽略信息
	if !foundSelfId {
		// 如果没有找到任何与selfid相匹配的标签，将输入置为空,代表不响应这一条信息
		input = ""
	}

	// 去除前后的空格
	cleaned := strings.TrimSpace(input)
	return cleaned
}

func PostSensitiveMessages() error {
	port := config.GetPort() // 从config包获取端口号
	var portStr string
	if config.GetLotus() == "" {
		portStr = fmt.Sprintf("http://127.0.0.1:%d/gensokyo", port) // 根据端口号构建URL
	} else {
		portStr = config.GetLotus() + "/gensokyo"
	}

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
func SendSSEPrivateMessage(userID int64, content string, promptstr string) {
	punctuations := []rune{'。', '！', '？', '，', ',', '.', '!', '?', '~'}
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
			MemoryLoadCommand := config.GetMemoryLoadCommand()
			promptKeyboard := config.GetPromptkeyboard()

			if len(MemoryLoadCommand) > 0 {
				selectedRestoreResponse := MemoryLoadCommand[rand.Intn(len(MemoryLoadCommand))]
				if len(promptKeyboard) > 0 {
					promptKeyboard[0] = selectedRestoreResponse
				}
			}

			messageSSE.PromptKeyboard = promptKeyboard
		}

		// 发送SSE消息函数
		SendPrivateMessageSSE(userID, messageSSE, promptstr)
	}
}

// SendSSEPrivateMessageWithKeyboard 分割并发送消息的核心逻辑，直接遍历字符串
func SendSSEPrivateMessageWithKeyboard(userID int64, content string, keyboard []string, promptstr string) {
	punctuations := []rune{'。', '！', '？', '，', ',', '.', '!', '?', '~'}
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
			var promptKeyboard []string
			if len(keyboard) == 0 {
				RestoreResponses := config.GetRestoreCommand()
				promptKeyboard = config.GetPromptkeyboard()

				if len(RestoreResponses) > 0 {
					selectedRestoreResponse := RestoreResponses[rand.Intn(len(RestoreResponses))]
					if len(promptKeyboard) > 0 {
						promptKeyboard[0] = selectedRestoreResponse
					}
				}
			} else {
				promptKeyboard = keyboard
			}

			messageSSE.PromptKeyboard = promptKeyboard
		}

		// 发送SSE消息函数
		SendPrivateMessageSSE(userID, messageSSE, promptstr)
	}
}

// SendSSEPrivateMessageByline 分割并发送消息的核心逻辑，直接遍历字符串
func SendSSEPrivateMessageByLine(userID int64, content string, keyboard []string, promptstr string) {
	// 直接使用 strings.Split 按行分割字符串
	parts := strings.Split(content, "\n")

	// 根据parts长度处理状态
	for i, part := range parts {
		if part == "" {
			continue // 跳过空行
		}

		state := 1
		if i == len(parts)-2 { // 倒数第二部分
			state = 11
		} else if i == len(parts)-1 { // 最后一部分
			state = 20
		}

		// 构造消息体并发送
		messageSSE := structs.InterfaceBody{
			Content: part + "\n",
			State:   state,
		}

		if state == 20 { // 对最后一部分特殊处理
			var promptKeyboard []string
			if len(keyboard) == 0 {
				RestoreResponses := config.GetRestoreCommand()
				promptKeyboard = config.GetPromptkeyboard()

				if len(RestoreResponses) > 0 {
					selectedRestoreResponse := RestoreResponses[rand.Intn(len(RestoreResponses))]
					if len(promptKeyboard) > 0 {
						promptKeyboard[0] = selectedRestoreResponse
					}
				}
			} else {
				promptKeyboard = keyboard
			}

			messageSSE.PromptKeyboard = promptKeyboard
		}

		// 发送SSE消息函数
		SendPrivateMessageSSE(userID, messageSSE, promptstr)
	}
}

// SendSSEPrivateSafeMessage 分割并发送安全消息的核心逻辑，直接遍历字符串
func SendSSEPrivateSafeMessage(userID int64, saveresponse string, promptstr string) {
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

	SendPrivateMessageSSE(userID, messageSSE, promptstr)

	// 中间
	messageSSE = structs.InterfaceBody{
		Content: parts[1],
		State:   11,
	}
	SendPrivateMessageSSE(userID, messageSSE, promptstr)

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
	SendPrivateMessageSSE(userID, messageSSE, promptstr)
}

// SendSSEPrivateRestoreMessage 分割并发送重置消息的核心逻辑，直接遍历字符串
func SendSSEPrivateRestoreMessage(userID int64, RestoreResponse string, promptstr string) {
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

	SendPrivateMessageSSE(userID, messageSSE, promptstr)

	//中间
	messageSSE = structs.InterfaceBody{
		Content: parts[1],
		State:   11,
	}
	SendPrivateMessageSSE(userID, messageSSE, promptstr)

	// 从配置中获取promptkeyboard
	promptkeyboard := config.GetPromptkeyboard()

	// 创建InterfaceBody结构体实例
	messageSSE = structs.InterfaceBody{
		Content:        parts[2],       // 假设空格字符串是期望的内容
		State:          20,             // 假设的状态码
		PromptKeyboard: promptkeyboard, // 使用更新后的promptkeyboard
	}

	// 发送SSE私人消息
	SendPrivateMessageSSE(userID, messageSSE, promptstr)
}

// LanguageIntercept 检查文本语言，如果不在允许列表中，则返回 true 并发送消息
func LanguageIntercept(text string, message structs.OnebotGroupMessage, selfid string, promptstr string) bool {
	hintWords := config.GetGroupHintWords()
	// 遍历所有触发词，将其从文本中剔除
	for _, word := range hintWords {
		text = strings.Replace(text, word, "", -1)
	}
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
			SendPrivateMessage(message.UserID, responseMessage, selfid, promptstr)
		} else {
			SendSSEPrivateMessage(message.UserID, responseMessage, promptstr)
		}
	} else {
		SendGroupMessage(message.GroupID, message.UserID, responseMessage, selfid, promptstr)
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
func LengthIntercept(text string, message structs.OnebotGroupMessage, selfid string, promptstr string) bool {
	maxLen := config.GetQuestionMaxLenth()
	if len(text) > maxLen {
		// 长度超出限制，获取并发送响应消息
		responseMessage := config.GetQmlResponseMessages()

		// 根据消息类型发送响应
		if message.RealMessageType == "group_private" || message.MessageType == "private" {
			if !config.GetUsePrivateSSE() {
				SendPrivateMessage(message.UserID, responseMessage, selfid, promptstr)
			} else {
				SendSSEPrivateMessage(message.UserID, responseMessage, promptstr)
			}
		} else {
			SendGroupMessage(message.GroupID, message.UserID, responseMessage, selfid, promptstr)
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
	baseURL = baseURL + "/delete_msg"

	// 获取PathToken并检查其是否为空
	pathToken := config.GetPathToken()
	// 使用net/url包构建URL
	u, err := url.Parse(baseURL)
	if err != nil {
		panic("URL parsing failed: " + err.Error())
	}

	// 添加access_token参数
	query := u.Query()
	if pathToken != "" {
		query.Set("access_token", pathToken)
	}
	u.RawQuery = query.Encode()

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
	case "interaction":
		requestBody["group_id"] = id
	default:
		return fmt.Errorf("unsupported message type: %s", messageType)
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	fmtf.Printf("发送撤回请求:%v", string(requestBodyBytes))

	// 发送删除消息请求
	return sendDeleteRequest(u.String(), requestBodyBytes)
}

// MakeAlternating ensures that roles alternate between "user" and "assistant".
func MakeAlternating(messages []structs.MessageContent) []structs.MessageContent {
	if len(messages) < 2 {
		return messages // Not enough messages to need alternation or check
	}

	// Initialize placeholders for the last seen user and assistant content
	var lastUserContent, lastAssistantContent []structs.ContentItem

	correctedMessages := make([]structs.MessageContent, 0, len(messages))
	expectedRole := "user" // Start expecting "user" initially; this changes as we find roles

	for _, message := range messages {
		if message.Role != expectedRole {
			// If the current message does not match the expected role, insert the last seen content of the expected role
			if expectedRole == "user" && lastUserContent != nil {
				correctedMessages = append(correctedMessages, structs.MessageContent{Role: "user", Content: lastUserContent})
			} else if expectedRole == "assistant" && lastAssistantContent != nil {
				correctedMessages = append(correctedMessages, structs.MessageContent{Role: "assistant", Content: lastAssistantContent})
			}
		}

		// Append the current message and update last seen contents
		correctedMessages = append(correctedMessages, message)
		if message.Role == "user" {
			lastUserContent = message.Content
			expectedRole = "assistant"
		} else if message.Role == "assistant" {
			lastAssistantContent = message.Content
			expectedRole = "user"
		}
	}

	return correctedMessages
}

// ReplaceTextIn 使用给定的替换对列表对文本进行替换
func ReplaceTextIn(text string, promptstr string) string {
	// 调用 GetReplacementPairsIn 函数获取替换对列表
	replacementPairs := config.GetReplacementPairsIn(promptstr)

	if len(replacementPairs) == 0 {
		return text
	}

	// 遍历所有的替换对，并在文本中进行替换
	for _, pair := range replacementPairs {
		// 使用 strings.Replace 替换文本中的所有出现
		// 注意这里我们使用 -1 作为最后的参数，表示替换文本中的所有匹配项
		text = strings.Replace(text, pair.OriginalWord, pair.TargetWord, -1)
	}

	// 返回替换后的文本
	return text
}

// ReplaceTextOut 使用给定的替换对列表对文本进行替换
func ReplaceTextOut(text string, promptstr string) string {
	// 调用 GetReplacementPairsIn 函数获取替换对列表
	replacementPairs := config.GetReplacementPairsOut(promptstr)

	if len(replacementPairs) == 0 {
		return text
	}

	// 遍历所有的替换对，并在文本中进行替换
	for _, pair := range replacementPairs {
		// 使用 strings.Replace 替换文本中的所有出现
		// 注意这里我们使用 -1 作为最后的参数，表示替换文本中的所有匹配项
		text = strings.Replace(text, pair.OriginalWord, pair.TargetWord, -1)
	}

	// 返回替换后的文本
	return text
}
