package applogic

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
	"github.com/hoshinonyaruko/gensokyo-llm/utils"
)

// ResponseDataEnv 用于解析外层响应
type ResponseDataEnv struct {
	ConversationID string `json:"conversationId"`
	MessageID      string `json:"messageId"`
	Response       string `json:"response"` // 这里是可能包含文件链接的llm结果
}

// 分开发而且不使用sse
func GetAndSendEnv(msg string, promptstr string, message structs.OnebotGroupMessage, selfid string) {
	baseurl := config.GetAIPromptkeyboardPath(promptstr)
	fmtf.Printf("获取到keyboard baseurl:%v", baseurl)
	// 使用net/url包来构建和编码URL
	urlParams := url.Values{}
	if promptstr != "" {
		urlParams.Add("prompt", promptstr+"-env")
	} else {
		urlParams.Add("prompt", "env")
	}

	// 将查询参数编码后附加到基本URL上
	fullURL := baseurl
	if len(urlParams) > 0 {
		fullURL += "?" + urlParams.Encode()
	}

	fmtf.Printf("Generated PromptKeyboard URL:%v\n", fullURL)

	requestBody, err := json.Marshal(map[string]interface{}{
		"message":         msg,
		"conversationId":  "",
		"parentMessageId": "",
		"user_id":         "",
	})

	if err != nil {
		fmt.Printf("Error marshalling request: %v\n", err)
		return
	}

	resp, err := http.Post(fullURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}
	fmt.Printf("Response: %s\n", string(responseBody))

	var responseData ResponseDataEnv
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		fmt.Printf("Error unmarshalling response data: %v\n", err)
		return
	}

	// 处理图片
	processResponseData(&responseData)

	// 判断消息类型，如果是私人消息或私有群消息，发送私人消息；否则，根据配置决定是否发送群消息
	if message.RealMessageType == "group_private" || message.MessageType == "private" {
		utils.SendPrivateMessage(message.UserID, responseData.Response, selfid)
	} else {
		utils.SendGroupMessage(message.GroupID, message.UserID, responseData.Response, selfid)
	}
}

func processResponseData(responseData *ResponseDataEnv) {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}

	// 正则表达式匹配多种图片格式
	re := regexp.MustCompile(`\[(?:图片|pic|背景|image):(.*?)\]`)
	matches := re.FindAllStringSubmatch(responseData.Response, -1)

	for _, match := range matches {
		imagePath := match[1]
		fullPath := imagePath

		// 如果不是绝对路径，则认为是相对路径
		if !filepath.IsAbs(imagePath) {
			fullPath = filepath.Join(currentDir, imagePath)
		}

		// 判断文件是否位于当前程序目录内
		if !strings.HasPrefix(fullPath, currentDir) {
			fmt.Println("Image path is outside of the current directory")
			return
		}

		// 检查文件后缀
		if ext := filepath.Ext(imagePath); ext != ".jpg" && ext != ".png" && ext != ".gif" {
			fmt.Println("Unsupported image file type")
			return
		}

		// 读取图片文件
		imageData, err := os.ReadFile(fullPath)
		if err != nil {
			fmt.Printf("Error reading image file: %v\n", err)
			return
		}

		// 转换为base64
		base64Str := base64.StdEncoding.EncodeToString(imageData)
		cqImageTag := fmt.Sprintf("[CQ:image,file=base64://%s]", base64Str)

		// 替换原始标签
		responseData.Response = strings.Replace(responseData.Response, match[0], cqImageTag, 1)
	}
}
