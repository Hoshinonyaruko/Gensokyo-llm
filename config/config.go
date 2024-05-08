package config

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/prompt"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
	"gopkg.in/yaml.v3"
)

var (
	instance *Config
	mu       sync.Mutex
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type Config struct {
	Version  int              `yaml:"version"`
	Settings structs.Settings `yaml:"settings"`
}

// LoadConfig 从文件中加载配置并初始化单例配置
func LoadConfig(path string) (*Config, error) {
	mu.Lock()
	defer mu.Unlock()

	// 如果单例已经被初始化了，直接返回
	if instance != nil {
		return instance, nil
	}

	configData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	err = yaml.Unmarshal(configData, conf)
	if err != nil {
		return nil, err
	}

	// 设置单例实例
	instance = conf
	return instance, nil
}

// 获取secretId
func GetsecretId() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.SecretId
	}
	return "0"
}

// 获取secretKey
func GetsecretKey() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.SecretKey
	}
	return "0"
}

// 获取region
func Getregion() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.Region
	}
	return "0"
}

// 获取useSse
func GetuseSse(options ...string) bool {
	mu.Lock()
	defer mu.Unlock()
	return getUseSseInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getUseSseInternal(options ...string) bool {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.UseSse
		}
		return false
	}

	// 使用传入的 basename
	basename := options[0]
	useSseInterface, err := prompt.GetSettingFromFilename(basename, "UseSse")
	if err != nil {
		log.Println("Error retrieving UseSse:", err)
		return getUseSseInternal() // 递归调用内部函数，不传递任何参数
	}

	useSse, ok := useSseInterface.(bool)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for UseSse, fetching default")
		return getUseSseInternal() // 递归调用内部函数，不传递任何参数
	}

	return useSse
}

// 获取GetPort
func GetPort() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.Port
	}
	return 46230
}

// 获取GetSelfPath
func GetSelfPath() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.SelfPath
	}
	return ""
}

// 获取getHttpPath
func GetHttpPath() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.HttpPath
	}
	return "0"
}

// 获取getLotus
func GetLotus(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getLotusInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getLotusInternal(options ...string) string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.Lotus
		}
		return ""
	}

	// 使用传入的 basename
	basename := options[0]
	lotusInterface, err := prompt.GetSettingFromFilename(basename, "Lotus")
	if err != nil {
		log.Println("Error retrieving Lotus:", err)
		return getLotusInternal() // 递归调用内部函数，不传递任何参数
	}

	lotus, ok := lotusInterface.(string)
	if !ok || lotus == "" { // 检查是否断言失败或结果为空字符串
		log.Println("Type assertion failed or empty string for Lotus, fetching default")
		return getLotusInternal() // 递归调用内部函数，不传递任何参数
	}

	return lotus
}

// 获取SystemPrompt
func SystemPrompt() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil && len(instance.Settings.SystemPrompt) > 0 {
		prompts := instance.Settings.SystemPrompt
		if len(prompts) == 1 {
			// 如果只有一个成员，直接返回
			return prompts[0]
		} else {
			selectedIndex := rand.Intn(len(prompts))
			selectedPrompt := prompts[selectedIndex]
			fmtf.Printf("Selected system prompt: %s\n", selectedPrompt) // 输出你返回的是哪个
			return selectedPrompt
		}
	}
	//如果是nil返回0代表不使用系统提示词
	return "0"
}

// 获取IPWhiteList
func IPWhiteList() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.IPWhiteList
	}
	return nil
}

// 获取最大上下文
func GetMaxTokensHunyuan(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getMaxTokensHunyuanInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getMaxTokensHunyuanInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.MaxTokensHunyuan
		}
		return 4096 // 默认值
	}

	// 使用传入的 basename
	basename := options[0]
	maxTokensHunyuanInterface, err := prompt.GetSettingFromFilename(basename, "MaxTokensHunyuan")
	if err != nil {
		log.Println("Error retrieving MaxTokensHunyuan:", err)
		return getMaxTokensHunyuanInternal() // 递归调用内部函数，不传递任何参数
	}

	maxTokensHunyuan, ok := maxTokensHunyuanInterface.(int)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for MaxTokensHunyuan, fetching default")
		return getMaxTokensHunyuanInternal() // 递归调用内部函数，不传递任何参数
	}

	return maxTokensHunyuan
}

// 获取Api类型
func GetApiType() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.ApiType
	}
	return 0
}

// 获取WenxinAccessToken
func GetWenxinAccessToken() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.WenxinAccessToken
	}
	return "0"
}

// 获取WenxinApiPath
func GetWenxinApiPath(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getWenxinApiPathInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getWenxinApiPathInternal(options ...string) string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.WenxinApiPath
		}
		return "0"
	}

	// 使用传入的 basename
	basename := options[0]
	apiPathInterface, err := prompt.GetSettingFromFilename(basename, "WenxinApiPath")
	if err != nil {
		log.Println("Error retrieving WenxinApiPath:", err)
		return getWenxinApiPathInternal() // 递归调用内部函数，不传递任何参数
	}

	apiPath, ok := apiPathInterface.(string)
	if !ok || apiPath == "" { // 检查是否断言失败或结果为空字符串
		log.Println("Type assertion failed or empty string for WenxinApiPath, fetching default")
		return getWenxinApiPathInternal() // 递归调用内部函数，不传递任何参数
	}

	return apiPath
}

// 获取GetMaxTokenWenxin
func GetMaxTokenWenxin() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.MaxTokenWenxin
	}
	return 0
}

// 获取GptModel
func GetGptModel(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getGptModelInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getGptModelInternal(options ...string) string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.GptModel
		}
		return "0"
	}

	// 使用传入的 basename
	basename := options[0]
	gptModelInterface, err := prompt.GetSettingFromFilename(basename, "GptModel")
	if err != nil {
		log.Println("Error retrieving GptModel:", err)
		return getGptModelInternal() // 递归调用内部函数，不传递任何参数
	}

	gptModel, ok := gptModelInterface.(string)
	if !ok || gptModel == "" { // 检查是否断言失败或结果为空字符串
		fmt.Println("Type assertion failed or empty string for GptModel, fetching default")
		return getGptModelInternal() // 递归调用内部函数，不传递任何参数
	}

	return gptModel
}

// 获取GptApiPath
func GetGptApiPath(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getGptApiPathInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getGptApiPathInternal(options ...string) string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.GptApiPath
		}
		return ""
	}

	// 使用传入的 basename
	basename := options[0]
	gptApiPathInterface, err := prompt.GetSettingFromFilename(basename, "GptApiPath")
	if err != nil {
		log.Println("Error retrieving GptApiPath:", err)
		return getGptApiPathInternal() // 递归调用内部函数，不传递任何参数
	}

	gptApiPath, ok := gptApiPathInterface.(string)
	if !ok || gptApiPath == "" { // 检查是否断言失败或结果为空字符串
		fmt.Println("Type assertion failed or empty string for GptApiPath, fetching default")
		return getGptApiPathInternal() // 递归调用内部函数，不传递任何参数
	}

	return gptApiPath
}

// 获取GptToken
func GetGptToken(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getGptTokenInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getGptTokenInternal(options ...string) string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.GptToken
		}
		return ""
	}

	// 使用传入的 basename
	basename := options[0]
	gptTokenInterface, err := prompt.GetSettingFromFilename(basename, "GptToken")
	if err != nil {
		log.Println("Error retrieving GptToken:", err)
		return getGptTokenInternal() // 递归调用内部函数，不传递任何参数
	}

	gptToken, ok := gptTokenInterface.(string)
	if !ok || gptToken == "" { // 检查是否断言失败或结果为空字符串
		fmt.Println("Type assertion failed or empty string for GptToken, fetching default")
		return getGptTokenInternal() // 递归调用内部函数，不传递任何参数
	}

	return gptToken
}

// 获取MaxTokenGpt
func GetMaxTokenGpt(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getMaxTokenGptInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getMaxTokenGptInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.MaxTokenGpt
		}
		return 0
	}

	// 使用传入的 basename
	basename := options[0]
	maxTokenGptInterface, err := prompt.GetSettingFromFilename(basename, "MaxTokenGpt")
	if err != nil {
		log.Println("Error retrieving MaxTokenGpt:", err)
		return getMaxTokenGptInternal() // 递归调用内部函数，不传递任何参数
	}

	maxTokenGpt, ok := maxTokenGptInterface.(int)
	if !ok { // 检查是否断言失败
		fmt.Println("Type assertion failed for MaxTokenGpt, fetching default")
		return getMaxTokenGptInternal() // 递归调用内部函数，不传递任何参数
	}

	return maxTokenGpt
}

// gpt安全模式
func GetGptSafeMode() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.GptSafeMode
	}
	return false
}

// UrlSendPics
func GetUrlSendPics() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.UrlSendPics
	}
	return false
}

// 获取GptSseType
func GetGptSseType() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.GptSseType
	}
	return 0
}

// 是否开启群信息
func GetGroupmessage() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.Groupmessage
	}
	return false
}

// 获取SplitByPuntuations
func GetSplitByPuntuations() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.SplitByPuntuations
	}
	return 0
}

// 获取SplitByPuntuationsGroup
func GetSplitByPuntuationsGroup() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.SplitByPuntuationsGroup
	}
	return 0
}

// 获取HunyuanType
func GetHunyuanType() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.HunyuanType
	}
	return 0
}

// 获取FirstQ
func GetFirstQ() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil && len(instance.Settings.FirstQ) > 0 {
		questions := instance.Settings.FirstQ
		if len(questions) == 1 {
			// 如果只有一个成员，直接返回
			return questions[0]
		} else {
			// 随机选择一个返回
			selectedIndex := rand.Intn(len(questions))
			selectedQuestion := questions[selectedIndex]
			fmtf.Printf("Selected first question: %s\n", selectedQuestion) // 输出你返回的是哪个问题
			return selectedQuestion
		}
	}
	// 如果是nil或者空数组，返回空字符串
	return ""
}

// 获取FirstA
func GetFirstA() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil && len(instance.Settings.FirstA) > 0 {
		answers := instance.Settings.FirstA
		if len(answers) == 1 {
			// 如果只有一个成员，直接返回
			return answers[0]
		} else {
			// 随机选择一个返回
			selectedIndex := rand.Intn(len(answers))
			selectedAnswer := answers[selectedIndex]
			fmtf.Printf("Selected first answer: %s\n", selectedAnswer) // 输出你返回的是哪个回答
			return selectedAnswer
		}
	}
	// 如果是nil或者空数组，返回空字符串
	return ""
}

// 获取SecondQ
func GetSecondQ() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil && len(instance.Settings.SecondQ) > 0 {
		questions := instance.Settings.SecondQ
		if len(questions) == 1 {
			return questions[0]
		} else {
			selectedIndex := rand.Intn(len(questions))
			selectedQuestion := questions[selectedIndex]
			fmtf.Printf("Selected second question: %s\n", selectedQuestion)
			return selectedQuestion
		}
	}
	return ""
}

// 获取SecondA
func GetSecondA() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil && len(instance.Settings.SecondA) > 0 {
		answers := instance.Settings.SecondA
		if len(answers) == 1 {
			return answers[0]
		} else {
			selectedIndex := rand.Intn(len(answers))
			selectedAnswer := answers[selectedIndex]
			fmtf.Printf("Selected second answer: %s\n", selectedAnswer)
			return selectedAnswer
		}
	}
	return ""
}

// 获取ThirdQ
func GetThirdQ() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil && len(instance.Settings.ThirdQ) > 0 {
		questions := instance.Settings.ThirdQ
		if len(questions) == 1 {
			return questions[0]
		} else {
			selectedIndex := rand.Intn(len(questions))
			selectedQuestion := questions[selectedIndex]
			fmtf.Printf("Selected third question: %s\n", selectedQuestion)
			return selectedQuestion
		}
	}
	return ""
}

// 获取ThirdA
func GetThirdA() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil && len(instance.Settings.ThirdA) > 0 {
		answers := instance.Settings.ThirdA
		if len(answers) == 1 {
			return answers[0]
		} else {
			selectedIndex := rand.Intn(len(answers))
			selectedAnswer := answers[selectedIndex]
			fmtf.Printf("Selected third answer: %s\n", selectedAnswer)
			return selectedAnswer
		}
	}
	return ""
}

// 获取DefaultChangeWord
func GetDefaultChangeWord() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.DefaultChangeWord
	}
	return "*"
}

// 是否SensitiveMode
func GetSensitiveMode() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.SensitiveMode
	}
	return false
}

// 获取SensitiveModeType
func GetSensitiveModeType() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.SensitiveModeType
	}
	return 0
}

// 获取AntiPromptAttackPath
func GetAntiPromptAttackPath() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.AntiPromptAttackPath
	}
	return ""
}

// 获取ReverseUserPrompt
func GetReverseUserPrompt() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.ReverseUserPrompt
	}
	return false
}

// 获取IgnoreExtraTips
func GetIgnoreExtraTips() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.IgnoreExtraTips
	}
	return false
}

// GetRandomSaveResponse 从SaveResponses数组中随机选择一个字符串返回
func GetRandomSaveResponse() string {
	mu.Lock()
	defer mu.Unlock()

	// 检查SaveResponses是否为空或nil
	if len(instance.Settings.SaveResponses) > 0 {
		if len(instance.Settings.SaveResponses) == 1 {
			// 如果只有一个元素，直接返回这个元素
			return instance.Settings.SaveResponses[0]
		} else {
			// 如果有多个元素，随机选择一个返回
			selectedIndex := rand.Intn(len(instance.Settings.SaveResponses))
			selectedResponse := instance.Settings.SaveResponses[selectedIndex]
			fmtf.Printf("Selected save response: %s\n", selectedResponse)
			return selectedResponse
		}
	}
	// 如果数组为空，返回空字符串
	return ""
}

// GetRestoreResponses 从RestoreResponses数组中随机选择一个字符串返回
func GetRandomRestoreResponses() string {
	mu.Lock()
	defer mu.Unlock()

	// 检查RestoreResponses是否为空或nil
	if len(instance.Settings.RestoreResponses) > 0 {
		if len(instance.Settings.RestoreResponses) == 1 {
			// 如果只有一个元素，直接返回这个元素
			return instance.Settings.RestoreResponses[0]
		} else {
			// 如果有多个元素，随机选择一个返回
			selectedIndex := rand.Intn(len(instance.Settings.RestoreResponses))
			selectedResponse := instance.Settings.RestoreResponses[selectedIndex]
			fmtf.Printf("Selected save response: %s\n", selectedResponse)
			return selectedResponse
		}
	}
	// 如果数组为空，返回空字符串
	return ""
}

// 获取RestoreCommand
func GetRestoreCommand() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RestoreCommand
	}
	return nil
}

// 获取UsePrivateSSE
func GetUsePrivateSSE() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.UsePrivateSSE
	}
	return false
}

// GetPromptkeyboard 获取Promptkeyboard，如果超过3个成员则随机选择3个
func GetPromptkeyboard() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil && len(instance.Settings.Promptkeyboard) > 0 {
		promptKeyboard := instance.Settings.Promptkeyboard
		if len(promptKeyboard) <= 3 {
			return promptKeyboard
		}

		// 如果数组成员超过3个，随机选择3个返回
		selected := make([]string, 3)
		indexesSelected := make(map[int]bool)
		for i := 0; i < 3; i++ {
			index := r.Intn(len(promptKeyboard))
			// 确保不重复选择
			for indexesSelected[index] {
				index = r.Intn(len(promptKeyboard))
			}
			indexesSelected[index] = true
			selected[i] = promptKeyboard[index]
		}
		return selected
	}
	return nil
}

// 获取Savelogs
func GetSavelogs() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.Savelogs
	}
	return false
}

// 获取AntiPromptLimit
func GetAntiPromptLimit() float64 {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.AntiPromptLimit
	}
	return 0.9
}

// 获取UseCache，增加可选参数支持动态配置查询
func GetUseCache(options ...string) bool {
	mu.Lock()
	defer mu.Unlock()
	return getUseCacheInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getUseCacheInternal(options ...string) bool {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.UseCache
		}
		return false
	}

	// 使用传入的 basename
	basename := options[0]
	useCacheInterface, err := prompt.GetSettingFromFilename(basename, "UseCache")
	if err != nil {
		log.Println("Error retrieving UseCache:", err)
		return getUseCacheInternal() // 如果出错，递归调用自身，不传递任何参数
	}

	useCache, ok := useCacheInterface.(bool)
	if !ok {
		log.Println("Type assertion failed for UseCache")
		return getUseCacheInternal() // 如果类型断言失败，递归调用自身，不传递任何参数
	}

	return useCache
}

// 获取CacheThreshold
func GetCacheThreshold() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.CacheThreshold
	}
	return 0
}

// 获取CacheChance
func GetCacheChance() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.CacheChance
	}
	return 0
}

// 获取EmbeddingType
func GetEmbeddingType() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.EmbeddingType
	}
	return 0
}

// 获取WenxinEmbeddingUrl
func GetWenxinEmbeddingUrl() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.WenxinEmbeddingUrl
	}
	return ""
}

// 获取GptEmbeddingUrl
func GetGptEmbeddingUrl() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.GptEmbeddingUrl
	}
	return ""
}

// 获取PrintHanming
func GetPrintHanming() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.PrintHanming
	}
	return false
}

// 获取CacheK
func GetCacheK() float64 {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.CacheK
	}
	return 0
}

// 获取CacheN
func GetCacheN() int64 {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.CacheN
	}
	return 0
}

// 获取PrintVector
func GetPrintVector() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.PrintVector
	}
	return false
}

// 获取VToBThreshold
func GetVToBThreshold() float64 {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.VToBThreshold
	}
	return 0
}

// 获取GptModeration
func GetGptModeration() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.GptModeration
	}
	return false
}

// 获取WenxinTopp
func GetWenxinTopp() float64 {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.WenxinTopp
	}
	return 0
}

// 获取WnxinPenaltyScore
func GetWnxinPenaltyScore() float64 {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.WnxinPenaltyScore
	}
	return 0
}

// 获取WenxinMaxOutputTokens
func GetWenxinMaxOutputTokens() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.WenxinMaxOutputTokens
	}
	return 0
}

// 获取VectorSensitiveFilter
func GetVectorSensitiveFilter() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.VectorSensitiveFilter
	}
	return false
}

// 获取VertorSensitiveThreshold
func GetVertorSensitiveThreshold() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.VertorSensitiveThreshold
	}
	return 0
}

// GetAllowedLanguages 返回允许的语言列表
func GetAllowedLanguages() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.AllowedLanguages
	}
	return nil // 或返回一个默认的语言列表
}

// GetLanguagesResponseMessages 返回语言拦截响应消息列表
func GetLanguagesResponseMessages() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil && len(instance.Settings.LanguagesResponseMessages) > 0 {
		// 如果列表中只有一个消息，直接返回这个消息
		if len(instance.Settings.LanguagesResponseMessages) == 1 {
			return instance.Settings.LanguagesResponseMessages[0]
		}
		// 如果有多个消息，随机选择一个返回
		index := rand.Intn(len(instance.Settings.LanguagesResponseMessages))
		return instance.Settings.LanguagesResponseMessages[index]
	}
	return "" // 如果列表为空，返回空字符串
}

// 获取QuestionMaxLenth
func GetQuestionMaxLenth() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.QuestionMaxLenth
	}
	return 0
}

// GetQmlResponseMessages 返回语言拦截响应消息列表
func GetQmlResponseMessages() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil && len(instance.Settings.QmlResponseMessages) > 0 {
		// 如果列表中只有一个消息，直接返回这个消息
		if len(instance.Settings.QmlResponseMessages) == 1 {
			return instance.Settings.QmlResponseMessages[0]
		}
		// 如果有多个消息，随机选择一个返回
		index := rand.Intn(len(instance.Settings.QmlResponseMessages))
		return instance.Settings.QmlResponseMessages[index]
	}
	return "" // 如果列表为空，返回空字符串
}

// BlacklistResponseMessages 返回语言拦截响应消息列表
func GetBlacklistResponseMessages() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil && len(instance.Settings.BlacklistResponseMessages) > 0 {
		// 如果列表中只有一个消息，直接返回这个消息
		if len(instance.Settings.BlacklistResponseMessages) == 1 {
			return instance.Settings.BlacklistResponseMessages[0]
		}
		// 如果有多个消息，随机选择一个返回
		index := rand.Intn(len(instance.Settings.BlacklistResponseMessages))
		return instance.Settings.BlacklistResponseMessages[index]
	}
	return "" // 如果列表为空，返回空字符串
}

// 获取NoContext
func GetNoContext() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.NoContext
	}
	return false
}

// 获取WithdrawCommand
func GetWithdrawCommand() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.WithdrawCommand
	}
	return nil
}

// 获取FunctionMode
func GetFunctionMode() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.FunctionMode
	}
	return false
}

// 获取FunctionPath
func GetFunctionPath() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.FunctionPath
	}
	return ""
}

// 获取UseFunctionPromptkeyboard
func GetUseFunctionPromptkeyboard() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.UseFunctionPromptkeyboard
	}
	return false
}

// 获取UseAIPromptkeyboard
func GetUseAIPromptkeyboard() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.UseAIPromptkeyboard
	}
	return false
}

// 获取AIPromptkeyboardPath
func GetAIPromptkeyboardPath(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getAIPromptkeyboardPathInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getAIPromptkeyboardPathInternal(options ...string) string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.AIPromptkeyboardPath
		}
		return ""
	}

	// 使用传入的 basename
	basename := options[0]
	pathInterface, err := prompt.GetSettingFromFilename(basename, "AIPromptkeyboardPath")
	if err != nil {
		log.Println("Error retrieving AIPromptkeyboardPath:", err)
		return getAIPromptkeyboardPathInternal() // 递归调用内部函数，不传递任何参数
	}

	path, ok := pathInterface.(string)
	if !ok || path == "" { // 检查是否断言失败或结果为空字符串
		log.Println("Type assertion failed or empty string for AIPromptkeyboardPath, fetching default")
		return getAIPromptkeyboardPathInternal() // 递归调用内部函数，不传递任何参数
	}

	return path
}

// 获取RWKV API路径
func GetRwkvApiPath() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvApiPath
	}
	return ""
}

// 获取RWKV最大令牌数
func GetRwkvMaxTokens(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getRwkvMaxTokensInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getRwkvMaxTokensInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.RwkvMaxTokens
		}
		return 0
	}

	// 使用传入的 basename
	basename := options[0]
	maxTokensInterface, err := prompt.GetSettingFromFilename(basename, "RwkvMaxTokens")
	if err != nil {
		log.Println("Error retrieving RwkvMaxTokens:", err)
		return getRwkvMaxTokensInternal() // 递归调用内部函数，不传递任何参数
	}

	maxTokens, ok := maxTokensInterface.(int)
	if !ok { // 检查是否断言失败
		fmt.Println("Type assertion failed for RwkvMaxTokens, fetching default")
		return getRwkvMaxTokensInternal() // 递归调用内部函数，不传递任何参数
	}

	return maxTokens
}

// 获取RwkvSseType
func GetRwkvSseType() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvSseType
	}
	return 0
}

// 获取RWKV温度
func GetRwkvTemperature() float64 {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvTemperature
	}
	return 0.0
}

// 获取RWKV Top P
func GetRwkvTopP() float64 {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvTopP
	}
	return 0.0
}

// 获取RWKV存在惩罚
func GetRwkvPresencePenalty() float64 {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvPresencePenalty
	}
	return 0.0
}

// 获取RWKV频率惩罚
func GetRwkvFrequencyPenalty() float64 {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvFrequencyPenalty
	}
	return 0.0
}

// 获取RWKV惩罚衰减
func GetRwkvPenaltyDecay() float64 {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvPenaltyDecay
	}
	return 0.0
}

// 获取RWKV Top K
func GetRwkvTopK() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvTopK
	}
	return 0
}

// 获取RWKV是否全局惩罚
func GetRwkvGlobalPenalty() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvGlobalPenalty
	}
	return false
}

// 获取RWKV是否流模式
func GetRwkvStream() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvStream
	}
	return false
}

// 获取RWKV停止列表
func GetRwkvStop() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvStop
	}
	return nil
}

// 获取RWKV用户名
func GetRwkvUserName() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvUserName
	}
	return ""
}

// 获取RWKV助手名
func GetRwkvAssistantName() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvAssistantName
	}
	return ""
}

// 获取RWKV系统名称
func GetRwkvSystemName() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvSystemName
	}
	return ""
}

// 获取RWKV是否预处理
func GetRwkvPreSystem() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.RwkvPreSystem
	}
	return false
}

// 获取隐藏日志
func GetHideExtraLogs() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.HideExtraLogs
	}
	return false
}

// 获取wsServerToken
func GetWSServerToken() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.WSServerToken
	}
	return ""
}

// 获取PathToken
func GetPathToken() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.PathToken
	}
	return ""
}

// 获取开启全部api
func GetAllApi() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.AllApi
	}
	return false
}

// 获取Proxy
func GetProxy(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getProxyInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getProxyInternal(options ...string) string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.Proxy
		}
		return "" // 提供一个默认的 Proxy 值
	}

	// 使用传入的 basename
	basename := options[0]
	proxyInterface, err := prompt.GetSettingFromFilename(basename, "Proxy")
	if err != nil {
		log.Println("Error retrieving Proxy:", err)
		return getProxyInternal() // 递归调用内部函数，不传递任何参数
	}

	proxy, ok := proxyInterface.(string)
	if !ok || proxy == "" { // 检查是否断言失败或结果为空字符串
		fmt.Println("Type assertion failed or empty string for Proxy, fetching default")
		return getProxyInternal() // 递归调用内部函数，不传递任何参数
	}

	return proxy
}

// 获取 StandardGptApi
func GetStandardGptApi(options ...string) bool {
	mu.Lock()
	defer mu.Unlock()
	return getStandardGptApiInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getStandardGptApiInternal(options ...string) bool {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.StandardGptApi
		}
		return false
	}

	// 使用传入的 basename
	basename := options[0]
	standardGptApiInterface, err := prompt.GetSettingFromFilename(basename, "StandardGptApi")
	if err != nil {
		log.Println("Error retrieving StandardGptApi:", err)
		return getStandardGptApiInternal() // 递归调用内部函数，不传递任何参数
	}

	standardGptApi, ok := standardGptApiInterface.(bool)
	if !ok { // 检查是否断言失败
		fmt.Println("Type assertion failed for StandardGptApi, fetching default")
		return getStandardGptApiInternal() // 递归调用内部函数，不传递任何参数
	}

	return standardGptApi
}

// 获取 PromptMarkType
func GetPromptMarkType(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getPromptMarkTypeInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getPromptMarkTypeInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.PromptMarkType
		}
		return 0 // 默认返回 0 或一个合理的默认值
	}

	// 使用传入的 basename
	basename := options[0]
	promptMarkTypeInterface, err := prompt.GetSettingFromFilename(basename, "PromptMarkType")
	if err != nil {
		log.Println("Error retrieving PromptMarkType:", err)
		return getPromptMarkTypeInternal() // 递归调用内部函数，不传递任何参数
	}

	promptMarkType, ok := promptMarkTypeInterface.(int)
	if !ok { // 检查是否断言失败
		fmt.Println("Type assertion failed for PromptMarkType, fetching default")
		return getPromptMarkTypeInternal() // 递归调用内部函数，不传递任何参数
	}

	return promptMarkType
}

// 获取 PromptMarksLength
func GetPromptMarksLength(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getPromptMarksLengthInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getPromptMarksLengthInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.PromptMarksLength
		}
		return 0 // 默认返回 0 或一个合理的默认值
	}

	// 使用传入的 basename
	basename := options[0]
	promptMarksLengthInterface, err := prompt.GetSettingFromFilename(basename, "PromptMarksLength")
	if err != nil {
		log.Println("Error retrieving PromptMarksLength:", err)
		return getPromptMarksLengthInternal() // 递归调用内部函数，不传递任何参数
	}

	promptMarksLength, ok := promptMarksLengthInterface.(int)
	if !ok { // 检查是否断言失败
		fmt.Println("Type assertion failed for PromptMarksLength, fetching default")
		return getPromptMarksLengthInternal() // 递归调用内部函数，不传递任何参数
	}

	return promptMarksLength
}

// 获取 PromptMarks
func GetPromptMarks(options ...string) []string {
	mu.Lock()
	defer mu.Unlock()
	return getPromptMarksInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getPromptMarksInternal(options ...string) []string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.PromptMarks
		}
		return nil // 如果实例或设置未定义，返回nil
	}

	// 使用传入的 basename
	basename := options[0]
	promptMarksInterface, err := prompt.GetSettingFromFilename(basename, "PromptMarks")
	if err != nil {
		log.Println("Error retrieving PromptMarks:", err)
		return getPromptMarksInternal() // 递归调用内部函数，不传递任何参数
	}

	promptMarks, ok := promptMarksInterface.([]string)
	if !ok { // 检查是否断言失败
		fmt.Println("Type assertion failed for PromptMarks, fetching default")
		return getPromptMarksInternal() // 递归调用内部函数，不传递任何参数
	}

	return promptMarks
}

// 获取 EnhancedQA
func GetEnhancedQA(options ...string) bool {
	mu.Lock()
	defer mu.Unlock()
	return getEnhancedQAInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getEnhancedQAInternal(options ...string) bool {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.EnhancedQA
		}
		return false
	}

	// 使用传入的 basename
	basename := options[0]
	enhancedQAInterface, err := prompt.GetSettingFromFilename(basename, "EnhancedQA")
	if err != nil {
		log.Println("Error retrieving EnhancedQA:", err)
		return getEnhancedQAInternal() // 递归调用内部函数，不传递任何参数
	}

	enhancedQA, ok := enhancedQAInterface.(bool)
	if !ok { // 检查是否断言失败
		fmt.Println("Type assertion failed for EnhancedQA, fetching default")
		return getEnhancedQAInternal() // 递归调用内部函数，不传递任何参数
	}

	return enhancedQA
}

// 获取 PromptChoicesQ
func GetPromptChoicesQ(options ...string) []string {
	mu.Lock()
	defer mu.Unlock()
	return getPromptChoicesQInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getPromptChoicesQInternal(options ...string) []string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.PromptChoicesQ
		}
		return nil // 如果实例或设置未定义，返回nil
	}

	// 使用传入的 basename
	basename := options[0]
	promptChoicesInterface, err := prompt.GetSettingFromFilename(basename, "PromptChoicesQ")
	if err != nil {
		log.Println("Error retrieving PromptChoicesQ:", err)
		return getPromptChoicesQInternal() // 递归调用内部函数，不传递任何参数
	}

	promptChoices, ok := promptChoicesInterface.([]string)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for PromptChoicesQ, fetching default")
		return getPromptChoicesQInternal() // 递归调用内部函数，不传递任何参数
	}

	return promptChoices
}

// 获取 PromptChoicesA
func GetPromptChoicesA(options ...string) []string {
	mu.Lock()
	defer mu.Unlock()
	return getPromptChoicesAInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getPromptChoicesAInternal(options ...string) []string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.PromptChoicesA
		}
		return nil // 如果实例或设置未定义，返回nil
	}

	// 使用传入的 basename
	basename := options[0]
	promptChoicesInterface, err := prompt.GetSettingFromFilename(basename, "PromptChoicesA")
	if err != nil {
		log.Println("Error retrieving PromptChoicesA:", err)
		return getPromptChoicesAInternal() // 递归调用内部函数，不传递任何参数
	}

	promptChoices, ok := promptChoicesInterface.([]string)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for PromptChoicesA, fetching default")
		return getPromptChoicesAInternal() // 递归调用内部函数，不传递任何参数
	}

	return promptChoices
}

// 获取enhancedpromptChoices
func GetEnhancedPromptChoices(options ...string) bool {
	mu.Lock()
	defer mu.Unlock()
	return getEnhancedPromptChoicesInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getEnhancedPromptChoicesInternal(options ...string) bool {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.EnhancedPromptChoices
		}
		return false // 如果实例或设置未定义，返回默认值false
	}

	// 使用传入的 basename
	basename := options[0]
	enhancedPromptChoicesInterface, err := prompt.GetSettingFromFilename(basename, "EnhancedPromptChoices")
	if err != nil {
		log.Println("Error retrieving EnhancedPromptChoices:", err)
		return getEnhancedPromptChoicesInternal() // 递归调用内部函数，不传递任何参数
	}

	enhancedPromptChoices, ok := enhancedPromptChoicesInterface.(bool)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for EnhancedPromptChoices, fetching default")
		return getEnhancedPromptChoicesInternal() // 递归调用内部函数，不传递任何参数
	}

	return enhancedPromptChoices
}

// 获取switchOnQ
func GetSwitchOnQ(options ...string) []string {
	mu.Lock()
	defer mu.Unlock()
	return getSwitchOnQInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getSwitchOnQInternal(options ...string) []string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.SwitchOnQ
		}
		return nil // 默认值为空数组
	}

	// 使用传入的 basename
	basename := options[0]
	switchOnQInterface, err := prompt.GetSettingFromFilename(basename, "SwitchOnQ")
	if err != nil {
		log.Println("Error retrieving SwitchOnQ:", err)
		return getSwitchOnQInternal() // 递归调用内部函数，不传递任何参数
	}

	switchOnQ, ok := switchOnQInterface.([]string)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for SwitchOnQ, fetching default")
		return getSwitchOnQInternal() // 递归调用内部函数，不传递任何参数
	}

	return switchOnQ
}

// 获取switchOnA
func GetSwitchOnA(options ...string) []string {
	mu.Lock()
	defer mu.Unlock()
	return getSwitchOnAInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getSwitchOnAInternal(options ...string) []string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.SwitchOnA
		}
		return nil // 默认值为空数组
	}

	// 使用传入的 basename
	basename := options[0]
	switchOnAInterface, err := prompt.GetSettingFromFilename(basename, "SwitchOnA")
	if err != nil {
		log.Println("Error retrieving SwitchOnA:", err)
		return getSwitchOnAInternal() // 递归调用内部函数，不传递任何参数
	}

	switchOnA, ok := switchOnAInterface.([]string)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for SwitchOnA, fetching default")
		return getSwitchOnAInternal() // 递归调用内部函数，不传递任何参数
	}

	return switchOnA
}

// 获取ExitOnQ
func GetExitOnQ(options ...string) []string {
	mu.Lock()
	defer mu.Unlock()
	return getExitOnQInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getExitOnQInternal(options ...string) []string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.ExitOnQ
		}
		return nil // 默认值为空数组
	}

	// 使用传入的 basename
	basename := options[0]
	exitOnQInterface, err := prompt.GetSettingFromFilename(basename, "ExitOnQ")
	if err != nil {
		log.Println("Error retrieving ExitOnQ:", err)
		return getExitOnQInternal() // 递归调用内部函数，不传递任何参数
	}

	exitOnQ, ok := exitOnQInterface.([]string)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for ExitOnQ, fetching default")
		return getExitOnQInternal() // 递归调用内部函数，不传递任何参数
	}

	return exitOnQ
}

// 获取ExitOnA
func GetExitOnA(options ...string) []string {
	mu.Lock()
	defer mu.Unlock()
	return getExitOnAInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getExitOnAInternal(options ...string) []string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.ExitOnA
		}
		return nil // 默认值为空数组
	}

	// 使用传入的 basename
	basename := options[0]
	exitOnAInterface, err := prompt.GetSettingFromFilename(basename, "ExitOnA")
	if err != nil {
		log.Println("Error retrieving ExitOnA:", err)
		return getExitOnAInternal() // 递归调用内部函数，不传递任何参数
	}

	exitOnA, ok := exitOnAInterface.([]string)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for ExitOnA, fetching default")
		return getExitOnAInternal() // 递归调用内部函数，不传递任何参数
	}

	return exitOnA
}

// 获取EnvType
func GetEnvType(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getEnvTypeInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getEnvTypeInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.EnvType
		}
		return 0 // 如果实例或设置未定义，返回默认值0
	}

	// 使用传入的 basename
	basename := options[0]
	envTypeInterface, err := prompt.GetSettingFromFilename(basename, "EnvType")
	if err != nil {
		log.Println("Error retrieving EnvType:", err)
		return getEnvTypeInternal() // 递归调用内部函数，不传递任何参数
	}

	envType, ok := envTypeInterface.(int)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for EnvType, fetching default")
		return getEnvTypeInternal() // 递归调用内部函数，不传递任何参数
	}

	return envType
}

// 获取 PromptCoverQ
func GetPromptCoverQ(options ...string) []string {
	mu.Lock()
	defer mu.Unlock()
	return getPromptCoverQInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getPromptCoverQInternal(options ...string) []string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.PromptCoverQ
		}
		return nil // 如果实例或设置未定义，返回nil
	}

	// 使用传入的 basename
	basename := options[0]
	promptCoverInterface, err := prompt.GetSettingFromFilename(basename, "PromptCoverQ")
	if err != nil {
		log.Println("Error retrieving PromptCoverQ:", err)
		return getPromptCoverQInternal() // 递归调用内部函数，不传递任何参数
	}

	promptCover, ok := promptCoverInterface.([]string)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for PromptCoverQ, fetching default")
		return getPromptCoverQInternal() // 递归调用内部函数，不传递任何参数
	}

	return promptCover
}

// 获取 PromptCoverA
func GetPromptCoverA(options ...string) []string {
	mu.Lock()
	defer mu.Unlock()
	return getPromptCoverAInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getPromptCoverAInternal(options ...string) []string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.PromptCoverA
		}
		return nil // 如果实例或设置未定义，返回nil
	}

	// 使用传入的 basename
	basename := options[0]
	promptCoverInterface, err := prompt.GetSettingFromFilename(basename, "PromptCoverA")
	if err != nil {
		log.Println("Error retrieving PromptCoverA:", err)
		return getPromptCoverAInternal() // 递归调用内部函数，不传递任何参数
	}

	promptCover, ok := promptCoverInterface.([]string)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for PromptCoverA, fetching default")
		return getPromptCoverAInternal() // 递归调用内部函数，不传递任何参数
	}

	return promptCover
}

// 获取 EnvPics
func GetEnvPics(options ...string) []string {
	mu.Lock()
	defer mu.Unlock()
	return getEnvPicsInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getEnvPicsInternal(options ...string) []string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.EnvPics
		}
		return nil // 如果实例或设置未定义，返回nil
	}

	// 使用传入的 basename
	basename := options[0]
	envPicsInterface, err := prompt.GetSettingFromFilename(basename, "EnvPics")
	if err != nil {
		log.Println("Error retrieving EnvPics:", err)
		return getEnvPicsInternal() // 递归调用内部函数，不传递任何参数
	}

	envPics, ok := envPicsInterface.([]string)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for EnvPics, fetching default")
		return getEnvPicsInternal() // 递归调用内部函数，不传递任何参数
	}

	return envPics
}

// 获取 EnvContents
func GetEnvContents(options ...string) []string {
	mu.Lock()
	defer mu.Unlock()
	return getEnvContentsInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getEnvContentsInternal(options ...string) []string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.EnvContents
		}
		return nil // 如果实例或设置未定义，返回nil
	}

	// 使用传入的 basename
	basename := options[0]
	envContentsInterface, err := prompt.GetSettingFromFilename(basename, "EnvContents")
	if err != nil {
		log.Println("Error retrieving EnvContents:", err)
		return getEnvContentsInternal() // 递归调用内部函数，不传递任何参数
	}

	envContents, ok := envContentsInterface.([]string)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for EnvContents, fetching default")
		return getEnvContentsInternal() // 递归调用内部函数，不传递任何参数
	}

	return envContents
}

// 群内md气泡
func GetMdPromptKeyboardAtGroup() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.MdPromptKeyboardAtGroup
	}
	return false
}

// 第四个气泡
func GetNo4Promptkeyboard() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.No4Promptkeyboard
	}
	return false
}
