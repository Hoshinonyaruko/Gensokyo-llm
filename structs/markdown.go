package structs

type RenderData struct {
	Label        string `json:"label"`
	VisitedLabel string `json:"visited_label"`
	Style        int    `json:"style"`
}

type Permission struct {
	Type           int      `json:"type"`
	SpecifyRoleIDs []string `json:"specify_role_ids"`
}

type Action struct {
	Type                 int        `json:"type"`
	Permission           Permission `json:"permission"`
	ClickLimit           int        `json:"click_limit"`
	UnsupportTips        string     `json:"unsupport_tips"`
	Data                 string     `json:"data"`
	AtBotShowChannelList bool       `json:"at_bot_show_channel_list"`
	Enter                bool       `json:"enter"` //指令按钮可用，点击按钮后直接自动发送 data，默认 false。支持版本 8983
	Reply                bool       `json:"reply"` //指令按钮可用，指令是否带引用回复本消息，默认 false。支持版本 8983
}

type Button struct {
	ID         string     `json:"id"`
	RenderData RenderData `json:"render_data"`
	Action     Action     `json:"action"`
}

type Row struct {
	Buttons []Button `json:"buttons"`
}

type KeyboardContent struct {
	Rows []Row `json:"rows"`
}

type Keyboard struct {
	Content KeyboardContent `json:"content"`
}

type Markdown struct {
	Content string `json:"content"`
}

type PromptKeyboardMarkdown struct {
	Markdown  Markdown `json:"markdown"`
	Keyboard  Keyboard `json:"keyboard"`
	Content   string   `json:"content"`
	MsgID     string   `json:"msg_id"`
	Timestamp string   `json:"timestamp"`
	MsgType   int      `json:"msg_type"`
}
