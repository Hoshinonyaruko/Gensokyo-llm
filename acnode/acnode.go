package acnode

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"unicode/utf16"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
)

// 定义包级别的全局变量
var ac *AhoCorasick
var wac *AhoCorasick

// init函数用于初始化操作
func init() {
	ac = NewAhoCorasick()
	wac = NewAhoCorasick()

	// 载入敏感词库
	if err := loadWordsIntoAC(ac, "sensitive_words.txt"); err != nil {
		log.Fatalf("初始化敏感词库失败：%v", err)
		// 注意，log.Fatalf会调用os.Exit(1)终止程序，因此后面的return不是必须的
	}

	// 载入白名单词库
	if err := loadWordsIntoAC(wac, "white.txt"); err != nil {
		log.Fatalf("初始化白名单词库失败：%v", err)
		// 同上，这里的return也不是必须的
	}

	// 移除了启动HTTP服务器的代码
}

type ACNode struct {
	children    map[rune]*ACNode
	fail        *ACNode
	isEnd       bool
	length      int
	replaceText string // 添加替换文本字段
}

type AhoCorasick struct {
	root *ACNode
}

// Replacement结构体来记录替换信息
type Replacement struct {
	Start int    // 替换起始位置
	End   int    // 替换结束位置
	Text  string // 替换文本
}

func NewAhoCorasick() *AhoCorasick {
	return &AhoCorasick{
		root: &ACNode{children: make(map[rune]*ACNode)},
	}
}

func (ac *AhoCorasick) Insert(word, replaceText string) {
	node := ac.root
	for _, ch := range word {
		if _, ok := node.children[ch]; !ok {
			node.children[ch] = &ACNode{children: make(map[rune]*ACNode)}
		}
		node = node.children[ch]
	}
	node.isEnd = true
	node.length = len([]rune(word))
	node.replaceText = replaceText // 存储替换文本
}

func (ac *AhoCorasick) BuildFailPointer() {
	queue := []*ACNode{ac.root}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for ch, child := range current.children {
			if current == ac.root {
				child.fail = ac.root
			} else {
				fail := current.fail
				for fail != nil {
					if next, ok := fail.children[ch]; ok {
						child.fail = next
						break
					}
					fail = fail.fail
				}
				if fail == nil {
					child.fail = ac.root
				}
			}
			queue = append(queue, child)
		}
	}
}

func (ac *AhoCorasick) FilterWithWhitelist(text string, whiteListedPositions []Position) string {
	node := ac.root
	runes := []rune(text)
	changes := false // 标记是否有替换发生

	// 创建一个替换列表，用于记录所有替换操作
	var replacements []Replacement

	for i, ch := range runes {
		for node != ac.root && node.children[ch] == nil {
			node = node.fail
		}
		if next, ok := node.children[ch]; ok {
			node = next
		}

		tmp := node
		for tmp != ac.root {
			if tmp.isEnd {
				isInWhiteList := false
				for _, pos := range whiteListedPositions {
					if i-pos.Start+1 >= tmp.length && i <= pos.End {
						isInWhiteList = true
						break
					}
				}

				if !isInWhiteList {
					start := i - tmp.length + 1
					replacements = append(replacements, Replacement{
						Start: start,
						End:   i,
						Text:  tmp.replaceText, // 使用节点存储的替换文本
					})
					changes = true
					break // 找到匹配，退出循环
				}
			}
			tmp = tmp.fail
		}
	}

	// 使用applyReplacements函数替换原有的替换逻辑
	if changes {
		newText := applyReplacements(text, replacements)
		return newText
	}
	return text
}

// 假设Replacement定义如前所述

// Step 1: 合并重叠替换
func mergeOverlappingReplacements(replacements []Replacement) []Replacement {
	if len(replacements) == 0 {
		return replacements
	}

	// 按Start排序
	sort.Slice(replacements, func(i, j int) bool {
		if replacements[i].Start == replacements[j].Start {
			return replacements[i].End > replacements[j].End // 如果Start相同，更长的在前
		}
		return replacements[i].Start < replacements[j].Start
	})

	merged := []Replacement{replacements[0]}
	for _, current := range replacements[1:] {
		last := &merged[len(merged)-1]
		if current.Start <= last.End { // 检查重叠
			if current.End > last.End {
				last.End = current.End   // 扩展当前项以包括重叠
				last.Text = current.Text // 假设新的替换文本更优先
			}
		} else {
			merged = append(merged, current)
		}
	}
	return merged
}

// Step 2 & 3: 实施替换
func applyReplacements(text string, replacements []Replacement) string {
	runes := []rune(text)
	var result []rune
	lastIndex := 0
	for _, r := range mergeOverlappingReplacements(replacements) {
		// 添加未被替换的部分
		result = append(result, runes[lastIndex:r.Start]...)
		// 添加替换文本
		result = append(result, []rune(r.Text)...)
		lastIndex = r.End + 1
	}
	// 添加最后一部分未被替换的文本
	result = append(result, runes[lastIndex:]...)
	return string(result)
}

type Position struct {
	Start int
	End   int
}

func (wac *AhoCorasick) MatchPositions(text string) []Position {
	node := wac.root
	runes := []rune(text)
	positions := []Position{} // 用于储存匹配到的白名单词的位置

	//log.Printf("开始匹配白名单文本：%s", text)

	for i, ch := range runes {
		for node != wac.root && node.children[ch] == nil {
			node = node.fail
		}

		if next, ok := node.children[ch]; ok {
			node = next
		}

		tmp := node
		for tmp != wac.root {
			if tmp.isEnd {
				//log.Printf("找到白名单匹配词结束点，位于索引：%d，匹配词长度：%d", i, tmp.length)

				startPos := i - tmp.length + 1
				endPos := i
				positions = append(positions, Position{Start: startPos, End: endPos})

			}
			tmp = tmp.fail
		}
	}

	//log.Printf("匹配到的位置：%v", positions)
	return positions
}

func loadWordsIntoAC(ac *AhoCorasick, filename string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// 如果文件不存在，则创建一个空文件
		file, err := os.Create(filename)
		if err != nil {
			return fmtf.Errorf("failed to create the file: %v", err)
		}
		file.Close() // 创建后立即关闭文件，因为下面会再次打开它用于读写
	}
	// 打开原文件
	file, err := os.Open(filename)
	if err != nil {
		return fmtf.Errorf("failed to open the sensitive words file: %v", err)
	}
	defer file.Close()

	// 创建一个临时的buffer来存储修改后的内容
	var buffer bytes.Buffer

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "####")
		word := parts[0]
		DefaultChangeWord := config.GetDefaultChangeWord()
		replaceText := DefaultChangeWord // 默认替换文本
		if len(parts) > 1 && parts[1] != "" {
			replaceText = parts[1] // 使用指定的替换文本
		} else {
			// 如果不存在####~，则添加
			line = word + "####" + DefaultChangeWord
		}

		// 将修改后的行写入buffer
		buffer.WriteString(line + "\n")

		// 插入到AC Trie中
		ac.Insert(word, replaceText)

		// 对于Unicode转义的处理，可能需要根据实际情况调整
		unicodeWord := convertToUnicodeEscape(word)
		ac.Insert(unicodeWord, replaceText)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// 构建失败指针
	ac.BuildFailPointer()

	// 将buffer中的内容写回到原文件或新文件中
	// 如果要覆盖原文件，请先关闭原文件
	file.Close()                                       // 关闭原文件以便覆盖
	err = os.WriteFile(filename, buffer.Bytes(), 0644) // 覆盖原文件
	if err != nil {
		return fmtf.Errorf("failed to write back to the sensitive words file: %v", err)
	}

	return nil
}

// 将字符串转换为其Unicode转义序列表示形式
func convertToUnicodeEscape(str string) string {
	runes := []rune(str)
	utf16Runes := utf16.Encode(runes)
	var unicodeEscapeBuilder strings.Builder

	for _, r := range utf16Runes {
		unicodeEscapeBuilder.WriteString(fmt.Sprintf("\\u%04x", r))
	}

	return unicodeEscapeBuilder.String()
}

// 改写后的函数，接受word参数，并返回处理结果
func CheckWord(word string) string {
	if word == "" {
		log.Println("错误请求：缺少 'word' 参数")
		return "错误：缺少 'word' 参数"
	}

	if len([]rune(word)) > 5000 {
		if strings.Contains(word, "[CQ:image,file=base64://") {
			// 当word包含特定字符串时原样返回
			fmtf.Printf("原样返回的文本：%s", word)
			return word
		}
		log.Printf("错误请求：字符数超过最大限制（5000字符）。内容：%s", word)
		return "错误：字符数超过最大限制（5000字符）"
	}

	// 使用全局的wac进行白名单匹配
	whiteListedPositions := wac.MatchPositions(word)

	// 使用全局的ac进行过滤，并结合白名单
	result := ac.FilterWithWhitelist(word, whiteListedPositions)

	return result
}
