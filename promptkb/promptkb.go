package promptkb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
)

// ResponseDataPromptKeyboard 用于解析外层响应
type ResponseDataPromptKeyboard struct {
	ConversationID string `json:"conversationId"`
	MessageID      string `json:"messageId"`
	Response       string `json:"response"` // 这里是嵌套的JSON字符串
}

// 你要扮演一个json生成器,根据我下一句提交的QA内容,推断我可能会继续问的问题,生成json数组格式的结果,如:输入Q我好累啊A要休息一下吗,返回["嗯，我想要休息","我想喝杯咖啡","你平时怎么休息呢"]，返回需要是["","",""]需要2-3个结果
func GetPromptKeyboardAI(msg string, promptstr string) []string {
	baseurl := config.GetAIPromptkeyboardPath(promptstr)
	fmtf.Printf("获取到keyboard baseurl:%v", baseurl)
	var keyboardprompt string
	// 使用net/url包来构建和编码URL
	urlParams := url.Values{}
	if promptstr != "" {
		urlParams.Add("prompt", promptstr+"-keyboard")
		keyboardprompt = promptstr + "-keyboard"
	} else {
		urlParams.Add("prompt", "keyboard")
		keyboardprompt = "keyboard"
	}

	// 将查询参数编码后附加到基本URL上
	fullURL := baseurl
	if len(urlParams) > 0 {
		fullURL += "?" + urlParams.Encode()
	}

	fmtf.Printf("Generated PromptKeyboard URL:%v\n", fullURL)

	// 按提示词区分的细化替换 这里主要不是为了安全和敏感词,而是细化效果,也就没有使用acnode提高效率
	msg = ReplaceTextIn(msg, keyboardprompt)

	requestBody, err := json.Marshal(map[string]interface{}{
		"message":         msg,
		"conversationId":  "",
		"parentMessageId": "",
		"user_id":         "",
	})

	if err != nil {
		fmtf.Printf("Error marshalling request: %v\n", err)
		return config.GetPromptkeyboard()
	}

	resp, err := http.Post(fullURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmtf.Printf("Error sending request: %v\n", err)
		return config.GetPromptkeyboard()
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmtf.Printf("Error reading response body: %v\n", err)
		return config.GetPromptkeyboard()
	}

	fmtf.Printf("气泡Response: %s\n", string(responseBody))

	// 使用正则表达式替换不正确的转义字符
	re := regexp.MustCompile(`\\+`)
	responseBody = re.ReplaceAll(responseBody, []byte(`\`))
	// 使用正则表达式移除所有换行符和前后的空白字符
	re = regexp.MustCompile(`\s*\n\s*`)
	responseBody = re.ReplaceAll(responseBody, []byte(""))

	// 移除 CRLF
	responseBody = bytes.ReplaceAll(responseBody, []byte{13, 10}, []byte(""))

	// 移除 92 110 32 32
	responseBody = bytes.ReplaceAll(responseBody, []byte{92, 110, 32, 32}, []byte(""))

	// 移除 92 110
	responseBody = bytes.ReplaceAll(responseBody, []byte{92, 110}, []byte(""))

	// 移除单独的 CR 和 LF
	responseBody = bytes.ReplaceAll(responseBody, []byte{10}, []byte("")) // LF
	responseBody = bytes.ReplaceAll(responseBody, []byte{13}, []byte("")) // CR

	fmtf.Printf("清洗后气泡Response: %v\n", responseBody)
	//fmtf.Printf("清洗后气泡Response: %s\n", string(responseBody))

	var responseData ResponseDataPromptKeyboard
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		fmtf.Printf("Error unmarshalling response data: %v\n", err)
		return config.GetPromptkeyboard()
	}

	var keyboardPrompts []string
	// 预处理响应数据，移除可能的换行符
	preprocessedResponse := strings.TrimSpace(responseData.Response)

	// 假设 originalContent 是包含错误转义的原始字符串
	preprocessedResponse = strings.ReplaceAll(preprocessedResponse, `\\`, `\`)
	preprocessedResponse = strings.ReplaceAll(preprocessedResponse, `\\`, `\`)
	preprocessedResponse = strings.ReplaceAll(preprocessedResponse, `\\`, `\`)

	// 预处理响应数据，移除非JSON内容
	preprocessedResponse = preprocessResponse(preprocessedResponse)

	fmtf.Printf("准备解析气泡: %v\n", preprocessedResponse)

	// 尝试直接解析JSON
	err = json.Unmarshal([]byte(preprocessedResponse), &keyboardPrompts)
	if err != nil {
		fmt.Printf("Error unmarshalling nested response: %v\n", err)
		return config.GetPromptkeyboard()
	}

	// 检查keyboardPrompts数量是否足够
	if len(keyboardPrompts) < 3 {
		promptkeyboard := config.GetPromptkeyboard()
		randomIndex := rand.Intn(len(promptkeyboard))
		keyboardPrompts = append(keyboardPrompts, promptkeyboard[randomIndex])
	}

	// 应用 ReplaceTextOut 替换规则
	keyboardPrompts = ApplyReplaceTextOutToKeyboardPrompts(keyboardPrompts, keyboardprompt)

	return keyboardPrompts
}

func ApplyReplaceTextOutToKeyboardPrompts(keyboardPrompts []string, promptstr string) []string {
	for i, prompt := range keyboardPrompts {
		keyboardPrompts[i] = ReplaceTextOut(prompt, promptstr)
	}
	return keyboardPrompts
}

func preprocessResponse(response string) string {
	// 移除转义符
	cleanResponse := strings.ReplaceAll(response, "\\n", "\n")

	// 正则表达式匹配有效的JSON数组
	re := regexp.MustCompile(`\[\s*".*?"\s*\]`)
	matches := re.FindAllString(cleanResponse, -1)

	// 选择最佳匹配（偏好更多的元素）
	bestMatch := ""
	for _, match := range matches {
		if len(match) > len(bestMatch) {
			bestMatch = match
		}
	}

	return bestMatch
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
