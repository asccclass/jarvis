package main

import (
	"fmt"
	"net/http"
	"os"

	SherryServer "github.com/asccclass/sherryserver"
	"github.com/joho/godotenv"
)

// GlobalName 儲存 AI 顯示名稱，於啟動時由 envfile 的 SystemName 初始化
// 回覆 WebSocket 訊息時，display_name 欄位將填入此值
var GlobalName string

func main() {
	if err := godotenv.Load("envfile"); err != nil {
		fmt.Println(err.Error())
		return
	}

	// 初始化 AI 顯示名稱（display_name 回覆時使用）
	GlobalName = os.Getenv("SystemName")
	if GlobalName == "" {
		GlobalName = os.Getenv("WEBSOCKET_USER_ID")
	}
	if GlobalName == "" {
		GlobalName = "Jarvis"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	documentRoot := os.Getenv("DocumentRoot")
	if documentRoot == "" {
		documentRoot = "www/html"
	}
	templateRoot := os.Getenv("TemplateRoot")
	if templateRoot == "" {
		templateRoot = "www/template"
	}

	server, err := SherryServer.NewServer(":"+port, documentRoot, templateRoot)
	if err != nil {
		panic(err)
	}
	router := NewRouter(server, documentRoot)
	if router == nil {
		fmt.Println("router return nil")
		return
	}

	// 初始化 SQLite 資料庫
	dbPath := os.Getenv("DBPath")
	db, err := InitDB(dbPath)
	if err != nil {
		fmt.Printf("⚠️ 無法初始化離線資料庫: %v\n", err)
	}

	hub := newHub(db)
	go hub.run()
	router.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		webhookHandler(hub, w, r)
	})
	router.HandleFunc("/subscribe", subscribe)
	server.Server.Handler = router // server.CheckCROS(router)  // 需要自行implement, overwrite 預設的
	server.Start()
}
