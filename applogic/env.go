package applogic

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/server"
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
func (app *App) GetAndSendEnv(msg string, promptstr string, message structs.OnebotGroupMessage, selfid string, PromptStrStat int, PromptLength int) {
	var responseData ResponseDataEnv
	EnvContents := config.GetEnvContents(promptstr + "-env")
	//如果没有人工写的EnvContents,使用ai生成,速度慢的感人,还影响对话效果和气泡keyboard
	if len(EnvContents) == 0 {
		// 故事模式规则 应用 PromptCoverQ
		app.ApplyPromptCoverQ(promptstr+"-env", &msg, &message)
		// 生成后续场景 暂时共用keyboard baseurl
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

		if err := json.Unmarshal(responseBody, &responseData); err != nil {
			fmt.Printf("Error unmarshalling response data: %v\n", err)
			return
		}
	} else {
		responseData.Response = processSelection(PromptStrStat, PromptLength, promptstr)

		// 打印或处理 responseData
		fmt.Println("最终env响应:", responseData.Response)
	}

	// 处理图片
	newResponse := processResponseData(responseData)

	// 当本轮文字是1: 图片也是 1: 时
	if newResponse == "" {
		fmt.Println("最终env响应为空")
		return
	}

	// 判断消息类型，如果是私人消息或私有群消息，发送私人消息；否则，根据配置决定是否发送群消息
	if message.RealMessageType == "group_private" || message.MessageType == "private" {
		utils.SendPrivateMessageRaw(message.UserID, newResponse, selfid)
	} else {
		utils.SendGroupMessage(message.GroupID, message.UserID, newResponse, selfid, promptstr)
	}
}

func selectBasedOnSenceId(envItems []string, senceId int) (selectedItems []string) {
	prefix := fmt.Sprintf("%d:", senceId)
	filtered := make([]string, 0)
	defaults := make([]string, 0)

	// 预编译正则表达式以匹配任何以数字和冒号开头的字符串
	re, err := regexp.Compile(`^\d+:`)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
		return envItems // 在编译正则表达式出错时，返回原始数组
	}

	for _, item := range envItems {
		if strings.HasPrefix(item, prefix) {
			filtered = append(filtered, strings.TrimPrefix(item, prefix))
		} else if !re.MatchString(item) {
			defaults = append(defaults, item)
		}
	}

	if len(filtered) > 0 {
		return filtered
	} else if len(defaults) > 0 {
		return defaults
	}
	return envItems
}

func stripPrefix(s string) string {
	colonIndex := strings.Index(s, ":")
	if colonIndex != -1 && colonIndex < len(s)-1 && allDigits(s[:colonIndex]) {
		return s[colonIndex+1:]
	}
	return s
}

func allDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func getRandomItem(items []string) string {
	if len(items) == 0 {
		return "默认内容"
	}
	index := rand.Intn(len(items))
	return items[index]
}

func processSelection(promptStrStat int, promptLength int, promptStr string) string {
	senceId := promptLength - promptStrStat
	fmtf.Printf("processSelection senceId:%v", senceId)

	envContents := config.GetEnvContents(fmt.Sprintf("%s-env", promptStr))
	selectedContent := stripPrefix(getRandomItem(selectBasedOnSenceId(envContents, senceId)))

	envPics := config.GetEnvPics(fmt.Sprintf("%s-env", promptStr))
	selectedPic := stripPrefix(getRandomItem(selectBasedOnSenceId(envPics, senceId)))

	if selectedPic != "默认内容" {
		return fmt.Sprintf("%s %s", selectedContent, selectedPic) // 添加空格分隔内容和图片链接
	}
	return selectedContent
}

// processResponseData 处理响应数据并返回处理后的字符串
func processResponseData(responseData ResponseDataEnv) string {
	// 判断是否使用HTTP图片地址
	if config.GetUrlSendPics() {
		return processWithHttpImages(responseData)
	}
	return processWithLocalImages(responseData)
}

// processWithLocalImages 处理本地图片转为CQ码
func processWithLocalImages(responseData ResponseDataEnv) string {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return responseData.Response
	}
	fmt.Println("Current directory:", currentDir)

	// 使用更加宽泛的正则表达式匹配图片标签
	re := regexp.MustCompile(`\[(?:图片|pic|背景|image):([^\]]+)\]`)
	matches := re.FindAllStringSubmatch(responseData.Response, -1)

	if len(matches) == 0 {
		fmt.Println("No image tags found in response")
		return responseData.Response // 如果没有找到匹配项，返回原始响应
	}

	// 创建一个新的响应字符串
	newResponse := responseData.Response

	for _, match := range matches {
		imagePath := match[1]
		fullPath := imagePath

		if !filepath.IsAbs(imagePath) {
			fullPath = filepath.Join(currentDir, imagePath)
		}

		fmt.Println("Full image path:", fullPath)

		if !strings.HasPrefix(fullPath, currentDir) {
			fmt.Println("Image path is outside of the current directory")
			continue // 跳过这个图片，处理下一个
		}

		ext := filepath.Ext(imagePath)
		if ext != ".jpg" && ext != ".png" && ext != ".gif" {
			fmt.Println("Unsupported image file type:", ext)
			continue // 跳过这个图片，处理下一个
		}

		imageData, err := os.ReadFile(fullPath)
		if err != nil {
			fmt.Printf("Error reading image file: %v\n", err)
			continue // 跳过这个图片，处理下一个
		}

		base64Str := base64.StdEncoding.EncodeToString(imageData)
		cqImageTag := fmt.Sprintf("[CQ:image,file=base64://%s]", base64Str)
		newResponse = strings.Replace(newResponse, match[0], cqImageTag, 1)
	}

	return newResponse
}

// processWithHttpImages 上传图片并使用HTTP地址转为CQ码
func processWithHttpImages(responseData ResponseDataEnv) string {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return responseData.Response
	}
	fmt.Println("Current directory:", currentDir)

	// 使用更加宽泛的正则表达式匹配图片标签
	re := regexp.MustCompile(`\[(?:图片|pic|背景|image):([^\]]+)\]`)
	matches := re.FindAllStringSubmatch(responseData.Response, -1)
	if len(matches) == 0 {
		fmt.Println("No image tags found in response")
		return responseData.Response // 如果没有找到匹配项，返回原始响应
	}

	newResponse := responseData.Response
	for _, match := range matches {
		imagePath := match[1]
		fullPath := imagePath

		if !filepath.IsAbs(imagePath) {
			fullPath = filepath.Join(currentDir, imagePath)
		}

		fmt.Println("Full image path:", fullPath)

		if !strings.HasPrefix(fullPath, currentDir) {
			fmt.Println("Image path is outside of the current directory")
			continue // 跳过这个图片，处理下一个
		}

		ext := filepath.Ext(imagePath)
		if ext != ".jpg" && ext != ".png" && ext != ".gif" {
			fmt.Println("Unsupported image file type:", ext)
			continue // 跳过这个图片，处理下一个
		}

		imageData, err := os.ReadFile(fullPath)
		if err != nil {
			fmt.Printf("Error reading image file: %v\n", err)
			continue // 跳过这个图片，处理下一个
		}

		base64Str := base64.StdEncoding.EncodeToString(imageData)
		httpImageUrl, err := server.OriginalUploadBehavior(base64Str)
		if err != nil {
			continue
		}

		cqImageTag := fmt.Sprintf("[CQ:image,file=%s]", httpImageUrl)
		newResponse = strings.Replace(newResponse, match[0], cqImageTag, 1)
	}

	return newResponse
}
