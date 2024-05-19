package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
)

type WebSocketServerClient struct {
	SelfID string
	Conn   *websocket.Conn
}

// 维护所有活跃连接的切片
var clients = []*WebSocketServerClient{}
var lock sync.Mutex
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有跨域请求
	},
}

// 全局变量和互斥锁
var (
	selfIDs   []string
	selfIDsMu sync.Mutex
)

// AddSelfID 添加一个 ID 到全局切片中
func AddSelfID(id string) {
	selfIDsMu.Lock()
	defer selfIDsMu.Unlock()
	selfIDs = append(selfIDs, id)
}

// GetSelfIDs 返回当前保存的所有 ID
func GetSelfIDs() []string {
	selfIDsMu.Lock()
	defer selfIDsMu.Unlock()
	// 返回切片的副本以防止外部修改
	copiedIDs := make([]string, len(selfIDs))
	copy(copiedIDs, selfIDs)
	return copiedIDs
}

// IsSelfIDExists 检查一个 ID 是否存在于全局切片中
func IsSelfIDExists(id string) bool {
	selfIDsMu.Lock()
	defer selfIDsMu.Unlock()
	for _, sid := range selfIDs {
		if sid == id {
			return true
		}
	}
	return false
}

// 用于处理WebSocket连接
func WsHandler(w http.ResponseWriter, r *http.Request, config *config.Config) {
	// 从请求头或URL查询参数中提取token
	tokenFromHeader := r.Header.Get("Authorization")
	selfID := r.Header.Get("X-Self-ID")
	fmtf.Printf("接入机器人X-Self-ID[%v]", selfID)
	// 加入到数组里
	AddSelfID(selfID)
	var token string
	if strings.HasPrefix(tokenFromHeader, "Token ") {
		token = strings.TrimPrefix(tokenFromHeader, "Token ")
	} else if strings.HasPrefix(tokenFromHeader, "Bearer ") {
		token = strings.TrimPrefix(tokenFromHeader, "Bearer ")
	} else {
		token = tokenFromHeader
	}
	if token == "" {
		token = r.URL.Query().Get("access_token")
	}

	// 验证token
	validToken := config.Settings.WSServerToken
	if validToken != "" && (token == "" || token != validToken) {
		if token == "" {
			log.Printf("Connection failed due to missing token. Headers: %v", r.Header)
			http.Error(w, "Missing token", http.StatusUnauthorized)
		} else {
			log.Printf("Connection failed due to incorrect token. Headers: %v, Provided token: %s", r.Header, token)
			http.Error(w, "Incorrect token", http.StatusForbidden)
		}
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %+v", err)
		return
	}
	defer conn.Close()

	lock.Lock()
	clients = append(clients, &WebSocketServerClient{
		SelfID: selfID,
		Conn:   conn,
	})
	lock.Unlock()

	clientIP := r.RemoteAddr
	log.Printf("WebSocket client connected. IP: %s", clientIP)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		if messageType == websocket.TextMessage {
			processWSMessage(p)
		}
	}
}

// 处理收到的信息
func processWSMessage(msg []byte) {
	var genericMap map[string]interface{}
	if err := json.Unmarshal(msg, &genericMap); err != nil {
		log.Printf("Error unmarshalling message to map: %v, Original message: %s\n", err, string(msg))
		return
	}

	// Assuming there's a way to distinguish notice messages, for example, checking if notice_type exists
	if noticeType, ok := genericMap["notice_type"].(string); ok && noticeType != "" {
		var noticeEvent structs.NoticeEvent
		if err := json.Unmarshal(msg, &noticeEvent); err != nil {
			log.Printf("Error unmarshalling notice event: %v\n", err)
			return
		}
		fmt.Printf("Processed a notice event of type '%s' from group %d.\n", noticeEvent.NoticeType, noticeEvent.GroupID)
		//进入处理流程

	} else if postType, ok := genericMap["post_type"].(string); ok {
		switch postType {
		case "message":
			var messageEvent structs.OnebotGroupMessage
			if err := json.Unmarshal(msg, &messageEvent); err != nil {
				log.Printf("Error unmarshalling message event: %v\n", err)
				return
			}
			fmt.Printf("Processed a message event from group %d.\n", messageEvent.GroupID)
			//进入处理流程

			// 将消息事件序列化为JSON
			data, err := json.Marshal(messageEvent)
			if err != nil {
				log.Printf("Error marshalling message event: %v\n", err)
				return
			}

			port := config.GetPort()
			// 构造请求URL
			var url string
			if config.GetLotus() == "" {
				url = "http://127.0.0.1:" + fmt.Sprint(port) + "/gensokyo"
			} else {
				url = config.GetLotus() + "/gensokyo"
			}

			// 创建POST请求
			resp, err := http.Post(url, "application/json", bytes.NewReader(data))
			if err != nil {
				log.Printf("Failed to send POST request: %v\n", err)
				return
			}
			defer resp.Body.Close()

			// 读取响应
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read response body: %v\n", err)
				return
			}

			log.Printf("Received response: %s\n", responseBody)

		case "meta_event":
			var metaEvent structs.MetaEvent
			if err := json.Unmarshal(msg, &metaEvent); err != nil {
				log.Printf("Error unmarshalling meta event: %v\n", err)
				return
			}
			fmt.Printf("Processed a meta event, heartbeat interval: %d.\n", metaEvent.Interval)
			//进入 处理流程

		}
	} else {
		log.Printf("Unknown message type or missing post type:[%v]\n", string(msg))
	}
}

// 发信息给client
func SendMessageBySelfID(selfID string, message map[string]interface{}) error {
	lock.Lock()
	defer lock.Unlock()

	for _, client := range clients {
		if client.SelfID == selfID {
			msgBytes, err := json.Marshal(message)
			if err != nil {
				return fmt.Errorf("error marshalling message: %v", err)
			}
			return client.Conn.WriteMessage(websocket.TextMessage, msgBytes)
		}
	}

	return fmt.Errorf("no connection found for selfID: %s", selfID)
}

func (client *WebSocketServerClient) Close() error {
	return client.Conn.Close()
}

func CloseAllConnections() {
	lock.Lock()
	defer lock.Unlock()

	for _, client := range clients {
		err := client.Close()
		if err != nil {
			log.Printf("failed to close connection for selfID %s: %v", client.SelfID, err)
		}
	}
	clients = nil // 清空切片，避免悬挂引用
}
