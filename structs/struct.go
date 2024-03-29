package structs

type Message struct {
	ConversationID  string `json:"conversationId"`
	ParentMessageID string `json:"parentMessageId"`
	Text            string `json:"message"`
	Role            string `json:"role"`
	CreatedAt       string `json:"created_at"`
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

// 定义用于累积使用情况的结构（如果API提供此信息）
type GPTUsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
