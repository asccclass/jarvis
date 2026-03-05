package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/SherClockHolmes/webpush-go"
)

// 這些金鑰請務必安全存放
var (
	vapidPublicKey  = os.Getenv("VAPIDKEY_PUBLIC")
	vapidPrivateKey = os.Getenv("VAPIDKEY_PRIVATE")
)

// 儲存訂閱資訊 (實際專案建議存入資料庫)
var subscriptions = make(map[string]webpush.Subscription)

func subscribe(w http.ResponseWriter, r *http.Request) {
	var sub webpush.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// 以 Jii 哥 作為 ID 儲存
	subscriptions["Jii 哥"] = sub
	w.WriteHeader(http.StatusCreated)
	fmt.Println("✅ 已收到 Jii 哥 的推送訂閱")
}

func notifyHandler(w http.ResponseWriter, r *http.Request) {
	sub, ok := subscriptions["Jii 哥"]
	if !ok {
		http.Error(w, "找不到訂閱資訊", http.StatusNotFound)
		return
	}

	// 模擬 PCAI 任務完成後的通知
	message := "報告 Jii 哥，PCAI 已完成您的行事曆分析。"

	resp, err := webpush.SendNotification([]byte(message), &sub, &webpush.Options{
		Subscriber:      "mailto:justgps@gmail.com",
		VAPIDPublicKey:  vapidPublicKey,
		VAPIDPrivateKey: vapidPrivateKey,
		TTL:             30,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	w.Write([]byte("通知已送出"))
}
