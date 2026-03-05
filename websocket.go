package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// WebhookRequest 定義前端傳來的資料結構
type WebhookRequest struct {
	Query     string    `json:"query"`
	Timestamp time.Time `json:"timestamp"`
	User      string    `json:"user"`
}

// WebhookResponse 定義回傳給前端的資料結構
type WebhookResponse struct {
	Reply  string `json:"reply"`
	Status string `json:"status"`
}

const (
	// 允許寫入訊息至 Client 的時間
	writeWait = 10 * time.Second
	// 允許從 Client 接收下一個 Pong 訊息的時間
	pongWait = 60 * time.Second
	// 發送 Ping 的頻率（必須小於 pongWait）
	pingPeriod = (pongWait * 9) / 10
	// 最大訊息大小
	maxMessageSize = 512 * 1024
)

// 設定 Upgrader，允許跨域連線 (CORS)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 開發環境允許所有來源，部署時應更嚴格
	},
}

func webhookHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("⚠️ 升級失敗:", err)
		return
	}

	// 取得本機固定身分 userID（由環境變數 WEBSOCKET_USER_ID 賦予，不可變）
	userID := os.Getenv("WEBSOCKET_USER_ID")
	if userID == "" {
		// fallback：從 URL query 取得，僅用於開發除錯
		userID = r.URL.Query().Get("user")
	}
	if userID == "" {
		userID = "unknown"
	}

	client := &Client{Conn: conn, UserID: userID}
	hub.register <- client

	// --- 發送此使用者的所有離線訊息 ---
	if hub.db != nil {
		offlineMsgs, err := GetOfflineMessages(hub.db, userID)
		if err == nil && len(offlineMsgs) > 0 {
			fmt.Printf("📦 找到 %d 筆 [%s] 的離線訊息，準備發送\n", len(offlineMsgs), userID)
			for _, msg := range offlineMsgs {
				conn.WriteMessage(websocket.TextMessage, []byte(msg))
			}
			DeleteOfflineMessages(hub.db, userID)
		}
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("🚨 PCAI 邏輯保護: %v", r)
			}
			hub.unregister <- client
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			var incoming BroadcastMessage
			json.Unmarshal(message, &incoming)

			// 取得發送者的顯示名稱（接收時語意：填入發送者的 display_name 或 user_id）
			senderDisplay := incoming.DisplayName
			if senderDisplay == "" {
				senderDisplay = incoming.UserID
			}

			if incoming.Type == "image_analysis" {
				fmt.Printf("📸 收到來自 %s 的圖片分析請求", senderDisplay)
				// 此處串接您的 Ollama 或其他多模態模型 (如 LLaVA)
				// go analyzeImage(hub, incoming)
			} else {
				fmt.Println("收到:", string(message))

				hub.broadcast <- BroadcastMessage{
					Channel:     "pcai",
					Content:     string(message),
					UserID:      userID,        // 固定本機身分（env WEBSOCKET_USER_ID）
					DisplayName: senderDisplay, // 接收時填入發送者的顯示名稱
					Type:        incoming.Type,
				}
			}
		}
	}()
}

// Reply 封裝 AI 回覆訊息並廣播
// - user_id：固定讀取 WEBSOCKET_USER_ID，代表本機（AI）的唯一身分
// - display_name：填入 GlobalName（AI 的顯示名稱）
func Reply(hub *Hub, replyTo string, content string, msgType string) {
	userID := os.Getenv("WEBSOCKET_USER_ID")
	if userID == "" {
		userID = "jarvis"
	}

	hub.broadcast <- BroadcastMessage{
		Channel:     "pcai",
		Content:     content,
		UserID:      userID,     // 固定本機身分
		DisplayName: GlobalName, // 回覆時填 AI 名稱
		ReplyTo:     replyTo,
		Type:        msgType,
	}
}
