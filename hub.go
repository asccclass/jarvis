package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

// BroadcastMessage 定義廣播的 JSON 格式
type BroadcastMessage struct {
	Channel   string `json:"channel"`
	Content   string `json:"message"`
	User      string `json:"user_id,omitempty"`
	ImageData string `json:"data,omitempty"` // 存放 Base64 圖片數據
	ReplyTo   string `json:"reply_to"`
	Type      string `json:"type"` // "command" 或 "response"
}

// Client 包含 WebSocket 連線與使用者識別碼
type Client struct {
	Conn   *websocket.Conn
	UserID string
}

// Hub 負責管理所有連線與廣播
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan BroadcastMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
	db         *sql.DB
}

func newHub(db *sql.DB) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan BroadcastMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		db:         db,
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			fmt.Printf("📱 新裝置已加入控制中樞: [%s]\n", client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Conn.Close()
			}
			h.mu.Unlock()
			fmt.Printf("🔌 裝置已中斷連線: [%s]\n", client.UserID)

		case message := <-h.broadcast:
			// 將結構體轉為 JSON
			msgBytes, _ := json.Marshal(message)

			// 如果有指定 reply_to，檢查目標是否在線
			targetOnline := false

			h.mu.Lock()
			for client := range h.clients {
				// 如果沒有指定 reply_to，或者是指定的對象，就發送
				if message.ReplyTo == "" || message.ReplyTo == client.UserID {
					err := client.Conn.WriteMessage(websocket.TextMessage, msgBytes)
					if err != nil {
						fmt.Printf("❌ 廣播失敗或斷線: %v", err)
						client.Conn.Close()
						delete(h.clients, client)
					} else {
						if message.ReplyTo == client.UserID {
							targetOnline = true
						}
					}
				}
			}
			h.mu.Unlock()

			// 如果有指定接收者，但該接收者不在線，存入離線資料庫
			if message.ReplyTo != "" && !targetOnline && h.db != nil {
				err := SaveOfflineMessage(h.db, message.ReplyTo, string(msgBytes))
				if err != nil {
					fmt.Printf("⚠️ 離線訊息儲存失敗: %v\n", err)
				}
			}
		}
	}
}
