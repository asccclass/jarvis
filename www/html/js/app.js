const VAPID_PUBLIC_KEY = "BJIylsrpQo4-n-tUJY6dafgTplaJncAC2eZuvZ-JACVd4CnVetD39KABd8fBWOmHcf3moynlvGoMWeGnMmi_-XY";

// --- 離線佇列管理器 ---
const OfflineManager = {
    dbName: 'PCAI_OfflineDB',
    storeName: 'queue',

    // 初始化資料庫
    async init() {
        return new Promise((resolve) => {
            const request = indexedDB.open(this.dbName, 1);
            request.onupgradeneeded = (e) => {
                const db = e.target.result;
                db.createObjectStore(this.storeName, { autoIncrement: true });
            };
            request.onsuccess = (e) => resolve(e.target.result);
        });
    },

    // 儲存訊息到佇列
    async save(payload) {
        const db = await this.init();
        const tx = db.transaction(this.storeName, 'readwrite');
        tx.objectStore(this.storeName).add(payload);
        console.log('📦 訊息已存入離線佇列');
    },

    // 補發所有暫存訊息
    async processQueue(socket) {
        const db = await this.init();
        const tx = db.transaction(this.storeName, 'readwrite');
        const store = tx.objectStore(this.storeName);
        const request = store.openCursor();

        request.onsuccess = (e) => {
            const cursor = e.target.result;
            if (cursor) {
                if (socket.readyState === WebSocket.OPEN) {
                    console.log('🚀 補發離線訊息:', cursor.value);
                    socket.send(JSON.stringify(cursor.value));
                    store.delete(cursor.key); // 發送成功後刪除
                    cursor.continue();
                }
            }
        };
    }
};


async function enableNotifications() {
    // 1. 請求權限
    const permission = await Notification.requestPermission();
    if (permission !== 'granted') {
        console.error('❌ 推送權限被拒絕');
        return;
    }

    // 2. 取得 Service Worker 註冊資訊
    const registration = await navigator.serviceWorker.ready;

    // 3. 訂閱 Push 伺服器
    const subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(VAPID_PUBLIC_KEY)
    });

    // 4. 將訂閱資訊傳送至 Go 後端
    await fetch('https://jarvis.justdrink.com.tw/subscribe', {
        method: 'POST',
        body: JSON.stringify(subscription),
        headers: { 'Content-Type': 'application/json' }
    });

    alert('✅ PCAI 推送通知已設定完成');
}

// 輔助函式：轉換 Base64 Key
function urlBase64ToUint8Array(base64String) {
    const padding = '='.repeat((4 - base64String.length % 4) % 4);
    const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
    const rawData = window.atob(base64);
    return Uint8Array.from([...rawData].map((char) => char.charCodeAt(0)));
}
