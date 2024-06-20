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

func LoadConfig(path string) (*Config, error) {
	mu.Lock()
	defer mu.Unlock()

	conf, err := loadConfigFromFile(path)
	if err != nil {
		return nil, err
	}

	instance = conf
	return instance, nil
}

func loadConfigFromFile(path string) (*Config, error) {
	configData, err := os.ReadFile(path)
	if err != nil {
		log.Println("Failed to read file:", err)
		return nil, err
	}

	conf := &Config{}
	if err := yaml.Unmarshal(configData, conf); err != nil {
		log.Printf("failed to unmarshal YAML[%v]:%v", path, err)
		return nil, err
	}

	log.Printf("成功加载配置文件 %s\n", path)
	return conf, nil
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
func GetuseSse(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getUseSseInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getUseSseInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.UseSse
		}
		return 0
	}

	// 使用传入的 basename
	basename := options[0]
	useSseInterface, err := prompt.GetSettingFromFilename(basename, "UseSse")
	if err != nil {
		log.Println("Error retrieving UseSse:", err)
		return getUseSseInternal() // 递归调用内部函数，不传递任何参数
	}

	useSse, ok := useSseInterface.(int)
	if !ok || useSse == 0 { // 检查是否断言失败 或者是0
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

// 获取HttpPaths
func GetHttpPaths() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.HttpPaths
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

	if maxTokensHunyuan == 0 {
		return getMaxTokensHunyuanInternal()
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

	if apiPath == "" {
		return getWenxinApiPathInternal()
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

	if gptModel == "" {
		return getGptModelInternal()
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

	if gptToken == "" {
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

	if maxTokenGpt == 0 {
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
func GetSplitByPuntuations(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getSplitByPuntuationsInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getSplitByPuntuationsInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.SplitByPuntuations
		}
		return 0 // 默认值或错误处理
	}

	// 使用传入的 basename 来查找特定配置
	basename := options[0]
	SplitByPuntuationsInterface, err := prompt.GetSettingFromFilename(basename, "SplitByPuntuations")
	if err != nil {
		log.Println("Error retrieving SplitByPuntuations:", err)
		return getSplitByPuntuationsInternal() // 递归调用内部函数，不传递任何参数
	}

	SplitByPuntuations, ok := SplitByPuntuationsInterface.(int)
	if !ok { // 检查类型断言是否失败
		fmt.Println("Type assertion failed for SplitByPuntuations, fetching default")
		return getSplitByPuntuationsInternal() // 递归调用内部函数，不传递任何参数
	}

	if SplitByPuntuations == 0 {
		return getSplitByPuntuationsInternal() // 递归调用内部函数，不传递任何参数
	}

	return SplitByPuntuations
}

// 获取GetSplitByPuntuationsGroup
func GetSplitByPuntuationsGroup(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getSplitByPuntuationsGroupInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getSplitByPuntuationsGroupInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.SplitByPuntuationsGroup
		}
		return 0 // 默认值或错误处理
	}

	// 使用传入的 basename 来查找特定配置
	basename := options[0]
	SplitByPuntuationsGroupInterface, err := prompt.GetSettingFromFilename(basename, "SplitByPuntuationsGroup")
	if err != nil {
		log.Println("Error retrieving SplitByPuntuationsGroup:", err)
		return getSplitByPuntuationsGroupInternal() // 递归调用内部函数，不传递任何参数
	}

	SplitByPuntuationsGroup, ok := SplitByPuntuationsGroupInterface.(int)
	if !ok { // 检查类型断言是否失败
		fmt.Println("Type assertion failed for SplitByPuntuationsGroup, fetching default")
		return getSplitByPuntuationsGroupInternal() // 递归调用内部函数，不传递任何参数
	}

	if SplitByPuntuationsGroup == 0 {
		return getSplitByPuntuationsGroupInternal() // 递归调用内部函数，不传递任何参数
	}

	return SplitByPuntuationsGroup
}

// 获取GroupHintChance
func GetGroupHintChance(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getGroupHintChanceInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getGroupHintChanceInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.GroupHintChance
		}
		return 0 // 默认值或错误处理
	}

	// 使用传入的 basename 来查找特定配置
	basename := options[0]
	GroupHintChanceInterface, err := prompt.GetSettingFromFilename(basename, "GroupHintChance")
	if err != nil {
		log.Println("Error retrieving GroupHintChance:", err)
		return getGroupHintChanceInternal() // 递归调用内部函数，不传递任何参数
	}

	GroupHintChance, ok := GroupHintChanceInterface.(int)
	if !ok { // 检查类型断言是否失败
		fmt.Println("Type assertion failed for GroupHintChance, fetching default")
		return getGroupHintChanceInternal() // 递归调用内部函数，不传递任何参数
	}

	if GroupHintChance == 0 {
		return getGroupHintChanceInternal() // 递归调用内部函数，不传递任何参数
	}

	return GroupHintChance
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
func GetUseCache(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getUseCacheInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getUseCacheInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.UseCache
		}
		return 0
	}

	// 使用传入的 basename
	basename := options[0]
	useCacheInterface, err := prompt.GetSettingFromFilename(basename, "UseCache")
	if err != nil {
		log.Println("Error retrieving UseCache:", err)
		return getUseCacheInternal() // 如果出错，递归调用自身，不传递任何参数
	}

	useCache, ok := useCacheInterface.(int)
	if !ok || useCache == 0 {
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

// 获取GroupContext
func GetGroupContext(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getGroupContextInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getGroupContextInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.GroupContext
		}
		return 0
	}

	// 使用传入的 basename
	basename := options[0]
	GroupContextInterface, err := prompt.GetSettingFromFilename(basename, "GroupContext")
	if err != nil {
		log.Println("Error retrieving GroupContext:", err)
		return getGroupContextInternal() // 递归调用内部函数，不传递任何参数
	}

	GroupContext, ok := GroupContextInterface.(int)
	if !ok || GroupContext == 0 { // 检查是否断言失败或结果为空字符串
		log.Println("Type assertion failed or empty string for GroupContext, fetching default")
		return getGroupContextInternal() // 递归调用内部函数，不传递任何参数
	}

	return GroupContext
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

// 获取MemoryCommand
func GetMemoryCommand() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.MemoryCommand
	}
	return nil
}

// 获取MemoryLoadCommand
func GetMemoryLoadCommand() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.MemoryLoadCommand
	}
	return nil
}

// 获取NewConversationCommand
func GetNewConversationCommand() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.NewConversationCommand
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

// 获取MemoryListMD
func GetMemoryListMD() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.MemoryListMD
	}
	return 0
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

	if maxTokens == 0 {
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
func GetPromptMarks(options ...string) []structs.BranchConfig {
	mu.Lock()
	defer mu.Unlock()
	return getPromptMarksInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getPromptMarksInternal(options ...string) []structs.BranchConfig {
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

	promptMarks, ok := promptMarksInterface.([]structs.BranchConfig)
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
func GetPromptChoicesQ(options ...string) []structs.PromptChoice {
	mu.Lock()
	defer mu.Unlock()
	return getPromptChoicesQInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getPromptChoicesQInternal(options ...string) []structs.PromptChoice {
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

	promptChoices, ok := promptChoicesInterface.([]structs.PromptChoice)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for PromptChoicesQ, fetching default")
		return getPromptChoicesQInternal() // 递归调用内部函数，不传递任何参数
	}

	return promptChoices
}

// 获取 PromptChanceQ
func GetPromptChanceQ(options ...string) []structs.PromptChance {
	mu.Lock()
	defer mu.Unlock()
	return getPromptChanceQInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getPromptChanceQInternal(options ...string) []structs.PromptChance {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.PromptChanceQ
		}
		return nil // 如果实例或设置未定义，返回nil
	}

	// 使用传入的 basename
	basename := options[0]
	promptChanceInterface, err := prompt.GetSettingFromFilename(basename, "PromptChanceQ")
	if err != nil {
		log.Println("Error retrieving PromptChanceQ:", err)
		return getPromptChanceQInternal() // 递归调用内部函数，不传递任何参数
	}

	promptChances, ok := promptChanceInterface.([]structs.PromptChance)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for PromptChanceQ, fetching default")
		return getPromptChanceQInternal() // 递归调用内部函数，不传递任何参数
	}

	return promptChances
}

// 获取 PromptChoicesA
func GetPromptChoicesA(options ...string) []structs.PromptChoice {
	mu.Lock()
	defer mu.Unlock()
	return getPromptChoicesAInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getPromptChoicesAInternal(options ...string) []structs.PromptChoice {
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

	promptChoices, ok := promptChoicesInterface.([]structs.PromptChoice)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for PromptChoicesA, fetching default")
		return getPromptChoicesAInternal() // 递归调用内部函数，不传递任何参数
	}

	return promptChoices
}

// 获取switchOnQ
func GetSwitchOnQ(options ...string) []structs.PromptSwitch {
	mu.Lock()
	defer mu.Unlock()
	return getSwitchOnQInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getSwitchOnQInternal(options ...string) []structs.PromptSwitch {
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

	switchOnQ, ok := switchOnQInterface.([]structs.PromptSwitch)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for SwitchOnQ, fetching default")
		return getSwitchOnQInternal() // 递归调用内部函数，不传递任何参数
	}

	return switchOnQ
}

// 获取switchOnA
func GetSwitchOnA(options ...string) []structs.PromptSwitch {
	mu.Lock()
	defer mu.Unlock()
	return getSwitchOnAInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getSwitchOnAInternal(options ...string) []structs.PromptSwitch {
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

	switchOnA, ok := switchOnAInterface.([]structs.PromptSwitch)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for SwitchOnA, fetching default")
		return getSwitchOnAInternal() // 递归调用内部函数，不传递任何参数
	}

	return switchOnA
}

// 获取ExitOnQ
func GetExitOnQ(options ...string) []structs.PromptExit {
	mu.Lock()
	defer mu.Unlock()
	return getExitOnQInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getExitOnQInternal(options ...string) []structs.PromptExit {
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

	exitOnQ, ok := exitOnQInterface.([]structs.PromptExit)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for ExitOnQ, fetching default")
		return getExitOnQInternal() // 递归调用内部函数，不传递任何参数
	}

	return exitOnQ
}

// 获取ExitOnA
func GetExitOnA(options ...string) []structs.PromptExit {
	mu.Lock()
	defer mu.Unlock()
	return getExitOnAInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getExitOnAInternal(options ...string) []structs.PromptExit {
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

	exitOnA, ok := exitOnAInterface.([]structs.PromptExit)
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
func GetPromptCoverQ(options ...string) []structs.PromptChoice {
	mu.Lock()
	defer mu.Unlock()
	return getPromptCoverQInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getPromptCoverQInternal(options ...string) []structs.PromptChoice {
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

	promptCover, ok := promptCoverInterface.([]structs.PromptChoice)
	if !ok { // 检查是否断言失败
		log.Println("Type assertion failed for PromptCoverQ, fetching default")
		return getPromptCoverQInternal() // 递归调用内部函数，不传递任何参数
	}

	return promptCover
}

// 获取 PromptCoverA
func GetPromptCoverA(options ...string) []structs.PromptChoice {
	mu.Lock()
	defer mu.Unlock()
	return getPromptCoverAInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getPromptCoverAInternal(options ...string) []structs.PromptChoice {
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

	promptCover, ok := promptCoverInterface.([]structs.PromptChoice)
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

// GetTyqwApiPath 获取TYQW API路径，可接受basename作为参数
func GetTyqwApiPath(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getTyqwApiPathInternal(options...)
}

// getTyqwApiPathInternal 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getTyqwApiPathInternal(options ...string) string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.TyqwApiPath
		}
		return "" // 默认值或错误处理
	}

	// 使用传入的 basename
	basename := options[0]
	apiPathInterface, err := prompt.GetSettingFromFilename(basename, "TyqwApiPath")
	if err != nil {
		log.Println("Error retrieving TyqwApiPath:", err)
		return getTyqwApiPathInternal() // 递归调用内部函数，不传递任何参数
	}

	apiPath, ok := apiPathInterface.(string)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for TyqwApiPath, fetching default")
		return getTyqwApiPathInternal() // 递归调用内部函数，不传递任何参数
	}

	if apiPath == "" {
		return getTyqwApiPathInternal()
	}

	return apiPath
}

// 获取TYQW最大Token数量，可接受basename作为参数
func GetTyqwMaxTokens(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getTyqwMaxTokensInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getTyqwMaxTokensInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.TyqwMaxTokens
		}
		return 0 // 默认值或错误处理
	}

	// 使用传入的 basename
	basename := options[0]
	maxTokensInterface, err := prompt.GetSettingFromFilename(basename, "TyqwMaxTokens")
	if err != nil {
		log.Println("Error retrieving TyqwMaxTokens:", err)
		return getTyqwMaxTokensInternal() // 递归调用内部函数，不传递任何参数
	}

	maxTokens, ok := maxTokensInterface.(int)
	if !ok { // 检查类型断言是否失败
		fmt.Println("Type assertion failed for TyqwMaxTokens, fetching default")
		return getTyqwMaxTokensInternal() // 递归调用内部函数，不传递任何参数
	}

	if maxTokens == 0 {
		return getTyqwMaxTokensInternal()
	}

	return maxTokens
}

// GetTyqwTemperature 获取TYQW温度设置，可接受basename作为参数
func GetTyqwTemperature(options ...string) float64 {
	mu.Lock()
	defer mu.Unlock()
	return getTyqwTemperatureInternal(options...)
}

// getTyqwTemperatureInternal 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getTyqwTemperatureInternal(options ...string) float64 {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.TyqwTemperature
		}
		return 0.0 // 默认值或错误处理
	}

	// 使用传入的 basename
	basename := options[0]
	temperatureInterface, err := prompt.GetSettingFromFilename(basename, "TyqwTemperature")
	if err != nil {
		log.Println("Error retrieving TyqwTemperature:", err)
		return getTyqwTemperatureInternal() // 递归调用内部函数，不传递任何参数
	}

	temperature, ok := temperatureInterface.(float64)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for TyqwTemperature, fetching default")
		return getTyqwTemperatureInternal() // 递归调用内部函数，不传递任何参数
	}

	if temperature == 0 {
		return getTyqwTemperatureInternal()
	}

	return temperature
}

// GetTyqwTopP 获取TYQW Top P值，可接受basename作为参数
func GetTyqwTopP(options ...string) float64 {
	mu.Lock()
	defer mu.Unlock()
	return getTyqwTopPInternal(options...)
}

// getTyqwTopPInternal 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getTyqwTopPInternal(options ...string) float64 {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.TyqwTopP
		}
		return 0.0 // 默认值或错误处理
	}

	// 使用传入的 basename
	basename := options[0]
	topPInterface, err := prompt.GetSettingFromFilename(basename, "TyqwTopP")
	if err != nil {
		log.Println("Error retrieving TyqwTopP:", err)
		return getTyqwTopPInternal() // 递归调用内部函数，不传递任何参数
	}

	topP, ok := topPInterface.(float64)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for TyqwTopP, fetching default")
		return getTyqwTopPInternal() // 递归调用内部函数，不传递任何参数
	}

	if topP == 0 {
		return getTyqwTopPInternal()
	}

	return topP
}

// GetTyqwTopK 获取TYQW Top K设置，可接受basename作为参数
func GetTyqwTopK(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getTyqwTopKInternal(options...)
}

// getTyqwTopKInternal 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getTyqwTopKInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.TyqwTopK
		}
		return 0 // 默认值或错误处理
	}

	// 使用传入的 basename
	basename := options[0]
	topKInterface, err := prompt.GetSettingFromFilename(basename, "TyqwTopK")
	if err != nil {
		log.Println("Error retrieving TyqwTopK:", err)
		return getTyqwTopKInternal() // 递归调用内部函数，不传递任何参数
	}

	topK, ok := topKInterface.(int)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for TyqwTopK, fetching default")
		return getTyqwTopKInternal() // 递归调用内部函数，不传递任何参数
	}

	if topK == 0 {
		return getTyqwTopKInternal()
	}

	return topK
}

// GetTyqwSseType 获取TYQW SSE类型，可接受basename作为参数
func GetTyqwSseType(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getTyqwSseTypeInternal(options...)
}

// getTyqwSseTypeInternal 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getTyqwSseTypeInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.TyqwSseType
		}
		return 0 // 默认值或错误处理
	}

	// 使用传入的 basename
	basename := options[0]
	sseTypeInterface, err := prompt.GetSettingFromFilename(basename, "TyqwSseType")
	if err != nil {
		log.Println("Error retrieving TyqwSseType:", err)
		return getTyqwSseTypeInternal() // 递归调用内部函数，不传递任何参数
	}

	sseType, ok := sseTypeInterface.(int)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for TyqwSseType, fetching default")
		return getTyqwSseTypeInternal() // 递归调用内部函数，不传递任何参数
	}

	if sseType == 0 {
		return getTyqwSseTypeInternal()
	}

	return sseType
}

// 获取TYQW用户名
func GetTyqwUserName() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.TyqwUserName
	}
	return "" // 默认值或错误处理
}

// 获取TYQW助手名称
func GetTyqwAssistantName() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.TyqwAssistantName
	}
	return "" // 默认值或错误处理
}

// 获取TYQW系统名称
func GetTyqwSystemName() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.TyqwSystemName
	}
	return "" // 默认值或错误处理
}

// 获取TYQW是否在系统层面进行预处理
func GetTyqwPreSystem() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.TyqwPreSystem
	}
	return false // 默认值或错误处理
}

// 获取TYQW重复度惩罚因子
func GetTyqwRepetitionPenalty() float64 {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.TyqwRepetitionPenalty
	}
	return 1.0 // 默认值或错误处理
}

// 获取TYQW停止标记
func GetTyqwStopTokens() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.TyqwStop
	}
	return nil // 默认值或错误处理
}

// 获取TYQW随机数种子
func GetTyqwSeed() int64 {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.TyqwSeed
	}
	return 0 // 默认值或错误处理
}

// 获取TYQW是否启用互联网搜索
func GetTyqwEnableSearch() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.TyqwEnableSearch
	}
	return false // 默认值或错误处理
}

// GetTyqwModel 获取TYQW模型名称，可接受basename作为参数
func GetTyqwModel(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getTyqwModelInternal(options...)
}

// getTyqwModelInternal 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getTyqwModelInternal(options ...string) string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.TyqwModel
		}
		return "default-model" // 默认值或错误处理
	}

	// 使用传入的 basename
	basename := options[0]
	modelInterface, err := prompt.GetSettingFromFilename(basename, "TyqwModel")
	if err != nil {
		log.Println("Error retrieving TyqwModel:", err)
		return getTyqwModelInternal() // 递归调用内部函数，不传递任何参数
	}

	model, ok := modelInterface.(string)
	if !ok || model == "" { // 检查类型断言是否失败或结果为空字符串
		log.Println("Type assertion failed or empty string for TyqwModel, fetching default")
		return getTyqwModelInternal() // 递归调用内部函数，不传递任何参数
	}

	if model == "" {
		return getTyqwModelInternal()
	}

	return model
}

// GetTyqwKey 获取TYQW API Key，可接受basename作为参数
func GetTyqwKey(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getTyqwKeyInternal(options...)
}

// getTyqwKeyInternal 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getTyqwKeyInternal(options ...string) string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.TyqwApiKey
		}
		return "" // 默认值或错误处理，表示没有找到有效的API Key
	}

	// 使用传入的 basename
	basename := options[0]
	apiKeyInterface, err := prompt.GetSettingFromFilename(basename, "TyqwApiKey")
	if err != nil {
		log.Println("Error retrieving TyqwApiKey:", err)
		return getTyqwKeyInternal() // 递归调用内部函数，不传递任何参数
	}

	apiKey, ok := apiKeyInterface.(string)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for TyqwApiKey, fetching default")
		return getTyqwKeyInternal() // 递归调用内部函数，不传递任何参数
	}

	if apiKey == "" {
		return getTyqwKeyInternal()
	}

	return apiKey
}

// 获取TYQW Workspace
func GetTyqworkspace() (string, error) {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		if instance.Settings.TyqwWorkspace == "" {
			return "", fmt.Errorf("workspace is not configured") // 错误处理，当workspace未配置时
		}
		return instance.Settings.TyqwWorkspace, nil
	}
	return "", fmt.Errorf("configuration instance is not initialized") // 错误处理，当配置实例未初始化时
}

// GetGlmApiPath 获取GLM API路径，可接受basename作为参数
func GetGlmApiPath(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getGlmApiPathInternal(options...)
}

// getGlmApiPathInternal 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getGlmApiPathInternal(options ...string) string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.GlmApiPath
		}
		return "" // 默认值或错误处理
	}

	// 使用传入的 basename
	basename := options[0]
	apiPathInterface, err := prompt.GetSettingFromFilename(basename, "GlmApiPath")
	if err != nil {
		log.Println("Error retrieving GlmApiPath:", err)
		return getGlmApiPathInternal() // 递归调用内部函数，不传递任何参数
	}

	apiPath, ok := apiPathInterface.(string)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for GlmApiPath, fetching default")
		return getGlmApiPathInternal() // 递归调用内部函数，不传递任何参数
	}

	if apiPath == "" {
		return getGlmApiPathInternal() // 递归调用内部函数，不传递任何参数
	}

	return apiPath
}

// GetGlmModel 获取模型编码，可接受basename作为参数
func GetGlmModel(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getGlmModelInternal(options...)
}

// getGlmModelInternal 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getGlmModelInternal(options ...string) string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.GlmModel
		}
		return "" // 默认值或错误处理
	}

	// 使用传入的 basename
	basename := options[0]
	modelInterface, err := prompt.GetSettingFromFilename(basename, "GlmModel")
	if err != nil {
		log.Println("Error retrieving GlmModel:", err)
		return getGlmModelInternal() // 递归调用内部函数，不传递任何参数
	}

	model, ok := modelInterface.(string)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for GlmModel, fetching default")
		return getGlmModelInternal() // 递归调用内部函数，不传递任何参数
	}

	if model == "" {
		return getGlmModelInternal() // 递归调用内部函数，不传递任何参数
	}

	return model
}

// GetGlmApiKey 获取glm密钥，可接受basename作为参数
func GetGlmApiKey(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getGlmApiKeyInternal(options...)
}

// getGlmApiKeyInternal 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getGlmApiKeyInternal(options ...string) string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.GlmApiKey
		}
		return "" // 默认值或错误处理
	}

	// 使用传入的 basename
	basename := options[0]
	apiKeyInterface, err := prompt.GetSettingFromFilename(basename, "GlmApiKey")
	if err != nil {
		log.Println("Error retrieving GlmApiKey:", err)
		return getGlmApiKeyInternal() // 递归调用内部函数，不传递任何参数
	}

	apiKey, ok := apiKeyInterface.(string)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for GlmApiKey, fetching default")
		return getGlmApiKeyInternal() // 递归调用内部函数，不传递任何参数
	}

	if apiKey == "" {
		return getGlmApiKeyInternal() // 递归调用内部函数，不传递任何参数
	}

	return apiKey
}

// GetGlmMaxTokens 获取模型输出的最大tokens数，可接受basename作为参数
func GetGlmMaxTokens(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getGlmMaxTokensInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getGlmMaxTokensInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.GlmMaxTokens
		}
		return 1024 // 默认值或错误处理
	}

	// 使用传入的 basename 来查找特定配置
	basename := options[0]
	maxTokensInterface, err := prompt.GetSettingFromFilename(basename, "GlmMaxTokens")
	if err != nil {
		log.Println("Error retrieving GlmMaxTokens:", err)
		return getGlmMaxTokensInternal() // 递归调用内部函数，不传递任何参数
	}

	maxTokens, ok := maxTokensInterface.(int)
	if !ok { // 检查类型断言是否失败
		fmt.Println("Type assertion failed for GlmMaxTokens, fetching default")
		return getGlmMaxTokensInternal() // 递归调用内部函数，不传递任何参数
	}

	if maxTokens == 0 {
		return getGlmMaxTokensInternal() // 递归调用内部函数，不传递任何参数
	}

	return maxTokens
}

// GetGlmTemperature 获取模型的采样温度，可接受basename作为参数
func GetGlmTemperature(options ...string) float64 {
	mu.Lock()
	defer mu.Unlock()
	return getGlmTemperatureInternal(options...)
}

// getGlmTemperatureInternal 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getGlmTemperatureInternal(options ...string) float64 {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.GlmTemperature
		}
		return 0.95 // 默认值或错误处理
	}

	// 使用传入的 basename
	basename := options[0]
	temperatureInterface, err := prompt.GetSettingFromFilename(basename, "GlmTemperature")
	if err != nil {
		log.Println("Error retrieving GlmTemperature:", err)
		return getGlmTemperatureInternal() // 递归调用内部函数，不传递任何参数
	}

	temperature, ok := temperatureInterface.(float64)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for GlmTemperature, fetching default")
		return getGlmTemperatureInternal() // 递归调用内部函数，不传递任何参数
	}

	if temperature == 0 {
		return getGlmTemperatureInternal() // 递归调用内部函数，不传递任何参数
	}

	return temperature
}

// GetGlmDoSample 获取是否启用采样策略
func GetGlmDoSample() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.GlmDoSample
	}
	return true // 返回默认值
}

// GetGlmToolChoice 获取工具选择策略
func GetGlmToolChoice() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.GlmToolChoice
	}
	return "auto" // 返回默认值
}

// GetGlmUserID 获取终端用户的唯一ID
func GetGlmUserID() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.GlmUserID
	}
	return "" // 如果没有配置则返回空字符串
}

// GetGlmRequestID 获取请求的唯一标识
func GetGlmRequestID() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.GlmRequestID
	}
	return "" // 返回默认值，表示没有设置
}

// GetGlmTopP 获取核取样概率，可接受basename作为参数
func GetGlmTopP(options ...string) float64 {
	mu.Lock()
	defer mu.Unlock()
	return getGlmTopPInternal(options...)
}

// getGlmTopPInternal 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getGlmTopPInternal(options ...string) float64 {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.GlmTopP
		}
		return 0.7 // 默认值或错误处理
	}

	// 使用传入的 basename
	basename := options[0]
	topPInterface, err := prompt.GetSettingFromFilename(basename, "GlmTopP")
	if err != nil {
		log.Println("Error retrieving GlmTopP:", err)
		return getGlmTopPInternal() // 递归调用内部函数，不传递任何参数
	}

	topP, ok := topPInterface.(float64)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for GlmTopP, fetching default")
		return getGlmTopPInternal() // 递归调用内部函数，不传递任何参数
	}

	if topP == 0 {
		return getGlmTopPInternal() // 递归调用内部函数，不传递任何参数
	}

	return topP
}

// GetGlmStop 获取停止生成的词列表
func GetGlmStop() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.GlmStop
	}
	return nil // 返回空切片，表示没有设置停止词
}

// GetGlmTools 获取可调用的工具列表
func GetGlmTools() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.GlmTools
	}
	return []string{} // 返回空切片，表示没有工具设置
}

// GetGroupHintWords 获取GroupHintWords列表，可接受basename作为参数
func GetGroupHintWords(options ...string) []string {
	mu.Lock()
	defer mu.Unlock()
	return getGroupHintWordsInternal(options...)
}

// getGroupHintWordsInternal 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getGroupHintWordsInternal(options ...string) []string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.GroupHintWords
		}
		return nil // 默认值或错误处理
	}

	// 使用传入的 basename
	basename := options[0]
	hintWordsInterface, err := prompt.GetSettingFromFilename(basename, "GroupHintWords")
	if err != nil {
		log.Println("Error retrieving GroupHintWords:", err)
		return getGroupHintWordsInternal() // 递归调用内部函数，不传递任何参数
	}

	hintWords, ok := hintWordsInterface.([]string)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for GroupHintWords, fetching default")
		return getGroupHintWordsInternal() // 递归调用内部函数，不传递任何参数
	}

	return hintWords
}

// 获取HunyuanStreamModeration值
func GetHunyuanStreamModeration(options ...string) bool {
	mu.Lock()
	defer mu.Unlock()
	return getHunyuanStreamModerationInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getHunyuanStreamModerationInternal(options ...string) bool {
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.HunyuanStreamModeration
		}
		return false
	}

	basename := options[0]
	valueInterface, err := prompt.GetSettingFromFilename(basename, "HunyuanStreamModeration")
	if err != nil {
		log.Println("Error retrieving HunyuanStreamModeration:", err)
		return getHunyuanStreamModerationInternal()
	}

	value, ok := valueInterface.(bool)
	if !ok || !value {
		log.Println("Fetching default HunyuanStreamModeration")
		return getHunyuanStreamModerationInternal()
	}

	return value
}

// 获取TopPHunyuan值
func GetTopPHunyuan(options ...string) float64 {
	mu.Lock()
	defer mu.Unlock()
	return getTopPHunyuanInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getTopPHunyuanInternal(options ...string) float64 {
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.TopPHunyuan
		}
		return 0.0
	}

	basename := options[0]
	valueInterface, err := prompt.GetSettingFromFilename(basename, "TopPHunyuan")
	if err != nil {
		log.Println("Error retrieving TopPHunyuan:", err)
		return getTopPHunyuanInternal()
	}

	value, ok := valueInterface.(float64)
	if !ok || value == 0.0 {
		log.Println("Fetching default TopPHunyuan")
		return getTopPHunyuanInternal()
	}

	return value
}

// 获取TemperatureHunyuan值
func GetTemperatureHunyuan(options ...string) float64 {
	mu.Lock()
	defer mu.Unlock()
	return getTemperatureHunyuanInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getTemperatureHunyuanInternal(options ...string) float64 {
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.TemperatureHunyuan
		}
		return 0.0
	}

	basename := options[0]
	valueInterface, err := prompt.GetSettingFromFilename(basename, "TemperatureHunyuan")
	if err != nil {
		log.Println("Error retrieving TemperatureHunyuan:", err)
		return getTemperatureHunyuanInternal()
	}

	value, ok := valueInterface.(float64)
	if !ok || value == 0.0 {
		log.Println("Fetching default TemperatureHunyuan")
		return getTemperatureHunyuanInternal()
	}

	return value
}

// GetYuanqiConf return conf.YuanqiAssistantID, conf.YuanqiToken
func GetYuanqiConf(options ...string) (string, string) {
	mu.Lock()
	defer mu.Unlock()
	return getYuanqiConfInternal(options...)
}

// getYuanqiConfInternal 内部递归函数，处理配置获取逻辑
func getYuanqiConfInternal(options ...string) (string, string) {
	if len(options) == 0 || options[0] == "" {
		// 从instance中读取配置数组
		if instance != nil && len(instance.Settings.Yuanqiconfs) > 0 {
			// 从instance全局变量中随机选择一个配置
			index := rand.Intn(len(instance.Settings.Yuanqiconfs))
			conf := instance.Settings.Yuanqiconfs[index]
			return conf.YuanqiAssistantID, conf.YuanqiToken
		} else {
			return "", ""
		}
	}

	// 使用prompt包从指定的文件名中获取配置
	basename := options[0]
	confInterface, err := prompt.GetSettingFromFilename(basename, "Yuanqiconfs")
	if err != nil {
		log.Printf("Error retrieving settings from file: %s, error: %v", basename, err)
		return getYuanqiConfInternal() // 递归调用内部函数，不传递任何参数
	}

	// 断言获取的interface{}为[]YuanqiConf
	confs, ok := confInterface.([]structs.YuanqiConf)
	if !ok {
		log.Println("Type assertion failed for YuanqiConfs, attempting default behavior")
		return getYuanqiConfInternal() // 递归调用内部函数，尝试默认配置
	}

	if len(confs) == 0 {
		log.Println("No configurations found in file:", basename)
		return getYuanqiConfInternal() // 递归调用内部函数，尝试默认配置
	}

	// 随机选择一个配置返回
	index := rand.Intn(len(confs))
	conf := confs[index]
	return conf.YuanqiAssistantID, conf.YuanqiToken
}

// 获取 ReplacementPairsIn
func GetReplacementPairsIn(options ...string) []structs.ReplacementPair {
	mu.Lock()
	defer mu.Unlock()
	return getReplacementPairsInInternal(options...)
}

// getReplacementPairsInInternal 内部递归函数，处理配置获取逻辑
func getReplacementPairsInInternal(options ...string) []structs.ReplacementPair {
	// 从instance中读取配置数组
	if len(options) == 0 || options[0] == "" {
		if instance != nil && len(instance.Settings.ReplacementPairsIn) > 0 {
			return instance.Settings.ReplacementPairsIn
		} else {
			return nil
		}
	}

	// 使用prompt包从指定的文件名中获取配置
	basename := options[0]
	confInterface, err := prompt.GetSettingFromFilename(basename, "ReplacementPairsIn")
	if err != nil {
		log.Printf("Error retrieving settings from file: %s, error: %v", basename, err)
		return getReplacementPairsInInternal() // 递归调用内部函数，不传递任何参数
	}

	// 断言获取的interface{}为[]ReplacementPair
	confs, ok := confInterface.([]structs.ReplacementPair)
	if !ok {
		log.Println("Type assertion failed for ReplacementPair, attempting default behavior")
		return getReplacementPairsInInternal() // 递归调用内部函数，尝试默认配置
	}

	if len(confs) == 0 {
		log.Println("No configurations found in file:", basename)
		return getReplacementPairsInInternal() // 递归调用内部函数，尝试默认配置
	}

	return confs
}

// 获取 ReplacementPairsOut
func GetReplacementPairsOut(options ...string) []structs.ReplacementPair {
	mu.Lock()
	defer mu.Unlock()
	return getReplacementPairsOutInternal(options...)
}

// getReplacementPairsInOutternal 内部递归函数，处理配置获取逻辑
func getReplacementPairsOutInternal(options ...string) []structs.ReplacementPair {
	// 从instance中读取配置数组
	if len(options) == 0 || options[0] == "" {
		if instance != nil && len(instance.Settings.ReplacementPairsOut) > 0 {
			return instance.Settings.ReplacementPairsOut
		} else {
			return nil
		}
	}

	// 使用prompt包从指定的文件名中获取配置
	basename := options[0]
	confInterface, err := prompt.GetSettingFromFilename(basename, "ReplacementPairsOut")
	if err != nil {
		log.Printf("Error retrieving settings from file: %s, error: %v", basename, err)
		return getReplacementPairsOutInternal() // 递归调用内部函数，不传递任何参数
	}

	// 断言获取的interface{}为[]ReplacementPair
	confs, ok := confInterface.([]structs.ReplacementPair)
	if !ok {
		log.Println("Type assertion failed for ReplacementPair, attempting default behavior")
		return getReplacementPairsOutInternal() // 递归调用内部函数，尝试默认配置
	}

	if len(confs) == 0 {
		log.Println("No configurations found in file:", basename)
		return getReplacementPairsOutInternal() // 递归调用内部函数，尝试默认配置
	}

	return confs
}

// 获取助手版本
func GetYuanqiVersion(options ...string) float64 {
	mu.Lock()
	defer mu.Unlock()
	return getYuanqiVersionInternal(options...)
}

func getYuanqiVersionInternal(options ...string) float64 {
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.YuanqiVersion
		}
		return 0.0 // 默认值或错误处理
	}

	basename := options[0]
	versionInterface, err := prompt.GetSettingFromFilename(basename, "YuanqiVersion")
	if err != nil {
		log.Println("Error retrieving YuanqiVersion:", err)
		return getYuanqiVersionInternal() // 递归调用内部函数，不传递任何参数
	}

	version, ok := versionInterface.(float64)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for YuanqiVersion, fetching default")
		return getYuanqiVersionInternal() // 递归调用内部函数，不传递任何参数
	}

	if version == 0 {
		return getYuanqiVersionInternal() // 递归调用内部函数，不传递任何参数
	}

	return version
}

// 获取聊天类型
func GetYuanqiChatType(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getYuanqiChatTypeInternal(options...)
}

func getYuanqiChatTypeInternal(options ...string) string {
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.YuanqiChatType
		}
		return "published" // 默认值或错误处理
	}

	basename := options[0]
	chatTypeInterface, err := prompt.GetSettingFromFilename(basename, "YuanqiChatType")
	if err != nil {
		log.Println("Error retrieving YuanqiChatType:", err)
		return getYuanqiChatTypeInternal() // 递归调用内部函数，不传递任何参数
	}

	chatType, ok := chatTypeInterface.(string)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for YuanqiChatType, fetching default")
		return getYuanqiChatTypeInternal() // 递归调用内部函数，不传递任何参数
	}

	if chatType == "" {
		return getYuanqiChatTypeInternal() // 递归调用内部函数，不传递任何参数
	}

	return chatType
}

// 获取API地址
func GetYuanqiApiPath(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getYuanqiApiPathInternal(options...)
}

func getYuanqiApiPathInternal(options ...string) string {
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.YuanqiApiPath
		}
		return "https://open.hunyuan.tencent.com/openapi/v1/agent/chat/completion" // 默认值或错误处理
	}

	basename := options[0]
	chatTypeInterface, err := prompt.GetSettingFromFilename(basename, "YuanqiApiPath")
	if err != nil {
		log.Println("Error retrieving YuanqiApiPath:", err)
		return getYuanqiApiPathInternal() // 递归调用内部函数，不传递任何参数
	}

	YuanqiApiPath, ok := chatTypeInterface.(string)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for YuanqiApiPath, fetching default")
		return getYuanqiApiPathInternal() // 递归调用内部函数，不传递任何参数
	}

	if YuanqiApiPath == "" {
		return getYuanqiApiPathInternal() // 递归调用内部函数，不传递任何参数
	}

	return YuanqiApiPath
}

// 获取YuanqiMaxToken
func GetYuanqiMaxToken(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getYuanqiMaxTokenInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getYuanqiMaxTokenInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.YuanqiMaxToken
		}
		return 0
	}

	// 使用传入的 basename
	basename := options[0]
	YuanqiMaxTokenInterface, err := prompt.GetSettingFromFilename(basename, "YuanqiMaxToken")
	if err != nil {
		log.Println("Error retrieving MaxTokenYuanQi:", err)
		return getYuanqiMaxTokenInternal() // 递归调用内部函数，不传递任何参数
	}

	YuanqiMaxToken, ok := YuanqiMaxTokenInterface.(int)
	if !ok { // 检查是否断言失败
		fmt.Println("Type assertion failed for MaxTokenYuanQi, fetching default")
		return getYuanqiMaxTokenInternal() // 递归调用内部函数，不传递任何参数
	}

	if YuanqiMaxToken == 0 {
		return getYuanqiMaxTokenInternal() // 递归调用内部函数，不传递任何参数
	}

	return YuanqiMaxToken
}

// 获取GroupAddNicknameToQ
func GetGroupAddNicknameToQ(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getGroupAddNicknameToQInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getGroupAddNicknameToQInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.GroupAddNicknameToQ
		}
		return 0
	}

	// 使用传入的 basename
	basename := options[0]
	YuanqiMaxTokenInterface, err := prompt.GetSettingFromFilename(basename, "GroupAddNicknameToQ")
	if err != nil {
		log.Println("Error retrieving GroupAddNicknameToQ:", err)
		return getGroupAddNicknameToQInternal() // 递归调用内部函数，不传递任何参数
	}

	GroupAddNicknameToQ, ok := YuanqiMaxTokenInterface.(int)
	if !ok { // 检查是否断言失败
		fmt.Println("Type assertion failed for GroupAddNicknameToQ, fetching default")
		return getGroupAddNicknameToQInternal() // 递归调用内部函数，不传递任何参数
	}

	if GroupAddNicknameToQ == 0 {
		return getGroupAddNicknameToQInternal() // 递归调用内部函数，不传递任何参数
	}

	return GroupAddNicknameToQ
}

// 获取GroupAddCardToQ
func GetGroupAddCardToQ(options ...string) int {
	mu.Lock()
	defer mu.Unlock()
	return getGroupAddCardToQInternal(options...)
}

// 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getGroupAddCardToQInternal(options ...string) int {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			return instance.Settings.GroupAddCardToQ
		}
		return 0
	}

	// 使用传入的 basename
	basename := options[0]
	GroupAddCardToQInterface, err := prompt.GetSettingFromFilename(basename, "GroupAddCardToQ")
	if err != nil {
		log.Println("Error retrieving GroupAddCardToQ:", err)
		return getGroupAddCardToQInternal() // 递归调用内部函数，不传递任何参数
	}

	GroupAddCardToQ, ok := GroupAddCardToQInterface.(int)
	if !ok { // 检查是否断言失败
		fmt.Println("Type assertion failed for GroupAddCardToQ, fetching default")
		return getGroupAddCardToQInternal() // 递归调用内部函数，不传递任何参数
	}

	if GroupAddCardToQ == 0 {
		return getGroupAddCardToQInternal() // 递归调用内部函数，不传递任何参数
	}

	return GroupAddCardToQ
}

// 获取 SpecialNameToQ
func GetSpecialNameToQ(options ...string) []structs.ReplacementNamePair {
	mu.Lock()
	defer mu.Unlock()
	return getSpecialNameToQInternal(options...)
}

// getReplacementPairsInOutternal 内部递归函数，处理配置获取逻辑
func getSpecialNameToQInternal(options ...string) []structs.ReplacementNamePair {
	// 从instance中读取配置数组
	if len(options) == 0 || options[0] == "" {
		if instance != nil && len(instance.Settings.SpecialNameToQ) > 0 {
			return instance.Settings.SpecialNameToQ
		} else {
			return nil
		}
	}

	// 使用prompt包从指定的文件名中获取配置
	basename := options[0]
	confInterface, err := prompt.GetSettingFromFilename(basename, "SpecialNameToQ")
	if err != nil {
		log.Printf("Error retrieving settings from file: %s, error: %v", basename, err)
		return getSpecialNameToQInternal() // 递归调用内部函数，不传递任何参数
	}

	// 断言获取的interface{}为[]ReplacementPair
	confs, ok := confInterface.([]structs.ReplacementNamePair)
	if !ok {
		log.Println("Type assertion failed for ReplacementPair, attempting default behavior")
		return getSpecialNameToQInternal() // 递归调用内部函数，尝试默认配置
	}

	if len(confs) == 0 {
		log.Println("No configurations found in file:", basename)
		return getSpecialNameToQInternal() // 递归调用内部函数，尝试默认配置
	}

	return confs
}

// GetConversationPath 获取ConversationPath，可接受basename作为参数
func GetConversationPath(options ...string) string {
	mu.Lock()
	defer mu.Unlock()
	return getConversationPathInternal(options...)
}

// getTyqwApiPathInternal 内部逻辑执行函数，不处理锁，可以安全地递归调用
func getConversationPathInternal(options ...string) string {
	// 检查是否有参数传递进来，以及是否为空字符串
	if len(options) == 0 || options[0] == "" {
		if instance != nil {
			ConversationPath := instance.Settings.ConversationPath
			if ConversationPath != "" {
				return ConversationPath
			} else {
				return "/conversation" // 默认值 /conversation
			}
		}
		// instance未加载
		return "/conversation" // 默认值 /conversation
	}

	// 使用传入的 basename
	basename := options[0]
	ConversationPathInterface, err := prompt.GetSettingFromFilename(basename, "ConversationPath")
	if err != nil {
		log.Println("Error retrieving ConversationPath:", err)
		return getConversationPathInternal() // 递归调用内部函数，不传递任何参数
	}

	ConversationPath, ok := ConversationPathInterface.(string)
	if !ok { // 检查类型断言是否失败
		log.Println("Type assertion failed for ConversationPath, fetching default")
		return getConversationPathInternal() // 递归调用内部函数，不传递任何参数
	}

	if ConversationPath == "" {
		return getConversationPathInternal()
	}

	return ConversationPath
}
