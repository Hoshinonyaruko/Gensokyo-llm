package applogic

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
	"github.com/hoshinonyaruko/gensokyo-llm/utils"
)

// 将向量二值化和对应文本并存储到数据库
// insertVectorData插入向量数据并返回新插入行的ID
func (app *App) insertVectorDataSensitive(text string, vector []float64) (int64, error) {
	binaryVector := vectorToBinaryConcurrent(vector) // 使用二值化函数

	var sum float64
	for _, v := range vector {
		sum += v * v
	}
	norm := math.Sqrt(sum)
	n := config.GetCacheN()
	k := config.GetCacheK()
	// 先进行四舍五入，然后转换为int64
	l := int64(math.Round(norm * k))
	groupID := l % n
	if config.GetPrintHanming() {
		fmtf.Printf("(norm*k): %v\n", norm*k)
		fmtf.Printf("(norm*k) mod n ==== (%v) mod %v\n", l, n)
		fmtf.Printf("groupid : %v\n", groupID)
	}

	result, err := app.DB.Exec("INSERT INTO sensitive_words (text, vector, norm, group_id) VALUES (?, ?, ?, ?)", text, binaryVector, norm, groupID)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId() // 获取新插入行的ID
	if err != nil {
		return 0, err
	}

	return id, nil
}

// searchSimilarText函数根据汉明距离搜索数据库中与给定向量相似的文本
func (app *App) searchSimilarTextSensitive(vector []float64, threshold int, targetGroupID int64) ([]TextDistance, []int, error) {
	binaryVector := vectorToBinaryConcurrent(vector) // 二值化查询向量
	var results []TextDistance
	var ids []int

	rows, err := app.DB.Query("SELECT id, text, vector FROM sensitive_words WHERE group_id = ?", targetGroupID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var text string
		var dbVectorBytes []byte
		if err := rows.Scan(&id, &text, &dbVectorBytes); err != nil {
			continue
		}
		//fmtf.Printf("二值一,%v,二值二,%v", binaryVector, dbVectorBytes)
		distance := hammingDistanceOptimized(binaryVector, dbVectorBytes)
		if config.GetPrintHanming() {
			fmtf.Printf("匹配到敏感文本,%v,汉明距离,%v,当前阈值,%v\n", text, distance, threshold)
		}
		if distance <= threshold {
			results = append(results, TextDistance{Text: text, Distance: distance})
			ids = append(ids, id)
		}
	}

	// 根据汉明距离对结果进行排序，并保持ids数组与results数组的一致性
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Distance < results[j].Distance
	})

	// 由于我们按照results排序，ids数组也需要相应排序，以保持位置一致性
	sortedIds := make([]int, len(ids))
	for index := range results {
		sortedIds[index] = ids[index] // 注意：这里假设ids排序前后位置不变，因为results和ids是同时追加的
	}

	return results, sortedIds, nil
}

// searchForSingleVector函数根据单个向量搜索并返回按相似度排序的文本数组
func (app *App) searchForSingleVectorSensitive(vector []float64, threshold int) ([]string, []int, error) {
	// 计算目标组ID
	targetGroupID := calculateGroupID(vector)

	// 调用searchSimilarText函数进行搜索，现在它也返回匹配文本的ID数组
	textDistances, ids, err := app.searchSimilarTextSensitive(vector, threshold, targetGroupID)
	if err != nil {
		return nil, nil, err
	}

	// 从TextDistance数组中提取文本
	var similarTexts []string
	for _, td := range textDistances {
		similarTexts = append(similarTexts, td.Text)
	}

	// 返回相似的文本数组和对应的ID数组
	return similarTexts, ids, nil
}

func (app *App) ProcessSensitiveWords() error {
	// Step 1: 判断是否需要处理敏感词向量
	if !config.GetVectorSensitiveFilter() {
		fmt.Println("向量敏感词过滤未启用")
		return nil // 不需要处理敏感词向量
	}

	// Step 2: 读取敏感词列表
	file, err := os.Open("vector_sensitive.txt")
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，则创建文件
			_, err := os.Create("vector_sensitive.txt")
			if err != nil {
				return fmt.Errorf("创建 vector_sensitive.txt 文件时出错: %w", err)
			}
			// 文件被创建后，此次没有内容处理
			return nil
		}
		return fmt.Errorf("打开 vector_sensitive.txt 文件时出错: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()

		// 检查文本是否已存在于数据库中
		exists, err := app.textExistsInDatabase(text)
		if err != nil {
			return fmt.Errorf("查询数据库时出错: %w", err)
		}

		if exists {
			fmt.Printf("数据库中已存在敏感词：%s\n", text)
			continue
		}

		// 文本在数据库中不存在，计算其向量
		fmtf.Printf("计算向量的敏感词：%s\n", text)
		vector, err := app.CalculateTextEmbedding(text)
		if err != nil {
			return fmt.Errorf("计算文本向量时出错 '%s': %w", text, err)
		}

		// 将新的向量数据插入数据库
		id, err := app.insertVectorDataSensitive(text, vector)
		if err != nil {
			return fmt.Errorf("将敏感词向量数据插入数据库时出错: %w", err)
		}
		fmt.Printf("成功插入敏感词，ID为：%d\n", id)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("扫描 vector_sensitive.txt 文件时出错: %w", err)
	}

	return nil
}

func (app *App) ProcessSensitiveWordsV2() error {
	// Step 1: 判断是否需要处理敏感词向量
	if !config.GetVectorSensitiveFilter() {
		fmt.Println("向量敏感词过滤未启用")
		return nil // 不需要处理敏感词向量
	}

	// Step 2: 读取敏感词列表
	file, err := os.Open("vector_sensitive.txt")
	if err != nil {
		return fmt.Errorf("打开 vector_sensitive.txt 文件时出错: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()

		// 对每个敏感词重复计算向量10次
		for i := 0; i < 10; i++ {
			fmt.Printf("计算向量的敏感词：%s，尝试 #%d\n", text, i+1)
			vector, err := app.CalculateTextEmbedding(text)
			if err != nil {
				return fmt.Errorf("计算文本向量时出错 '%s': %w", text, err)
			}

			// 计算groupID
			var sum float64
			for _, v := range vector {
				sum += v * v
			}
			norm := math.Sqrt(sum)
			k := config.GetCacheK()
			l := int64(math.Round(norm * k))
			n := config.GetCacheN()
			groupID := l % n

			// 检查数据库中是否存在相同text和groupID的记录
			exists, err := app.textAndGroupIDExistsInDatabase(text, groupID)
			if err != nil {
				return fmt.Errorf("检查敏感词存在性时出错: %w", err)
			}

			if exists {
				fmt.Printf("数据库中已存在敏感词及分组ID：%s, %d\n", text, groupID)
				continue
			}

			// 将新的向量数据插入数据库
			id, err := app.insertVectorDataSensitive(text, vector)
			if err != nil {
				return fmt.Errorf("将敏感词向量数据插入数据库时出错: %w", err)
			}
			fmt.Printf("成功插入敏感词及分组ID，ID为：%d\n", id)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("扫描 vector_sensitive.txt 文件时出错: %w", err)
	}

	return nil
}

// 检查数据库中是否存在相同text和groupID的记录
func (app *App) textAndGroupIDExistsInDatabase(text string, groupID int64) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM sensitive_words WHERE text = ? AND group_id = ? LIMIT 1)"
	err := app.DB.QueryRow(query, text, groupID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("查询敏感词和分组ID时出错: %w", err)
	}
	return exists, nil
}

// textExistsInDatabase 检查给定的文本是否已存在于数据库中
func (app *App) textExistsInDatabase(text string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM sensitive_words WHERE text = ? LIMIT 1)`
	err := app.DB.QueryRow(query, text).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (app *App) InterceptSensitiveContent(vector []float64, message structs.OnebotGroupMessage, selfid string) (int, string, error) {
	// 自定义阈值
	Threshold := config.GetVertorSensitiveThreshold()

	// 进行搜索
	results, _, err := app.searchForSingleVectorSensitive(vector, Threshold)
	if err != nil {
		return 1, "", fmtf.Errorf("error searching for sensitive content: %w", err)
	}

	// 输出搜索到的result数组
	fmtf.Printf("Search results: %v\n", results)

	// 如果不为0，则表示匹配到了
	if len(results) > 0 {
		// 匹配到敏感内容
		fmt.Println("Sensitive content detected!")

		// 获取安全词响应
		saveresponse := config.GetRandomSaveResponse()

		// 根据消息类型和配置，决定如何响应
		if saveresponse != "" {
			if message.RealMessageType == "group_private" || message.MessageType == "private" {
				if !config.GetUsePrivateSSE() {
					utils.SendPrivateMessage(message.UserID, saveresponse, selfid)
				} else {
					utils.SendSSEPrivateSafeMessage(message.UserID, saveresponse)
				}
			} else {
				utils.SendGroupMessage(message.GroupID, message.UserID, saveresponse, selfid)
			}
			return 1, saveresponse, nil
		}
	} else {
		// 未匹配到敏感内容
		fmtf.Println("No sensitive content detected.")
	}

	return 0, "", nil
}
