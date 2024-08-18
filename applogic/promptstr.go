package applogic

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
)

func (app *App) InsertCustomTableRecord(userID int64, promptStr string, promptStrStat int, strs ...string) error {
	// 构建 SQL 语句，使用 UPSERT 逻辑
	sqlStr := `
    INSERT INTO custom_table (user_id, promptstr, promptstr_stat, str1, str2, str3, str4, str5, str6, str7, str8, str9, str10)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(user_id) DO UPDATE SET 
        promptstr = excluded.promptstr,
        promptstr_stat = excluded.promptstr_stat`

	// 为每个非nil str构建更新部分
	updateParts := make([]string, 10)
	params := make([]interface{}, 13)
	params[0] = userID
	params[1] = promptStr
	params[2] = promptStrStat

	for i, str := range strs {
		if i < 10 {
			params[i+3] = str
			if str != "" { // 只更新非空的str字段
				fieldName := fmt.Sprintf("str%d", i+1)
				updateParts[i] = fmt.Sprintf("%s = excluded.%s", fieldName, fieldName)
			}
		}
	}

	// 添加非空更新字段到SQL语句
	nonEmptyUpdates := []string{}
	for _, part := range updateParts {
		if part != "" {
			nonEmptyUpdates = append(nonEmptyUpdates, part)
		}
	}
	if len(nonEmptyUpdates) > 0 {
		sqlStr += ", " + strings.Join(nonEmptyUpdates, ", ")
	}

	sqlStr += ";" // 结束 SQL 语句

	// 填充剩余的nil值
	for j := len(strs) + 3; j < 13; j++ {
		params[j] = nil
	}

	// 执行 SQL 操作
	_, err := app.DB.Exec(sqlStr, params...)
	if err != nil {
		return fmt.Errorf("error inserting or updating record in custom_table: %w", err)
	}

	return nil
}

func (app *App) FetchCustomRecord(userID int64, fields ...string) (*structs.CustomRecord, error) {
	// Default fields now include promptstr_stat
	queryFields := "user_id, promptstr, promptstr_stat"
	if len(fields) > 0 {
		queryFields += ", " + strings.Join(fields, ", ")
	}

	// Construct the SQL query string
	queryStr := fmt.Sprintf("SELECT %s FROM custom_table WHERE user_id = ?", queryFields)

	row := app.DB.QueryRow(queryStr, userID)
	var record structs.CustomRecord
	// Initialize scan parameters including the new promptstr_stat
	scanArgs := []interface{}{&record.UserID, &record.PromptStr, &record.PromptStrStat}
	for i := 0; i < len(fields); i++ {
		idx := fieldIndex(fields[i])
		if idx >= 0 {
			scanArgs = append(scanArgs, &record.Strs[idx])
		}
	}

	err := row.Scan(scanArgs...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No record found
		}
		return nil, fmt.Errorf("error scanning custom_table record: %w", err)
	}

	return &record, nil
}

func (app *App) deleteCustomRecord(userID int64) error {
	deleteSQL := `DELETE FROM custom_table WHERE user_id = ?;`

	_, err := app.DB.Exec(deleteSQL, userID)
	if err != nil {
		return fmt.Errorf("error deleting record from custom_table: %w", err)
	}

	return nil
}

func (app *App) deleteCustomRecordSP(userID string) error {
	deleteSQL := `DELETE FROM custom_table WHERE user_id = ?;`

	_, err := app.DB.Exec(deleteSQL, userID)
	if err != nil {
		return fmt.Errorf("error deleting record from custom_table: %w", err)
	}

	return nil
}

// Helper function to get index from field name
func fieldIndex(field string) int {
	if strings.HasPrefix(field, "str") && len(field) > 3 {
		if idx, err := strconv.Atoi(field[3:]); err == nil && idx >= 1 && idx <= 10 {
			return idx - 1
		}
	}
	return -1
}

func (app *App) ProcessPromptMarks(userID int64, QorA string, promptStr *string) {

	// 获取 PromptMarks
	PromptMarks := config.GetPromptMarks(*promptStr)
	maxMatchCount := 0
	bestPromptStr := ""
	bestPromptMarksLength := 0

	for _, mark := range PromptMarks {
		// 如果没有设置keyword则不处理
		if len(mark.Keywords) != 0 {
			// 检查 QorA 是否包含 Keywords 中的任意一个成员
			matchCount := 0
			for _, keyword := range mark.Keywords {
				if strings.Contains(QorA, keyword) {
					matchCount++
				}
			}

			// 更新找到含有最多匹配项的新 promptStr
			if matchCount > maxMatchCount {
				maxMatchCount = matchCount
				bestPromptStr = mark.BranchName
				bestPromptMarksLength = config.GetPromptMarksLength(bestPromptStr)
			}
		}
	}
	// 如果找到有效的匹配，则插入记录
	if maxMatchCount > 0 {
		err := app.InsertCustomTableRecord(userID, bestPromptStr, bestPromptMarksLength)
		if err != nil {
			fmt.Println("Error inserting custom table record:", err)
			return
		}
		// 输出结果
		fmt.Printf("type1=流转prompt参数: %s, newPromptStrStat: %d\n", bestPromptStr, bestPromptMarksLength)
		*promptStr = bestPromptStr
	}
}
