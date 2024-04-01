package applogic

import (
	"database/sql"

	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/hunyuan"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
	"github.com/hoshinonyaruko/gensokyo-llm/utils"
)

type App struct {
	DB     *sql.DB
	Client *hunyuan.Client
}

func (app *App) createConversation(conversationID string) error {
	_, err := app.DB.Exec("INSERT INTO conversations (id) VALUES (?)", conversationID)
	return err
}

func (app *App) addMessage(msg structs.Message) (string, error) {
	fmtf.Printf("添加信息：%v\n", msg)
	// Generate a new UUID for message ID
	messageID := utils.GenerateUUID() // Implement this function to generate a UUID

	_, err := app.DB.Exec("INSERT INTO messages (id, conversation_id, parent_message_id, text, role) VALUES (?, ?, ?, ?, ?)",
		messageID, msg.ConversationID, msg.ParentMessageID, msg.Text, msg.Role)
	return messageID, err
}

func (app *App) EnsureTablesExist() error {
	createMessagesTableSQL := `
    CREATE TABLE IF NOT EXISTS messages (
        id VARCHAR(36) PRIMARY KEY,
        conversation_id VARCHAR(36) NOT NULL,
        parent_message_id VARCHAR(36),
        text TEXT NOT NULL,
        role TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`

	_, err := app.DB.Exec(createMessagesTableSQL)
	if err != nil {
		return fmtf.Errorf("error creating messages table: %w", err)
	}

	// 其他创建

	return nil
}

func (app *App) EnsureUserContextTableExists() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS user_context (
        user_id INTEGER PRIMARY KEY,
        conversation_id TEXT NOT NULL,
        parent_message_id TEXT
    );`

	_, err := app.DB.Exec(createTableSQL)
	if err != nil {
		return fmtf.Errorf("error creating user_context table: %w", err)
	}

	return nil
}

func (app *App) handleUserContext(userID int64) (string, string, error) {
	var conversationID, parentMessageID string

	// 检查用户上下文是否存在
	query := `SELECT conversation_id, parent_message_id FROM user_context WHERE user_id = ?`
	err := app.DB.QueryRow(query, userID).Scan(&conversationID, &parentMessageID)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户上下文不存在，创建新的
			conversationID = utils.GenerateUUID() // 假设generateUUID()是一个生成UUID的函数
			parentMessageID = ""

			// 插入新的用户上下文
			insertQuery := `INSERT INTO user_context (user_id, conversation_id, parent_message_id) VALUES (?, ?, ?)`
			_, err = app.DB.Exec(insertQuery, userID, conversationID, parentMessageID)
			if err != nil {
				return "", "", err
			}
		} else {
			// 查询过程中出现了其他错误
			return "", "", err
		}
	}

	// 返回conversationID和parentMessageID
	return conversationID, parentMessageID, nil
}

func (app *App) migrateUserToNewContext(userID int64) error {
	// 生成新的conversationID
	newConversationID := utils.GenerateUUID() // 假设generateUUID()是一个生成UUID的函数

	// 更新用户上下文
	updateQuery := `UPDATE user_context SET conversation_id = ?, parent_message_id = '' WHERE user_id = ?`
	_, err := app.DB.Exec(updateQuery, newConversationID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) updateUserContext(userID int64, parentMessageID string) error {
	updateQuery := `UPDATE user_context SET parent_message_id = ? WHERE user_id = ?`
	_, err := app.DB.Exec(updateQuery, parentMessageID, userID)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) getHistory(conversationID, parentMessageID string) ([]structs.Message, error) {
	var history []structs.Message

	// SQL 查询获取历史信息
	query := `SELECT text, role, created_at FROM messages
              WHERE conversation_id = ? AND created_at <= (SELECT created_at FROM messages WHERE id = ?)
              ORDER BY created_at ASC`
	rows, err := app.DB.Query(query, conversationID, parentMessageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var previousText string
	for rows.Next() {
		var msg structs.Message
		err := rows.Scan(&msg.Text, &msg.Role, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		if msg.Text == previousText {
			continue
		}
		previousText = msg.Text

		// 根据角色添加不同的消息格式
		historyEntry := structs.Message{
			Role: msg.Role,
			Text: msg.Text,
		}
		fmtf.Printf("加入:%v\n", historyEntry)
		history = append(history, historyEntry)
	}
	return history, nil
}
