package main

import(
   "fmt"
   "encoding/json"
	"sync"
	"github.com/gorilla/websocket"
)

// BroadcastMessage 定義廣播的 JSON 格式
type BroadcastMessage struct {
   Channel string `json:"channel"`
   Content string `json:"message"`
   User    string `json:"user_id,omitempty"`
   ImageData string `json:"data,omitempty"` // 存放 Base64 圖片數據
   ReplyTo string `json:"reply_to"`
   Type    string `json:"type"` // "command" 或 "response"
}

// Hub 負責管理所有連線與廣播
type Hub struct {
   clients    map[*websocket.Conn]bool
   broadcast  chan BroadcastMessage
   register   chan *websocket.Conn
   unregister chan *websocket.Conn
   mu         sync.Mutex
}

func newHub() *Hub {
   return &Hub{
	clients:    make(map[*websocket.Conn]bool),
	broadcast:  make(chan BroadcastMessage),
	register:   make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
   }
}

func (h *Hub) run() {
   for {
      select {
      case client := <-h.register:
	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()
	fmt.Println("📱 新裝置已加入控制中樞")

      case client := <-h.unregister:
	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		client.Close()
	}
	h.mu.Unlock()
	fmt.Println("🔌 裝置已中斷連線")

      case message := <-h.broadcast:
	// 將結構體轉為 JSON
	msgBytes, _ := json.Marshal(message)
	h.mu.Lock()
	for client := range h.clients {
		err := client.WriteMessage(websocket.TextMessage, msgBytes)
		if err != nil {
			fmt.Printf("❌ 廣播失敗: %v", err)
			client.Close()
			delete(h.clients, client)
		}
	}
         h.mu.Unlock()
      }
   }
}
