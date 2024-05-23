package structs

type ContentItem struct {
	Type    string   `json:"type,omitempty"`     // "text" or "file_url"
	Text    string   `json:"text,omitempty"`     // 文本内容
	FileURL *FileURL `json:"file_url,omitempty"` // 文件内容
}

type FileURL struct {
	Type string `json:"type"` // 文件类型: image, video, audio, pdf, doc, txt 等
	URL  string `json:"url"`  // 文件的URL地址
}

type MessageContent struct {
	Role    string        `json:"role"`    // "user" 或 "assistant"
	Content []ContentItem `json:"content"` // 内容列表
}

type RequestDataYuanQi struct {
	AssistantID string           `json:"assistant_id"`        // 助手ID
	Version     float64          `json:"version,omitempty"`   // 助手版本
	UserID      string           `json:"user_id"`             // 用户ID
	Stream      bool             `json:"stream"`              // 是否启用流式返回
	ChatType    string           `json:"chat_type,omitempty"` // 聊天类型
	Messages    []MessageContent `json:"messages"`            // 消息历史和当前消息
}
