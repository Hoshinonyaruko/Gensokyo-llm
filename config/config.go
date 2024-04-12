package config

import (
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"gopkg.in/yaml.v3"
)

var (
	instance *Config
	mu       sync.Mutex
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type Config struct {
	Version  int      `yaml:"version"`
	Settings Settings `yaml:"settings"`
}

type Settings struct {
	SecretId                  string   `yaml:"secretId"`
	SecretKey                 string   `yaml:"secretKey"`
	Region                    string   `yaml:"region"`
	UseSse                    bool     `yaml:"useSse"`
	Port                      int      `yaml:"port"`
	HttpPath                  string   `yaml:"path"`
	SystemPrompt              []string `yaml:"systemPrompt"`
	IPWhiteList               []string `yaml:"iPWhiteList"`
	MaxTokensHunyuan          int      `yaml:"maxTokensHunyuan"`
	ApiType                   int      `yaml:"apiType"`
	WenxinAccessToken         string   `yaml:"wenxinAccessToken"`
	WenxinApiPath             string   `yaml:"wenxinApiPath"`
	MaxTokenWenxin            int      `yaml:"maxTokenWenxin"`
	GptModel                  string   `yaml:"gptModel"`
	GptApiPath                string   `yaml:"gptApiPath"`
	GptToken                  string   `yaml:"gptToken"`
	MaxTokenGpt               int      `yaml:"maxTokenGpt"`
	GptSafeMode               bool     `yaml:"gptSafeMode"`
	GptSseType                int      `yaml:"gptSseType"`
	Groupmessage              bool     `yaml:"groupMessage"`
	SplitByPuntuations        int      `yaml:"splitByPuntuations"`
	HunyuanType               int      `yaml:"hunyuanType"`
	FirstQ                    []string `yaml:"firstQ"`
	FirstA                    []string `yaml:"firstA"`
	SecondQ                   []string `yaml:"secondQ"`
	SecondA                   []string `yaml:"secondA"`
	ThirdQ                    []string `yaml:"thirdQ"`
	ThirdA                    []string `yaml:"thirdA"`
	SensitiveMode             bool     `yaml:"sensitiveMode"`
	SensitiveModeType         int      `yaml:"sensitiveModeType"`
	DefaultChangeWord         string   `yaml:"defaultChangeWord"`
	AntiPromptAttackPath      string   `yaml:"antiPromptAttackPath"`
	ReverseUserPrompt         bool     `yaml:"reverseUserPrompt"`
	IgnoreExtraTips           bool     `yaml:"ignoreExtraTips"`
	SaveResponses             []string `yaml:"saveResponses"`
	RestoreCommand            []string `yaml:"restoreCommand"`
	RestoreResponses          []string `yaml:"restoreResponses"`
	UsePrivateSSE             bool     `yaml:"usePrivateSSE"`
	Promptkeyboard            []string `yaml:"promptkeyboard"`
	Savelogs                  bool     `yaml:"savelogs"`
	AntiPromptLimit           float64  `yaml:"antiPromptLimit"`
	UseCache                  bool     `yaml:"useCache"`
	CacheThreshold            int      `yaml:"cacheThreshold"`
	CacheChance               int      `yaml:"cacheChance"`
	EmbeddingType             int      `yaml:"embeddingType"`
	WenxinEmbeddingUrl        string   `yaml:"wenxinEmbeddingUrl"`
	GptEmbeddingUrl           string   `yaml:"gptEmbeddingUrl"`
	PrintHanming              bool     `yaml:"printHanming"`
	CacheK                    float64  `yaml:"cacheK"`
	CacheN                    int64    `yaml:"cacheN"`
	PrintVector               bool     `yaml:"printVector"`
	VToBThreshold             float64  `yaml:"vToBThreshold"`
	GptModeration             bool     `yaml:"gptModeration"`
	WenxinTopp                float64  `yaml:"wenxinTopp"`
	WnxinPenaltyScore         float64  `yaml:"wenxinPenaltyScore"`
	WenxinMaxOutputTokens     int      `yaml:"wenxinMaxOutputTokens"`
	VectorSensitiveFilter     bool     `yaml:"vectorSensitiveFilter"`
	VertorSensitiveThreshold  int      `yaml:"vertorSensitiveThreshold"`
	AllowedLanguages          []string `yaml:"allowedLanguages"`
	LanguagesResponseMessages []string `yaml:"langResponseMessages"`
	QuestionMaxLenth          int      `yaml:"questionMaxLenth"`
	QmlResponseMessages       []string `yaml:"qmlResponseMessages"`
	BlacklistResponseMessages []string `yaml:"blacklistResponseMessages"`
	NoContext                 bool     `yaml:"noContext"`
	WithdrawCommand           []string `yaml:"withdrawCommand"`
	FunctionMode              bool     `yaml:"functionMode"`
	FunctionPath              string   `yaml:"functionPath"`
	UseFunctionPromptkeyboard bool     `yaml:"useFunctionPromptkeyboard"`
	AIPromptkeyboardPath      string   `yaml:"AIPromptkeyboardPath"`
	UseAIPromptkeyboard       bool     `yaml:"useAIPromptkeyboard"`
	SplitByPuntuationsGroup   int      `yaml:"splitByPuntuationsGroup"`
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
func GetuseSse() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.UseSse
	}
	return false
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

// 获取getHttpPath
func GetHttpPath() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.HttpPath
	}
	return "0"
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
func GetMaxTokensHunyuan() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.MaxTokensHunyuan
	}
	return 4096
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
func GetWenxinApiPath() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.WenxinApiPath
	}
	return "0"
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
func GetGptModel() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.GptModel
	}
	return "0"
}

// 获取GptApiPath
func GetGptApiPath() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.GptApiPath
	}
	return "0"
}

// 获取GptToken
func GetGptToken() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.GptToken
	}
	return "0"
}

// 获取MaxTokenGpt
func GetMaxTokenGpt() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.MaxTokenGpt
	}
	return 0
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

// 获取UseCache
func GetUseCache() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.UseCache
	}
	return false
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
func GetAIPromptkeyboardPath() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.AIPromptkeyboardPath
	}
	return ""
}
