package prompt

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/hoshinonyaruko/gensokyo-llm/structs"
	"gopkg.in/yaml.v3"
)

type Prompt struct {
	Role    string `yaml:"role"`
	Content string `yaml:"content"`
}

type PromptFile struct {
	Prompts  []Prompt         `yaml:"Prompt"`
	Settings structs.Settings `yaml:"settings"`
}

var (
	promptsCache = make(map[string]PromptFile)
	lock         sync.RWMutex
	promptsDir   = "prompts" // 定义固定的目录名
)

func init() {
	// 通过 init 函数在包加载时就执行目录监控
	err := LoadPrompts()
	if err != nil {
		log.Fatal("Failed to load prompts:", err)
	}
}

// LoadPrompts 确保目录存在并尝试加载提示词文件
func LoadPrompts() error {
	// 构建目录路径
	directory := filepath.Join(".", promptsDir)

	// 尝试创建目录（如果不存在）
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		// 目录不存在，尝试创建它
		if err := os.MkdirAll(directory, os.ModePerm); err != nil {
			return err
		}
	}
	files, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yml" {
			loadFile(filepath.Join(directory, file.Name()))
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					loadFile(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(directory)
	if err != nil {
		return err
	}

	return nil
}

func loadFile(filename string) {
	lock.Lock()
	defer lock.Unlock()

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Println("Failed to read file:", err)
		return
	}

	var prompts PromptFile
	err = yaml.Unmarshal(data, &prompts)
	if err != nil {
		log.Println("Failed to unmarshal YAML:", err)
		return
	}

	baseName := filepath.Base(filename)
	promptsCache[baseName] = prompts
}

func GetMessagesFromFilename(basename string) ([]structs.Message, error) {
	lock.RLock()
	defer lock.RUnlock()

	filename := basename + ".yml"
	promptFile, exists := promptsCache[filename]
	if !exists {
		return nil, fmt.Errorf("no data for file: %s", filename)
	}

	var history []structs.Message
	for _, prompt := range promptFile.Prompts {
		history = append(history, structs.Message{
			Text: prompt.Content,
			Role: prompt.Role,
		})
	}

	return history, nil
}

// FindFirstSystemMessage 从消息列表中查找第一条角色为 "system" 的消息
func FindFirstSystemMessage(history []structs.Message) (structs.Message, error) {
	lock.RLock()
	defer lock.RUnlock()

	for _, message := range history {
		if message.Role == "system" || message.Role == "System" {
			return message, nil
		}
	}

	return structs.Message{}, fmt.Errorf("no system message found in history")
}

// 返回除了 "system" 角色之外的所有消息
func GetMessagesExcludingSystem(basename string) ([]structs.Message, error) {
	lock.RLock()
	defer lock.RUnlock()

	filename := basename + ".yml"
	promptFile, exists := promptsCache[filename]
	if !exists {
		return nil, fmt.Errorf("no data for file: %s", filename)
	}

	var history []structs.Message
	for _, prompt := range promptFile.Prompts {
		if prompt.Role != "system" && prompt.Role != "System" {
			history = append(history, structs.Message{
				Text: prompt.Content,
				Role: prompt.Role,
			})
		}
	}

	return history, nil
}

// 返回第一条 "system" 角色的消息文本
func GetFirstSystemMessage(basename string) (string, error) {
	lock.RLock()
	defer lock.RUnlock()

	filename := basename + ".yml"
	promptFile, exists := promptsCache[filename]
	if !exists {
		return "", fmt.Errorf("no data for file: %s", filename)
	}

	for _, prompt := range promptFile.Prompts {
		if prompt.Role == "system" || prompt.Role == "System" {
			return prompt.Content, nil
		}
	}

	return "", fmt.Errorf("no system message found in file: %s", filename)
}

// GetSettingFromFilename 用于获取配置文件中的特定设置
func GetSettingFromFilename(basename, settingName string) (interface{}, error) {
	lock.RLock()
	defer lock.RUnlock()

	filename := basename + ".yml"
	promptFile, exists := promptsCache[filename]
	if !exists {
		return nil, fmt.Errorf("no data for file: %s", filename)
	}

	// 使用反射获取Settings结构体中的字段
	rv := reflect.ValueOf(promptFile.Settings)
	field := rv.FieldByName(settingName)
	if !field.IsValid() {
		return nil, fmt.Errorf("no setting with name: %s", settingName)
	}

	// 返回字段的值，转换为interface{}
	return field.Interface(), nil
}
