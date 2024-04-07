package applogic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"runtime"
	"sort"
	"sync"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/hunyuan"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
)

type TextDistance struct {
	Text     string
	Distance int
}

func (app *App) CalculateTextEmbedding(text string) ([]float64, error) {
	embeddingType := config.GetEmbeddingType()
	switch embeddingType {
	case 0:
		return app.CalculateTextEmbeddingHunyuan(text)
	case 1:
		// 从头构造请求到其他API的接口
		apiURL := config.GetWenxinEmbeddingUrl()
		accessToken := config.GetWenxinAccessToken()

		// 构建请求URL
		url := fmt.Sprintf("%s?access_token=%s", apiURL, accessToken)

		// 构建请求负载
		payload := map[string]interface{}{
			"input": []string{text},
			// 可以添加其他必要的字段
		}
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("error marshaling payload: %v", err)
		}

		// 创建并发送POST请求
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error sending request: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %v", err)
		}

		// 解析响应数据
		var response structs.EmbeddingResponseErnie
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("error decoding response: %v", err)
		}

		// 提取embedding向量
		var embedding []float64
		for _, data := range response.Data {
			embedding = append(embedding, data.Embedding...)
		}

		if config.GetPrintVector() {
			fmt.Printf("百度返回的向量:%v\n", embedding)
		}

		return embedding, nil
	default:
		return nil, fmt.Errorf("unsupported embedding type: %d", embeddingType)
	}
}

// CalculateTextEmbedding 调用混元-Embedding接口将文本转换为向量表示。
func (app *App) CalculateTextEmbeddingHunyuan(text string) ([]float64, error) {

	// 实例化一个请求对象
	request := hunyuan.NewGetEmbeddingRequest()
	// 这里根据接口要求设置参数，例如文本内容
	request.Input = &text

	// 调用接口
	response, err := app.Client.GetEmbedding(request)
	if err != nil {
		return nil, err
	}

	// 处理返回的embedding数据
	var embedding []float64
	for _, data := range response.Response.Data {
		if data.Embedding != nil {
			for _, value := range data.Embedding {
				if value != nil {
					embedding = append(embedding, *value)
				}
			}
		}
	}

	if config.GetPrintVector() {
		fmtf.Printf("混元返回的向量:%v\n", embedding)
	}

	return embedding, nil
}

// GetRandomAnswer 根据问题文本随机获取一个答案
func (app *App) GetRandomAnswer(questionText string) (string, error) {
	var answerText string
	// 首先获取问题的ID
	var questionID int
	queryForID := `SELECT id FROM questions WHERE question_text = ?;`
	err := app.DB.QueryRow(queryForID, questionText).Scan(&questionID)
	if err != nil {
		return "", err // 可能是因为没有找到对应的问题
	}

	// 使用问题ID在qa_cache表中随机选择一个答案
	queryForAnswer := `SELECT answer_text FROM qa_cache WHERE question_id = ? ORDER BY RANDOM() LIMIT 1;`
	err = app.DB.QueryRow(queryForAnswer, questionID).Scan(&answerText)
	if err != nil {
		return "", err // 可能是因为没有找到对应的答案
	}
	return answerText, nil
}

// InsertQAEntry 将新的问题和答案对插入到数据库中
func (app *App) InsertQAEntry(questionText, answerText string, vectorDataID int) error {
	// 检查问题是否已存在，并获取问题的ID
	var questionID int
	queryForID := `SELECT id FROM questions WHERE question_text = ?;`
	err := app.DB.QueryRow(queryForID, questionText).Scan(&questionID)

	// 如果问题不存在，则插入新问题
	if err != nil {
		insertQuestionQuery := `INSERT INTO questions (question_text, vector_data_id) VALUES (?, ?);`
		result, err := app.DB.Exec(insertQuestionQuery, questionText, vectorDataID)
		if err != nil {
			return err // 插入问题失败
		}
		questionID64, err := result.LastInsertId()
		if err != nil {
			return err // 获取插入问题的ID失败
		}
		questionID = int(questionID64)
	}

	// 插入答案到qa_cache表中
	insertAnswerQuery := `INSERT INTO qa_cache (answer_text, question_id) VALUES (?, ?);`
	_, err = app.DB.Exec(insertAnswerQuery, answerText, questionID)
	if err != nil {
		return err // 插入答案失败
	}
	return nil
}

// 二进制向量处理
func vectorToBinaryConcurrent(vector []float64) []byte {
	var wg sync.WaitGroup
	segmentSize := (len(vector) + runtime.NumCPU() - 1) / runtime.NumCPU() // 确保向量被均匀分割
	binaryVector := make([]byte, len(vector))
	vtb := config.GetVToBThreshold()

	for i := 0; i < len(vector); i += segmentSize {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			end := start + segmentSize
			if end > len(vector) {
				end = len(vector)
			}
			//多携程向量快速二值化
			for j := start; j < end; j++ {
				if vector[j] >= vtb {
					binaryVector[j] = 1
				} else {
					binaryVector[j] = 0
				}
			}
		}(i)
	}

	wg.Wait()
	return binaryVector
}

// 计算两个二进制向量之间的汉明距离
// 如果向量长度不同，则比较它们的共同长度部分，并忽略较长向量的剩余部分。
func hammingDistanceOptimized(vecA, vecB []byte) int {
	distance := 0
	minLength := len(vecA)
	if len(vecB) < minLength {
		minLength = len(vecB)
	}
	for i := 0; i < minLength; i++ {
		xor := vecA[i] ^ vecB[i] // 对应位不同时，结果为1
		// 计算xor中1的个数，即汉明距离的一部分
		for xor != 0 {
			distance++
			xor &= xor - 1 // 清除最低位的1
		}
	}
	return distance
}

// 将向量二值化和对应文本并存储到数据库
// insertVectorData插入向量数据并返回新插入行的ID
func (app *App) insertVectorData(text string, vector []float64) (int64, error) {
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

	result, err := app.DB.Exec("INSERT INTO vector_data (text, vector, norm, group_id) VALUES (?, ?, ?, ?)", text, binaryVector, norm, groupID)
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
func (app *App) searchSimilarText(vector []float64, threshold int, targetGroupID int64) ([]TextDistance, []int, error) {
	binaryVector := vectorToBinaryConcurrent(vector) // 二值化查询向量
	var results []TextDistance
	var ids []int

	rows, err := app.DB.Query("SELECT id, text, vector FROM vector_data WHERE group_id = ?", targetGroupID)
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
			fmtf.Printf("匹配到文本,%v,汉明距离,%v,当前阈值,%v\n", text, distance, threshold)
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

func calculateGroupID(vector []float64) int64 {
	var sum float64
	for _, v := range vector {
		sum += v * v
	}
	norm := math.Sqrt(sum)
	k := config.GetCacheK()
	n := config.GetCacheN()
	// 先进行四舍五入，然后转换为int64
	l := int64(math.Round(norm * k))
	groupid := l % n // 通过范数计算出一个整数，并将其模n来分配到一个组
	if config.GetPrintHanming() {
		fmtf.Printf("(norm*k): %v\n", norm*k)
		fmtf.Printf("(norm*k) mod n ==== (%v) mod %v\n", l, n)
		fmtf.Printf("groupid : %v\n", groupid)
	}
	return groupid
}

// searchForSingleVector函数根据单个向量搜索并返回按相似度排序的文本数组
func (app *App) searchForSingleVector(vector []float64, threshold int) ([]string, []int, error) {
	// 计算目标组ID
	targetGroupID := calculateGroupID(vector)

	// 调用searchSimilarText函数进行搜索，现在它也返回匹配文本的ID数组
	textDistances, ids, err := app.searchSimilarText(vector, threshold, targetGroupID)
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
