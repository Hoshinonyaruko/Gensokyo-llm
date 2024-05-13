package structs

type Message struct {
	ConversationID  string `json:"conversationId"`
	ParentMessageID string `json:"parentMessageId"`
	Text            string `json:"message"`
	Role            string `json:"role"`
	CreatedAt       string `json:"created_at"`
}

type WXRequestMessage struct {
	ConversationID  string `json:"conversationId"`
	ParentMessageID string `json:"parentMessageId"`
	Text            string `json:"message"`
	Role            string `json:"role"`
	CreatedAt       string `json:"created_at"`
}

type WXRequestMessageF struct {
	ConversationID  string     `json:"conversationId"`
	ParentMessageID string     `json:"parentMessageId"`
	Text            string     `json:"message"`
	Role            string     `json:"role"`
	CreatedAt       string     `json:"created_at"`
	WXFunction      WXFunction `json:"functions,omitempty"`
}

type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

// 群信息事件
type OnebotGroupMessage struct {
	RawMessage      string      `json:"raw_message"`
	MessageID       int         `json:"message_id"`
	GroupID         int64       `json:"group_id"` // Can be either string or int depending on p.Settings.CompleteFields
	MessageType     string      `json:"message_type"`
	PostType        string      `json:"post_type"`
	SelfID          int64       `json:"self_id"` // Can be either string or int
	Sender          Sender      `json:"sender"`
	SubType         string      `json:"sub_type"`
	Time            int64       `json:"time"`
	Avatar          string      `json:"avatar,omitempty"`
	Echo            string      `json:"echo,omitempty"`
	Message         interface{} `json:"message"` // For array format
	MessageSeq      int         `json:"message_seq"`
	Font            int         `json:"font"`
	UserID          int64       `json:"user_id"`
	RealMessageType string      `json:"real_message_type,omitempty"`  //当前信息的真实类型 group group_private guild guild_private
	IsBindedGroupId bool        `json:"is_binded_group_id,omitempty"` //当前群号是否是binded后的
	IsBindedUserId  bool        `json:"is_binded_user_id,omitempty"`  //当前用户号号是否是binded后的
}

type Sender struct {
	Nickname string `json:"nickname"`
	TinyID   string `json:"tiny_id"`
	UserID   int64  `json:"user_id"`
	Role     string `json:"role,omitempty"`
	Card     string `json:"card,omitempty"`
	Sex      string `json:"sex,omitempty"`
	Age      int32  `json:"age,omitempty"`
	Area     string `json:"area,omitempty"`
	Level    string `json:"level,omitempty"`
	Title    string `json:"title,omitempty"`
}

// 定义请求消息的结构体
type WXMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

// 定义请求消息的结构体
type WXMessageF struct {
	Content      string         `json:"content"`
	Role         string         `json:"role"`
	FunctionCall WXFunctionCall `json:"function_call,omitempty"`
}

// 定义请求负载的结构体
type WXRequestPayload struct {
	Messages        []WXMessage `json:"messages"`
	Stream          bool        `json:"stream,omitempty"`
	Temperature     float64     `json:"temperature,omitempty"`
	TopP            float64     `json:"top_p,omitempty"`
	PenaltyScore    float64     `json:"penalty_score,omitempty"`
	System          string      `json:"system,omitempty"`
	Stop            []string    `json:"stop,omitempty"`
	MaxOutputTokens int         `json:"max_output_tokens,omitempty"`
	UserID          string      `json:"user_id,omitempty"`
}

// 定义请求负载的结构体
type WXRequestPayloadF struct {
	Messages        []WXMessage  `json:"messages"`
	Functions       []WXFunction `json:"functions,omitempty"`
	Stream          bool         `json:"stream,omitempty"`
	Temperature     float64      `json:"temperature,omitempty"`
	TopP            float64      `json:"top_p,omitempty"`
	PenaltyScore    float64      `json:"penalty_score,omitempty"`
	System          string       `json:"system,omitempty"`
	Stop            []string     `json:"stop,omitempty"`
	MaxOutputTokens int          `json:"max_output_tokens,omitempty"`
	UserID          string       `json:"user_id,omitempty"`
	ResponseFormat  string       `json:"response_format,omitempty"`
	ToolChoice      ToolChoice   `json:"tool_choice,omitempty"`
}

// Function 描述了一个可调用的函数的细节
type Function struct {
	Name string `json:"name"` // 函数名
}

// ToolChoice 描述了要使用的工具和具体的函数选择
type ToolChoice struct {
	Type     string   `json:"type"`     // 工具类型，这里固定为"function"
	Function Function `json:"function"` // 指定要使用的函数
}

type ChatGPTMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatGPTRequest struct {
	Model    string           `json:"model"`
	Messages []ChatGPTMessage `json:"messages"`
	SafeMode bool             `json:"safe_mode"`
}

type ChatGPTResponseChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type ChatGPTResponse struct {
	Choices []ChatGPTResponseChoice `json:"choices"`
}

// 定义事件数据的结构体，以匹配OpenAI返回的格式
type GPTEventData struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

type TyqwSSEData struct {
	Output struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
				Role    string `json:"role"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	} `json:"output"`
	Usage struct {
		TotalTokens  int `json:"total_tokens"`
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
}

// 定义用于累积使用情况的结构（如果API提供此信息）
type GPTUsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// InterfaceBody 结构体定义
type InterfaceBody struct {
	Content        string   `json:"content"`
	State          int      `json:"state"`
	PromptKeyboard []string `json:"prompt_keyboard"`
	ActionButton   int      `json:"action_button"`
	CallbackData   string   `json:"callback_data"`
}

// EmbeddingData 结构体用于解析embedding接口返回的数据
type EmbeddingDataErnie struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingResponse 结构体用于解析整个API响应
type EmbeddingResponseErnie struct {
	ID     string               `json:"id"`
	Object string               `json:"object"`
	Data   []EmbeddingDataErnie `json:"data"`
}

// Function 描述了一个可调用的函数的结构
type WXFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Responses   map[string]interface{} `json:"responses,omitempty"`
	Examples    [][]WXExample          `json:"examples,omitempty"`
}

// Example 描述了函数调用的一个示例
type WXExample struct {
	Role         string          `json:"role"`
	Content      string          `json:"content"`
	Name         string          `json:"name,omitempty"`
	FunctionCall *WXFunctionCall `json:"function_call,omitempty"`
}

// FunctionCall 描述了一个函数调用
type WXFunctionCall struct {
	Name      string                 `json:"name,omitempty"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Thought   string                 `json:"thought,omitempty"`
}

type Settings struct {
	AllApi                  bool     `yaml:"allApi"`
	SecretId                string   `yaml:"secretId"`
	SecretKey               string   `yaml:"secretKey"`
	Region                  string   `yaml:"region"`
	UseSse                  bool     `yaml:"useSse"`
	Port                    int      `yaml:"port"`
	SelfPath                string   `yaml:"selfPath"`
	HttpPath                string   `yaml:"path"`
	Lotus                   string   `yaml:"lotus"`
	PathToken               string   `yaml:"pathToken"`
	SystemPrompt            []string `yaml:"systemPrompt"`
	IPWhiteList             []string `yaml:"iPWhiteList"`
	ApiType                 int      `yaml:"apiType"`
	Proxy                   string   `yaml:"proxy"`
	UrlSendPics             bool     `yaml:"urlSendPics"`             // 自己构造图床加速图片发送
	MdPromptKeyboardAtGroup bool     `yaml:"mdPromptKeyboardAtGroup"` // 群内使用md能力模拟PromptKeyboard

	HunyuanType      int `yaml:"hunyuanType"`
	MaxTokensHunyuan int `yaml:"maxTokensHunyuan"`

	WenxinAccessToken     string  `yaml:"wenxinAccessToken"`
	WenxinApiPath         string  `yaml:"wenxinApiPath"`
	MaxTokenWenxin        int     `yaml:"maxTokenWenxin"`
	WenxinTopp            float64 `yaml:"wenxinTopp"`
	WnxinPenaltyScore     float64 `yaml:"wenxinPenaltyScore"`
	WenxinMaxOutputTokens int     `yaml:"wenxinMaxOutputTokens"`
	WenxinEmbeddingUrl    string  `yaml:"wenxinEmbeddingUrl"`

	GptModel        string `yaml:"gptModel"`
	GptApiPath      string `yaml:"gptApiPath"`
	GptToken        string `yaml:"gptToken"`
	MaxTokenGpt     int    `yaml:"maxTokenGpt"`
	GptSafeMode     bool   `yaml:"gptSafeMode"`
	GptSseType      int    `yaml:"gptSseType"`
	GptEmbeddingUrl string `yaml:"gptEmbeddingUrl"`
	StandardGptApi  bool   `yaml:"standardGptApi"`

	Groupmessage       bool `yaml:"groupMessage"`
	SplitByPuntuations int  `yaml:"splitByPuntuations"`

	FirstQ  []string `yaml:"firstQ"`
	FirstA  []string `yaml:"firstA"`
	SecondQ []string `yaml:"secondQ"`
	SecondA []string `yaml:"secondA"`
	ThirdQ  []string `yaml:"thirdQ"`
	ThirdA  []string `yaml:"thirdA"`

	SensitiveMode        bool     `yaml:"sensitiveMode"`
	SensitiveModeType    int      `yaml:"sensitiveModeType"`
	DefaultChangeWord    string   `yaml:"defaultChangeWord"`
	AntiPromptAttackPath string   `yaml:"antiPromptAttackPath"`
	ReverseUserPrompt    bool     `yaml:"reverseUserPrompt"`
	IgnoreExtraTips      bool     `yaml:"ignoreExtraTips"`
	SaveResponses        []string `yaml:"saveResponses"`
	RestoreCommand       []string `yaml:"restoreCommand"`
	RestoreResponses     []string `yaml:"restoreResponses"`
	UsePrivateSSE        bool     `yaml:"usePrivateSSE"`
	Promptkeyboard       []string `yaml:"promptkeyboard"`
	No4Promptkeyboard    bool     `yaml:"no4Promptkeyboard"`
	Savelogs             bool     `yaml:"savelogs"`
	AntiPromptLimit      float64  `yaml:"antiPromptLimit"`

	UseCache       bool `yaml:"useCache"`
	CacheThreshold int  `yaml:"cacheThreshold"`
	CacheChance    int  `yaml:"cacheChance"`
	EmbeddingType  int  `yaml:"embeddingType"`

	PrintHanming  bool    `yaml:"printHanming"`
	CacheK        float64 `yaml:"cacheK"`
	CacheN        int64   `yaml:"cacheN"`
	PrintVector   bool    `yaml:"printVector"`
	VToBThreshold float64 `yaml:"vToBThreshold"`
	GptModeration bool    `yaml:"gptModeration"`

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

	RwkvApiPath          string   `yaml:"rwkvApiPath"`
	RwkvMaxTokens        int      `yaml:"rwkvMaxTokens"`
	RwkvTemperature      float64  `yaml:"rwkvTemperature"`
	RwkvTopP             float64  `yaml:"rwkvTopP"`
	RwkvPresencePenalty  float64  `yaml:"rwkvPresencePenalty"`
	RwkvFrequencyPenalty float64  `yaml:"rwkvFrequencyPenalty"`
	RwkvPenaltyDecay     float64  `yaml:"rwkvPenaltyDecay"`
	RwkvTopK             int      `yaml:"rwkvTopK"`
	RwkvGlobalPenalty    bool     `yaml:"rwkvGlobalPenalty"`
	RwkvStream           bool     `yaml:"rwkvStream"`
	RwkvStop             []string `yaml:"rwkvStop"`
	RwkvUserName         string   `yaml:"rwkvUserName"`
	RwkvAssistantName    string   `yaml:"rwkvAssistantName"`
	RwkvSystemName       string   `yaml:"rwkvSystemName"`
	RwkvPreSystem        bool     `yaml:"rwkvPreSystem"`
	RwkvSseType          int      `yaml:"rwkvSseType"`
	HideExtraLogs        bool     `yaml:"hideExtraLogs"`

	TyqwApiPath           string   `yaml:"tyqwApiPath"`
	TyqwMaxTokens         int      `yaml:"tyqwMaxTokens"`
	TyqwTemperature       float64  `yaml:"tyqwTemperature"`
	TyqwTopP              float64  `yaml:"tyqwTopP"`
	TyqwPresencePenalty   float64  `yaml:"tyqwPresencePenalty"`
	TyqwFrequencyPenalty  float64  `yaml:"tyqwFrequencyPenalty"`
	TyqwPenaltyDecay      float64  `yaml:"tyqwPenaltyDecay"`
	TyqwTopK              int      `yaml:"tyqwTopK"`
	TyqwGlobalPenalty     bool     `yaml:"tyqwGlobalPenalty"`
	TyqwStop              []string `yaml:"tyqwStop"`
	TyqwUserName          string   `yaml:"tyqwUserName"`
	TyqwAssistantName     string   `yaml:"tyqwAssistantName"`
	TyqwSystemName        string   `yaml:"tyqwSystemName"`
	TyqwPreSystem         bool     `yaml:"tyqwPreSystem"`
	TyqwSseType           int      `yaml:"tyqwSseType"`
	TyqwRepetitionPenalty float64  `yaml:"tyqwRepetitionPenalty"`
	TyqwSeed              int64    `yaml:"tyqwSeed"`
	TyqwEnableSearch      bool     `yaml:"tyqwEnableSearch"`
	TyqwModel             string   `yaml:"tyqwModel"`
	TyqwApiKey            string   `yaml:"tyqwApiKey"`
	TyqwWorkspace         string   `yaml:"tyqwWorkspace"`

	WSServerToken string `yaml:"wsServerToken"`
	WSPath        string `yaml:"wsPath"`

	PromptMarkType        int      `yaml:"promptMarkType"`
	PromptMarksLength     int      `yaml:"promptMarksLength"`
	PromptMarks           []string `yaml:"promptMarks"`
	EnhancedQA            bool     `yaml:"enhancedQA"`
	PromptChoicesQ        []string `yaml:"promptChoicesQ"`
	PromptChoicesA        []string `yaml:"promptChoicesA"`
	EnhancedPromptChoices bool     `yaml:"enhancedPromptChoices"`
	SwitchOnQ             []string `yaml:"switchOnQ"`
	SwitchOnA             []string `yaml:"switchOnA"`
	ExitOnQ               []string `yaml:"exitOnQ"`
	ExitOnA               []string `yaml:"exitOnA"`
	EnvType               int      `yaml:"envType"`
	EnvPics               []string `yaml:"envPics"`     //ai太慢了,而且影响气泡了,只能手动了
	EnvContents           []string `yaml:"envContents"` //ai太慢了,而且影响气泡了,只能手动了
	PromptCoverQ          []string `yaml:"promptCoverQ"`
	PromptCoverA          []string `yaml:"promptCoverA"` //暂时用不上 待实现
}

type MetaEvent struct {
	PostType      string `json:"post_type"`
	MetaEventType string `json:"meta_event_type"`
	Time          int64  `json:"time"`
	SelfID        int64  `json:"self_id"`
	Interval      int    `json:"interval"`
	Status        struct {
		AppEnabled     bool  `json:"app_enabled"`
		AppGood        bool  `json:"app_good"`
		AppInitialized bool  `json:"app_initialized"`
		Good           bool  `json:"good"`
		Online         bool  `json:"online"`
		PluginsGood    *bool `json:"plugins_good"`
		Stat           struct {
			PacketReceived  int   `json:"packet_received"`
			PacketSent      int   `json:"packet_sent"`
			PacketLost      int   `json:"packet_lost"`
			MessageReceived int   `json:"message_received"`
			MessageSent     int   `json:"message_sent"`
			DisconnectTimes int   `json:"disconnect_times"`
			LostTimes       int   `json:"lost_times"`
			LastMessageTime int64 `json:"last_message_time"`
		} `json:"stat"`
	} `json:"status"`
}

type NoticeEvent struct {
	GroupID    int64  `json:"group_id"`
	NoticeType string `json:"notice_type"`
	OperatorID int64  `json:"operator_id"`
	PostType   string `json:"post_type"`
	SelfID     int64  `json:"self_id"`
	SubType    string `json:"sub_type"`
	Time       int64  `json:"time"`
	UserID     int64  `json:"user_id"`
}

type RobotStatus struct {
	SelfID          int64  `json:"self_id"`
	Date            string `json:"date"`
	Online          bool   `json:"online"`
	MessageReceived int    `json:"message_received"`
	MessageSent     int    `json:"message_sent"`
	LastMessageTime int64  `json:"last_message_time"`
	InvitesReceived int    `json:"invites_received"`
	KicksReceived   int    `json:"kicks_received"`
	DailyDAU        int    `json:"daily_dau"`
}

type OnebotActionMessage struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
	Echo   interface{}            `json:"echo,omitempty"`
}

type CustomRecord struct {
	UserID        int64
	PromptStr     string
	PromptStrStat int        // New integer field for storing promptstr_stat
	Strs          [10]string // Array to store str1 to str10
}
