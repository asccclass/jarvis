package main

import (
   "fmt"
   "time"
   "net/http"
   "encoding/json"
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
   Reply   string `json:"reply"`
   Status  string `json:"status"`
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
   hub.register <- conn
   go func() {
      defer func() {
	if r := recover(); r != nil {
		fmt.Printf("🚨 PCAI 邏輯保護: %v", r)
	}
	hub.unregister <- conn
      }()

      for {
         _, message, err := conn.ReadMessage()
	if err != nil {
	   break
	}
        var incoming BroadcastMessage
        json.Unmarshal(message, &incoming)
        if incoming.Type == "image_analysis" {
          fmt.Printf("📸 收到來自 %s 的圖片分析請求", incoming.User)
          // 此處串接您的 Ollama 或其他多模態模型 (如 LLaVA)
          // go analyzeImage(hub, incoming)
        } else {
           fmt.Println("收到:", string(message))
	   // 收到訊息後，封裝成 pcai 頻道並廣播
	   hub.broadcast <- BroadcastMessage{
	      Channel: "pcai",
	      Content: string(message),
	      User:    "Jii 哥", // 帶上你的使用者識別
	      Type:    incoming.Type,
	   }
        }
      }
   }()
}
