package applogic

import (
	"time"

	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
)

// 直接根据缓存来储存上下文
// 其实向量缓存是一个单轮的QA缓存,因为这个项目很初步,很显然无法应对上下文场景的缓存
// 通过这种方式,将每次缓存的内容也加入上下文,可能会有一个初步的效果提升.
func (app *App) AddSingleContext(message structs.OnebotGroupMessage, responseText string) bool {
	// 请求conversation api 增加当前用户上下文
	conversationID, parentMessageID, err := app.handleUserContext(message.UserID)
	if err != nil {
		fmtf.Printf("error in AddSingleContext app.handleUserContex :%v", err)
		return false
	}

	// 构造用户消息并添加到上下文
	userMessage := structs.Message{
		ConversationID:  conversationID,
		ParentMessageID: parentMessageID,
		Text:            message.Message.(string),
		Role:            "user",
		CreatedAt:       time.Now().Format(time.RFC3339),
	}
	userMessageID, err := app.addMessage(userMessage)
	if err != nil {
		fmtf.Printf("error in AddSingleContext app.addMessage(userMessage) :%v", err)
		return false
	}

	// 构造助理消息并添加到上下文
	assistantMessage := structs.Message{
		ConversationID:  conversationID,
		ParentMessageID: userMessageID,
		Text:            responseText,
		Role:            "assistant",
		CreatedAt:       time.Now().Format(time.RFC3339),
	}
	_, err = app.addMessage(assistantMessage)
	if err != nil {
		fmtf.Printf("error in AddSingleContext app.addMessage(assistantMessage) :%v", err)
		return false
	}

	return true
}

// 直接根据缓存来储存上下文
// 其实向量缓存是一个单轮的QA缓存,因为这个项目很初步,很显然无法应对上下文场景的缓存
// 通过这种方式,将每次缓存的内容也加入上下文,可能会有一个初步的效果提升.
func (app *App) AddSingleContextSP(message structs.OnebotGroupMessageS, responseText string) bool {
	// 请求conversation api 增加当前用户上下文
	conversationID, parentMessageID, err := app.handleUserContextSP(message.UserID)
	if err != nil {
		fmtf.Printf("error in AddSingleContext app.handleUserContex :%v", err)
		return false
	}

	// 构造用户消息并添加到上下文
	userMessage := structs.Message{
		ConversationID:  conversationID,
		ParentMessageID: parentMessageID,
		Text:            message.Message.(string),
		Role:            "user",
		CreatedAt:       time.Now().Format(time.RFC3339),
	}
	userMessageID, err := app.addMessage(userMessage)
	if err != nil {
		fmtf.Printf("error in AddSingleContext app.addMessage(userMessage) :%v", err)
		return false
	}

	// 构造助理消息并添加到上下文
	assistantMessage := structs.Message{
		ConversationID:  conversationID,
		ParentMessageID: userMessageID,
		Text:            responseText,
		Role:            "assistant",
		CreatedAt:       time.Now().Format(time.RFC3339),
	}
	_, err = app.addMessage(assistantMessage)
	if err != nil {
		fmtf.Printf("error in AddSingleContext app.addMessage(assistantMessage) :%v", err)
		return false
	}

	return true
}
