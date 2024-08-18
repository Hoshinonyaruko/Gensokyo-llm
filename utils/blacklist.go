package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
)

var blacklist = make(map[string]bool)
var mu sync.RWMutex

// LoadBlacklist 从给定的文件路径载入黑名单ID。
// 如果文件不存在，则创建该文件。
func LoadBlacklist(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 如果文件不存在，则创建一个新文件
			file, err = os.Create(filePath)
			if err != nil {
				return err // 创建文件失败，返回错误
			}
		} else {
			return err // 打开文件失败，且原因不是文件不存在
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	mu.Lock()
	defer mu.Unlock()
	blacklist = make(map[string]bool) // 重置黑名单

	for scanner.Scan() {
		blacklist[scanner.Text()] = true
	}

	return scanner.Err()
}

// isInBlacklist 检查给定的ID是否在黑名单中。
func IsInBlacklist(id string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, exists := blacklist[id]
	return exists
}

// watchBlacklist 监控黑名单文件的变动并动态更新。
func WatchBlacklist(filePath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error creating watcher:", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("Detected update to blacklist, reloading...")
					err := LoadBlacklist(filePath)
					if err != nil {
						log.Printf("Error reloading blacklist: %v", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()

	err = watcher.Add(filePath)
	if err != nil {
		log.Fatal("Error adding watcher to file:", err)
	}
	<-done // Keep the watcher alive
}

// BlacklistIntercept 检查用户ID是否在黑名单中，如果在，则发送预设消息
func BlacklistIntercept(message structs.OnebotGroupMessage, selfid string, promptstr string) bool {
	// 检查群ID是否在黑名单中
	if IsInBlacklist(strconv.FormatInt(message.GroupID, 10)) {
		// 获取黑名单响应消息
		responseMessage := config.GetBlacklistResponseMessages()

		// 根据消息类型发送响应
		if message.RealMessageType == "group_private" || message.MessageType == "private" {
			if !config.GetUsePrivateSSE() {
				SendPrivateMessage(message.UserID, responseMessage, selfid, promptstr)
			} else {
				SendSSEPrivateMessage(message.UserID, responseMessage, promptstr, selfid)
			}
		} else {
			SendGroupMessage(message.GroupID, message.UserID, responseMessage, selfid, promptstr)
		}

		fmt.Printf("groupid:[%v]这个群在黑名单中,被拦截\n", message.GroupID)
		return true // 拦截
	}

	// 检查用户ID是否在黑名单中
	if IsInBlacklist(strconv.FormatInt(message.UserID, 10)) {
		// 获取黑名单响应消息
		responseMessage := config.GetBlacklistResponseMessages()

		// 根据消息类型发送响应
		if message.RealMessageType == "group_private" || message.MessageType == "private" {
			if !config.GetUsePrivateSSE() {
				SendPrivateMessage(message.UserID, responseMessage, selfid, promptstr)
			} else {
				SendSSEPrivateMessage(message.UserID, responseMessage, promptstr, selfid)
			}
		} else {
			SendGroupMessage(message.GroupID, message.UserID, responseMessage, selfid, promptstr)
		}

		fmt.Printf("userid:[%v]这位用户在黑名单中,被拦截\n", message.UserID)
		return true // 拦截
	}

	return false // 用户ID不在黑名单中，不拦截
}

// BlacklistIntercept 检查用户ID是否在黑名单中，如果在，则发送预设消息
func BlacklistInterceptSP(message structs.OnebotGroupMessageS, selfid string, promptstr string) bool {
	// 检查群ID是否在黑名单中
	if IsInBlacklist(message.GroupID) {
		// 获取黑名单响应消息
		responseMessage := config.GetBlacklistResponseMessages()

		// 根据消息类型发送响应
		if message.RealMessageType == "group_private" || message.MessageType == "private" {
			if !config.GetUsePrivateSSE() {
				SendPrivateMessageSP(message.UserID, responseMessage, selfid, promptstr)
			} else {
				SendSSEPrivateMessageSP(message.UserID, responseMessage, promptstr, selfid)
			}
		} else {
			SendGroupMessageSP(message.GroupID, message.UserID, responseMessage, selfid, promptstr)
		}

		fmt.Printf("groupid:[%v]这个群在黑名单中,被拦截\n", message.GroupID)
		return true // 拦截
	}

	// 检查用户ID是否在黑名单中
	if IsInBlacklist(message.UserID) {
		// 获取黑名单响应消息
		responseMessage := config.GetBlacklistResponseMessages()

		// 根据消息类型发送响应
		if message.RealMessageType == "group_private" || message.MessageType == "private" {
			if !config.GetUsePrivateSSE() {
				SendPrivateMessageSP(message.UserID, responseMessage, selfid, promptstr)
			} else {
				SendSSEPrivateMessageSP(message.UserID, responseMessage, promptstr, selfid)
			}
		} else {
			SendGroupMessageSP(message.GroupID, message.UserID, responseMessage, selfid, promptstr)
		}

		fmt.Printf("userid:[%v]这位用户在黑名单中,被拦截\n", message.UserID)
		return true // 拦截
	}

	return false // 用户ID不在黑名单中，不拦截
}
