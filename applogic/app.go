package applogic

import (
	"database/sql"
	"fmt"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
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
	// 创建 messages 表
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
		return fmt.Errorf("error creating messages table: %w", err)
	}

	// 为 conversation_id 创建索引
	createConvIDIndexSQL := `CREATE INDEX IF NOT EXISTS idx_conversation_id ON messages(conversation_id);`

	_, err = app.DB.Exec(createConvIDIndexSQL)
	if err != nil {
		return fmt.Errorf("error creating index on messages(conversation_id): %w", err)
	}

	// 为 parent_message_id 创建索引（如果您需要通过 parent_message_id 查询）
	createParentMsgIDIndexSQL := `CREATE INDEX IF NOT EXISTS idx_parent_message_id ON messages(parent_message_id);`

	_, err = app.DB.Exec(createParentMsgIDIndexSQL)
	if err != nil {
		return fmt.Errorf("error creating index on messages(parent_message_id): %w", err)
	}

	// 为 created_at 创建索引（如果您需要对消息进行时间排序）
	createCreatedAtIndexSQL := `CREATE INDEX IF NOT EXISTS idx_created_at ON messages(created_at);`

	_, err = app.DB.Exec(createCreatedAtIndexSQL)
	if err != nil {
		return fmt.Errorf("error creating index on messages(created_at): %w", err)
	}

	// 其他创建

	return nil
}

// 问题Q 向量表
func (app *App) EnsureEmbeddingsTablesExist() error {
	createMessagesTableSQL := `
    CREATE TABLE IF NOT EXISTS vector_data (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        text TEXT NOT NULL,
        vector BLOB NOT NULL,
        norm FLOAT NOT NULL,
        group_id INTEGER NOT NULL
    );`

	_, err := app.DB.Exec(createMessagesTableSQL)
	if err != nil {
		return fmt.Errorf("error creating messages table: %w", err)
	}

	// 为group_id和norm添加索引
	createIndexSQL := `
    CREATE INDEX IF NOT EXISTS idx_group_id ON vector_data(group_id);
    CREATE INDEX IF NOT EXISTS idx_norm ON vector_data(norm);`

	_, err = app.DB.Exec(createIndexSQL)
	if err != nil {
		return fmtf.Errorf("error creating indexes: %w", err)
	}

	// 其他创建

	return nil
}

// 敏感词表
func (app *App) EnsureSensitiveWordsTableExists() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS sensitive_words (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        text TEXT NOT NULL,
        vector BLOB NOT NULL,
        norm FLOAT NOT NULL,
        group_id INTEGER NOT NULL
    );`

	_, err := app.DB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating sensitive_words table: %w", err)
	}

	// 为group_id和norm添加索引
	createIndexSQL := `
    CREATE INDEX IF NOT EXISTS idx_sensitive_words_group_id ON sensitive_words(group_id);
    CREATE INDEX IF NOT EXISTS idx_sensitive_words_norm ON sensitive_words(norm);`

	_, err = app.DB.Exec(createIndexSQL)
	if err != nil {
		return fmt.Errorf("error creating indexes: %w", err)
	}

	return nil
}

func (app *App) EnsureQATableExist() error {
	// 创建 questions 表
	createQuestionsTableSQL := `
    CREATE TABLE IF NOT EXISTS questions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        question_text TEXT NOT NULL,
        vector_data_id INTEGER NOT NULL,
        UNIQUE(question_text),
        FOREIGN KEY(vector_data_id) REFERENCES vector_data(id)
    );`

	_, err := app.DB.Exec(createQuestionsTableSQL)
	if err != nil {
		return fmt.Errorf("error creating questions table: %w", err)
	}

	// 创建 qa_cache 表
	createQACacheTableSQL := `
    CREATE TABLE IF NOT EXISTS qa_cache (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        answer_text TEXT NOT NULL,
        question_id INTEGER NOT NULL,
        FOREIGN KEY(question_id) REFERENCES questions(id)
    );`

	_, err = app.DB.Exec(createQACacheTableSQL)
	if err != nil {
		return fmt.Errorf("error creating qa_cache table: %w", err)
	}

	// 为 qa_cache 表的 question_id 字段创建索引
	createIndexSQL := `CREATE INDEX IF NOT EXISTS idx_question_id ON qa_cache(question_id);`

	_, err = app.DB.Exec(createIndexSQL)
	if err != nil {
		return fmt.Errorf("error creating index on qa_cache(question_id): %w", err)
	}

	return nil
}

func (app *App) EnsureCustomTableExist() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS custom_table (
        user_id INTEGER PRIMARY KEY,
        promptstr TEXT NOT NULL,
        promptstr_stat INTEGER,
        str1 TEXT,
        str2 TEXT,
        str3 TEXT,
        str4 TEXT,
        str5 TEXT,
        str6 TEXT,
        str7 TEXT,
        str8 TEXT,
        str9 TEXT,
        str10 TEXT
    );`

	_, err := app.DB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating custom_table: %w", err)
	}

	return nil
}

func (app *App) EnsureCustomTableExistSP() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS custom_table (
        user_id TEXT PRIMARY KEY,
        promptstr TEXT NOT NULL,
        promptstr_stat INTEGER,
        str1 TEXT,
        str2 TEXT,
        str3 TEXT,
        str4 TEXT,
        str5 TEXT,
        str6 TEXT,
        str7 TEXT,
        str8 TEXT,
        str9 TEXT,
        str10 TEXT
    );`

	_, err := app.DB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating custom_table: %w", err)
	}

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
		return fmt.Errorf("error creating user_context table: %w", err)
	}

	// 为 conversation_id 创建索引
	createConvIDIndexSQL := `CREATE INDEX IF NOT EXISTS idx_user_context_conversation_id ON user_context(conversation_id);`

	_, err = app.DB.Exec(createConvIDIndexSQL)
	if err != nil {
		return fmt.Errorf("error creating index on user_context(conversation_id): %w", err)
	}

	// 为 parent_message_id 创建索引
	// 只有当您需要根据 parent_message_id 进行查询时才添加此索引
	createParentMsgIDIndexSQL := `CREATE INDEX IF NOT EXISTS idx_user_context_parent_message_id ON user_context(parent_message_id);`

	_, err = app.DB.Exec(createParentMsgIDIndexSQL)
	if err != nil {
		return fmt.Errorf("error creating index on user_context(parent_message_id): %w", err)
	}

	return nil
}

func (app *App) EnsureUserContextTableExistsSP() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS user_context (
        user_id TEXT PRIMARY KEY,
        conversation_id TEXT NOT NULL,
        parent_message_id TEXT
    );`

	_, err := app.DB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating user_context table: %w", err)
	}

	// 为 conversation_id 创建索引
	createConvIDIndexSQL := `CREATE INDEX IF NOT EXISTS idx_user_context_conversation_id ON user_context(conversation_id);`

	_, err = app.DB.Exec(createConvIDIndexSQL)
	if err != nil {
		return fmt.Errorf("error creating index on user_context(conversation_id): %w", err)
	}

	// 为 parent_message_id 创建索引
	// 只有当您需要根据 parent_message_id 进行查询时才添加此索引
	createParentMsgIDIndexSQL := `CREATE INDEX IF NOT EXISTS idx_user_context_parent_message_id ON user_context(parent_message_id);`

	_, err = app.DB.Exec(createParentMsgIDIndexSQL)
	if err != nil {
		return fmt.Errorf("error creating index on user_context(parent_message_id): %w", err)
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

func (app *App) handleUserContextSP(userID string) (string, string, error) {
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

func (app *App) migrateUserToNewContextSP(userID string) error {
	// 生成新的conversationID
	newConversationID := utils.GenerateUUID() // 假设GenerateUUID()是一个生成UUID的函数

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

func (app *App) updateUserContextSP(userID string, parentMessageID string) error {
	updateQuery := `UPDATE user_context SET parent_message_id = ? WHERE user_id = ?`
	_, err := app.DB.Exec(updateQuery, parentMessageID, userID)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) updateUserContextPro(userID int64, conversationID, parentMessageID string) error {
	updateQuery := `
    UPDATE user_context
    SET conversation_id = ?, parent_message_id = ?
    WHERE user_id = ?;`
	_, err := app.DB.Exec(updateQuery, conversationID, parentMessageID, userID)
	if err != nil {
		return fmt.Errorf("error updating user context: %w", err)
	}
	return nil
}

func (app *App) updateUserContextProSP(userID string, conversationID, parentMessageID string) error {
	updateQuery := `
    UPDATE user_context
    SET conversation_id = ?, parent_message_id = ?
    WHERE user_id = ?;`
	_, err := app.DB.Exec(updateQuery, conversationID, parentMessageID, userID)
	if err != nil {
		return fmt.Errorf("error updating user context: %w", err)
	}
	return nil
}

func (app *App) getHistory(conversationID, parentMessageID string) ([]structs.Message, error) {
	// 如果不开启上下文
	if config.GetNoContext() {
		return nil, nil
	}
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
		//fmtf.Printf("加入:%v\n", historyEntry)
		history = append(history, historyEntry)
	}
	return history, nil
}

// 记忆表
func (app *App) EnsureUserMemoriesTableExists() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS user_memories (
        memory_id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        conversation_id TEXT NOT NULL,
        parent_message_id TEXT,
        conversation_title TEXT NOT NULL
    );`

	_, err := app.DB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating user_memories table: %w", err)
	}

	createUserIDIndexSQL := `CREATE INDEX IF NOT EXISTS idx_user_memories_user_id ON user_memories(user_id);`
	_, err = app.DB.Exec(createUserIDIndexSQL)
	if err != nil {
		return fmt.Errorf("error creating index on user_memories(user_id): %w", err)
	}

	createConvDetailsIndexSQL := `CREATE INDEX IF NOT EXISTS idx_user_memories_conversation_details ON user_memories(conversation_id, parent_message_id);`
	_, err = app.DB.Exec(createConvDetailsIndexSQL)
	if err != nil {
		return fmt.Errorf("error creating index on user_memories(conversation_id, parent_message_id): %w", err)
	}

	return nil
}

func (app *App) AddUserMemory(userID int64, conversationID, parentMessageID, conversationTitle string) error {
	// 插入新的记忆
	insertMemorySQL := `
    INSERT INTO user_memories (user_id, conversation_id, parent_message_id, conversation_title)
    VALUES (?, ?, ?, ?);`
	_, err := app.DB.Exec(insertMemorySQL, userID, conversationID, parentMessageID, conversationTitle)
	if err != nil {
		return fmt.Errorf("error inserting new memory: %w", err)
	}

	// 检查并保持记忆数量不超过10条
	return app.ensureMemoryLimit(userID)
}

func (app *App) AddUserMemorySP(userID string, conversationID, parentMessageID, conversationTitle string) error {
	// 插入新的记忆
	insertMemorySQL := `
    INSERT INTO user_memories (user_id, conversation_id, parent_message_id, conversation_title)
    VALUES (?, ?, ?, ?);`
	_, err := app.DB.Exec(insertMemorySQL, userID, conversationID, parentMessageID, conversationTitle)
	if err != nil {
		return fmt.Errorf("error inserting new memory: %w", err)
	}

	// 检查并保持记忆数量不超过10条
	return app.ensureMemoryLimitSP(userID)
}

func (app *App) updateConversationTitle(userID int64, conversationID, parentMessageID, newTitle string) error {
	// 定义SQL更新语句
	updateQuery := `
    UPDATE user_memories
    SET conversation_title = ?
    WHERE user_id = ? AND conversation_id = ? AND parent_message_id = ?;`

	// 执行SQL更新操作
	_, err := app.DB.Exec(updateQuery, newTitle, userID, conversationID, parentMessageID)
	if err != nil {
		return fmt.Errorf("error updating conversation title: %w", err)
	}

	return nil
}

func (app *App) updateConversationTitleSP(userID string, conversationID, parentMessageID, newTitle string) error {
	// 定义SQL更新语句
	updateQuery := `
    UPDATE user_memories
    SET conversation_title = ?
    WHERE user_id = ? AND conversation_id = ? AND parent_message_id = ?;`

	// 执行SQL更新操作
	_, err := app.DB.Exec(updateQuery, newTitle, userID, conversationID, parentMessageID)
	if err != nil {
		return fmt.Errorf("error updating conversation title: %w", err)
	}

	return nil
}

func (app *App) ensureMemoryLimit(userID int64) error {
	// 查询当前记忆总数
	countQuerySQL := `SELECT COUNT(*) FROM user_memories WHERE user_id = ?;`
	var count int
	row := app.DB.QueryRow(countQuerySQL, userID)
	err := row.Scan(&count)
	if err != nil {
		return fmt.Errorf("error counting memories: %w", err)
	}

	// 如果记忆超过5条，则删除最旧的记忆
	if count > 5 {
		deleteOldestMemorySQL := `
        DELETE FROM user_memories
        WHERE memory_id IN (
            SELECT memory_id FROM user_memories
            WHERE user_id = ?
            ORDER BY memory_id ASC
            LIMIT ?
        );`
		_, err := app.DB.Exec(deleteOldestMemorySQL, userID, count-5)
		if err != nil {
			return fmt.Errorf("error deleting old memories: %w", err)
		}
	}

	return nil
}

func (app *App) ensureMemoryLimitSP(userID string) error {
	// 查询当前记忆总数
	countQuerySQL := `SELECT COUNT(*) FROM user_memories WHERE user_id = ?;`
	var count int
	row := app.DB.QueryRow(countQuerySQL, userID)
	err := row.Scan(&count)
	if err != nil {
		return fmt.Errorf("error counting memories: %w", err)
	}

	// 如果记忆超过5条，则删除最旧的记忆
	if count > 5 {
		deleteOldestMemorySQL := `
        DELETE FROM user_memories
        WHERE memory_id IN (
            SELECT memory_id FROM user_memories
            WHERE user_id = ?
            ORDER BY memory_id ASC
            LIMIT ?
        );`
		_, err := app.DB.Exec(deleteOldestMemorySQL, userID, count-5)
		if err != nil {
			return fmt.Errorf("error deleting old memories: %w", err)
		}
	}

	return nil
}

func (app *App) GetUserMemories(userID int64) ([]structs.Memory, error) {
	// 定义查询SQL，获取所有相关的记忆
	querySQL := `
    SELECT conversation_id, parent_message_id, conversation_title
    FROM user_memories
    WHERE user_id = ?;
    `
	rows, err := app.DB.Query(querySQL, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying user memories: %w", err)
	}
	defer rows.Close() // 确保关闭rows以释放数据库资源

	var memories []structs.Memory
	for rows.Next() {
		var m structs.Memory
		if err := rows.Scan(&m.ConversationID, &m.ParentMessageID, &m.ConversationTitle); err != nil {
			return nil, fmt.Errorf("error scanning memory: %w", err)
		}
		memories = append(memories, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return memories, nil
}

func (app *App) GetUserMemoriesSP(userID string) ([]structs.Memory, error) {
	// 定义查询SQL，获取所有相关的记忆
	querySQL := `
    SELECT conversation_id, parent_message_id, conversation_title
    FROM user_memories
    WHERE user_id = ?;
    `
	rows, err := app.DB.Query(querySQL, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying user memories: %w", err)
	}
	defer rows.Close() // 确保关闭rows以释放数据库资源

	var memories []structs.Memory
	for rows.Next() {
		var m structs.Memory
		if err := rows.Scan(&m.ConversationID, &m.ParentMessageID, &m.ConversationTitle); err != nil {
			return nil, fmt.Errorf("error scanning memory: %w", err)
		}
		memories = append(memories, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return memories, nil
}
