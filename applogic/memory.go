package applogic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
	"github.com/hoshinonyaruko/gensokyo-llm/utils"
)

// 请你为我将我所输入的对话浓缩总结为一个标题
func GetMemoryTitle(msg string) string {
	baseurl := config.GetAIPromptkeyboardPath()
	fmtf.Printf("获取到keyboard baseurl(复用):%v", baseurl)
	// 使用net/url包来构建和编码URL
	urlParams := url.Values{}
	urlParams.Add("prompt", "memory")

	// 将查询参数编码后附加到基本URL上
	fullURL := baseurl
	if len(urlParams) > 0 {
		fullURL += "?" + urlParams.Encode()
	}

	fmtf.Printf("Generated GetMemoryTitle URL:%v\n", fullURL)

	requestBody, err := json.Marshal(map[string]interface{}{
		"message":         msg,
		"conversationId":  "",
		"parentMessageId": "",
		"user_id":         "",
	})

	if err != nil {
		fmt.Printf("Error marshalling request: %v\n", err)
		return "默认标题"
	}

	resp, err := http.Post(fullURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return "默认标题"
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return "默认标题"
	}
	fmt.Printf("Response: %s\n", string(responseBody))

	var responseData ResponseDataPromptKeyboard
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		fmt.Printf("Error unmarshalling response data: %v[%v]\n", err, string(responseBody))
		return "默认标题"
	}

	// 预处理响应数据，移除可能的换行符
	preprocessedResponse := strings.TrimSpace(responseData.Response)

	// 去除所有标点符号和空格
	cleanedResponse := strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) || unicode.IsSpace(r) {
			return -1 // 在 strings.Map 中，返回 -1 表示删除字符
		}
		return r
	}, preprocessedResponse)

	return cleanedResponse
}

// 保存记忆
func (app *App) handleSaveMemory(msg structs.OnebotGroupMessage, ConversationID string, ParentMessageID string, promptstr string) {
	conversationTitle := "2024-5-19/18:26" // 默认标题，根据实际需求可能需要调整为动态生成的时间戳

	userid := msg.UserID
	if config.GetGroupContext() == 2 && msg.MessageType != "private" {
		userid = msg.GroupID + msg.SelfID
	}

	// 添加用户记忆
	err := app.AddUserMemory(userid, ConversationID, ParentMessageID, conversationTitle)
	if err != nil {
		log.Printf("Error saving memory: %s", err)
		return
	}

	// 启动Go routine进行历史信息处理和标题更新
	go func() {
		userHistory, err := app.getHistory(ConversationID, ParentMessageID)
		if err != nil {
			log.Printf("Error retrieving history: %s", err)
			return
		}

		// 处理历史信息为特定格式的字符串
		memoryTitle := formatHistory(userHistory)
		newTitle := GetMemoryTitle(memoryTitle) // 获取最终的记忆标题

		// 更新记忆标题
		err = app.updateConversationTitle(userid, ConversationID, ParentMessageID, newTitle)
		if err != nil {
			log.Printf("Error updating conversation title: %s", err)
		}
	}()

	var keyboard []string // 准备一个空的键盘数组
	// 获取记忆载入命令
	memoryLoadCommands := config.GetMemoryLoadCommand()
	if len(memoryLoadCommands) > 0 {
		keyboard = append(keyboard, memoryLoadCommands[0]) // 添加第一个命令到键盘数组
	}

	// 发送保存成功的响应
	saveMemoryResponse := "记忆保存成功！"
	app.sendMemoryResponseWithkeyBoard(msg, saveMemoryResponse, keyboard, promptstr)
}

// 获取记忆列表
func (app *App) handleMemoryList(msg structs.OnebotGroupMessage, promptstr string) {

	userid := msg.UserID
	if config.GetGroupContext() == 2 && msg.MessageType != "private" {
		userid = msg.GroupID + msg.SelfID
	}

	memories, err := app.GetUserMemories(userid)
	if err != nil {
		log.Printf("Error retrieving memories: %s", err)
		return
	}

	// 组合格式化的文本
	var responseBuilder strings.Builder
	responseBuilder.WriteString("当前记忆列表：\n")

	// 准备键盘数组，最多包含4个标题
	var keyboard []string
	var loadMemoryCommand string
	// 获取载入记忆指令
	memoryLoadCommands := config.GetMemoryLoadCommand()
	if len(memoryLoadCommands) > 0 {
		loadMemoryCommand = memoryLoadCommands[0]
	} else {
		loadMemoryCommand = "未设置载入指令"
	}

	for _, memory := range memories {
		if config.GetMemoryListMD() == 0 {
			responseBuilder.WriteString(memory.ConversationTitle + "\n")
		}
		keyboard = append(keyboard, loadMemoryCommand+" "+memory.ConversationTitle) // 添加新的标题
	}

	var exampleTitle string
	if len(memories) > 0 {
		exampleTitle = string([]rune(memories[0].ConversationTitle)[:3])
	}

	if config.GetMemoryListMD() == 0 {
		responseBuilder.WriteString(fmt.Sprintf("提示：发送 %s 任意标题开头的前n字即可载入记忆\n如：%s %s", loadMemoryCommand, loadMemoryCommand, exampleTitle))
	} else {
		if len(keyboard) == 0 {
			responseBuilder.WriteString("目前还没有对话记忆...点按记忆按钮来保存新的记忆吧")
		} else {
			responseBuilder.WriteString("点击蓝色文字载入记忆")
		}

	}

	// 发送组合后的信息，包括键盘数组
	app.sendMemoryResponseByline(msg, responseBuilder.String(), keyboard, promptstr)
}

// 载入记忆
func (app *App) handleLoadMemory(msg structs.OnebotGroupMessage, checkResetCommand string, promptstr string) {

	userid := msg.UserID
	if config.GetGroupContext() == 2 && msg.MessageType != "private" {
		userid = msg.GroupID + msg.SelfID
	}

	// 从配置获取载入记忆指令
	memoryLoadCommands := config.GetMemoryLoadCommand()

	// 移除所有载入记忆指令部分
	for _, command := range memoryLoadCommands {
		checkResetCommand = strings.Replace(checkResetCommand, command, "", -1)
	}

	// 移除空格得到匹配词
	matchTerm := strings.TrimSpace(checkResetCommand)

	// 获取用户记忆
	memories, err := app.GetUserMemories(userid)
	if err != nil {
		log.Printf("Error retrieving memories: %s", err)
		app.sendMemoryResponse(msg, "获取记忆失败", promptstr)
		return
	}

	// 查找匹配的记忆
	var matchedMemory *structs.Memory
	for _, memory := range memories {
		if strings.HasPrefix(memory.ConversationTitle, matchTerm) {
			matchedMemory = &memory
			break
		}
	}

	if matchedMemory == nil {
		app.sendMemoryResponse(msg, "未找到匹配的记忆", promptstr)
		return
	}

	// 载入记忆
	err = app.updateUserContextPro(userid, matchedMemory.ConversationID, matchedMemory.ParentMessageID)
	if err != nil {
		log.Printf("Error adding memory: %s", err)
		app.sendMemoryResponse(msg, "载入记忆失败", promptstr)
		return
	}

	// 组合回复信息
	responseMessage := fmt.Sprintf("成功载入了标题为 '%s' 的记忆", matchedMemory.ConversationTitle)
	app.sendMemoryResponse(msg, responseMessage, promptstr)
}

func (app *App) sendMemoryResponseWithkeyBoard(msg structs.OnebotGroupMessage, response string, keyboard []string, promptstr string) {
	strSelfID := strconv.FormatInt(msg.SelfID, 10)
	if msg.RealMessageType == "group_private" || msg.MessageType == "private" {
		if !config.GetUsePrivateSSE() {
			utils.SendPrivateMessage(msg.UserID, response, strSelfID, promptstr)
		} else {
			utils.SendSSEPrivateMessageWithKeyboard(msg.UserID, response, keyboard, promptstr)
		}
	} else {
		utils.SendGroupMessage(msg.GroupID, msg.UserID, response, strSelfID, promptstr)
	}
}

func (app *App) sendMemoryResponse(msg structs.OnebotGroupMessage, response string, promptstr string) {
	strSelfID := strconv.FormatInt(msg.SelfID, 10)
	if msg.RealMessageType == "group_private" || msg.MessageType == "private" {
		if !config.GetUsePrivateSSE() {
			utils.SendPrivateMessage(msg.UserID, response, strSelfID, promptstr)
		} else {
			utils.SendSSEPrivateMessage(msg.UserID, response, promptstr)
		}
	} else {
		utils.SendGroupMessage(msg.GroupID, msg.UserID, response, strSelfID, promptstr)
	}
}

func (app *App) sendMemoryResponseByline(msg structs.OnebotGroupMessage, response string, keyboard []string, promptstr string) {
	strSelfID := strconv.FormatInt(msg.SelfID, 10)
	if msg.RealMessageType == "group_private" || msg.MessageType == "private" {
		if !config.GetUsePrivateSSE() {
			utils.SendPrivateMessage(msg.UserID, response, strSelfID, promptstr)
		} else {
			// 更新键盘数组，确保最多只有三个元素
			if len(keyboard) >= 3 {
				keyboard = keyboard[:3]
			}
			utils.SendSSEPrivateMessageByLine(msg.UserID, response, keyboard, promptstr)
		}
	} else {
		if config.GetMemoryListMD() == 0 {
			utils.SendGroupMessage(msg.GroupID, msg.UserID, response, strSelfID, promptstr)
		} else {
			// 更新键盘数组，确保最多只有五个元素
			if len(keyboard) >= 5 {
				keyboard = keyboard[:5]
			}
			utils.SendGroupMessageMdPromptKeyboardV2(msg.GroupID, msg.UserID, response, strSelfID, promptstr, keyboard)
		}
	}
}

func formatHistory(history []structs.Message) string {
	var result string
	for _, message := range history {
		rolePrefix := "Q:"
		if message.Role == "answer" {
			rolePrefix = "A:"
		}
		result += fmt.Sprintf("%s%s ", rolePrefix, message.Text)
	}
	return result
}

func (app *App) handleNewConversation(msg structs.OnebotGroupMessage, conversationID string, parentMessageID string, promotstr string) {
	// 使用预定义的时间戳作为会话标题
	conversationTitle := "2024-5-19/18:26" // 实际应用中应使用动态生成的时间戳
	userid := msg.UserID

	if config.GetGroupContext() == 2 && msg.MessageType != "private" {
		userid = msg.GroupID + msg.SelfID
	}

	// 添加用户记忆
	err := app.AddUserMemory(userid, conversationID, parentMessageID, conversationTitle)
	if err != nil {
		log.Printf("Error saving memory: %s", err)
		return
	}

	// 启动Go routine进行历史信息处理和标题更新
	go func() {
		userHistory, err := app.getHistory(conversationID, parentMessageID)
		if err != nil {
			log.Printf("Error retrieving history: %s", err)
			return
		}

		// 处理历史信息为特定格式的字符串
		memoryTitle := formatHistory(userHistory)
		newTitle := GetMemoryTitle(memoryTitle) // 获取最终的记忆标题

		// 更新记忆标题
		err = app.updateConversationTitle(userid, conversationID, parentMessageID, newTitle)
		if err != nil {
			log.Printf("Error updating conversation title: %s", err)
		}
	}()

	// 迁移用户到新的上下文
	app.migrateUserToNewContext(userid)

	// 获取并使用配置中指定的加载记忆指令
	loadCommand := config.GetMemoryLoadCommand()
	if len(loadCommand) > 0 {
		loadMemoryCommand := loadCommand[0] // 使用数组中的第一个指令
		saveMemoryResponse := fmt.Sprintf("旧的对话已经保存，可发送 %s 来查看，可以开始新的对话了！", loadMemoryCommand)
		app.sendMemoryResponse(msg, saveMemoryResponse, promotstr)
	}
}
