const CACHE_NAME = 'jarvis-v2';
const ASSETS = [
  '/',
  '/index.html',
  '/manifest.json',
  '/icons/',
  '/images/',
  '/actions/',
  '/js/app.js' // 假設你的主程式邏輯在這裡
];

// 安裝並快取資源
self.addEventListener('install', (e) => {
  e.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(ASSETS))
  );
});

// 攔截請求，優先從快取讀取（即使斷網也能開網頁）
self.addEventListener('fetch', (e) => {
  e.respondWith(
    caches.match(e.request).then((response) => {
      return response || fetch(e.request);
    })
  );
});

// 監聽來自頁面的訊息：處理離線儲存
self.addEventListener('message', (event) => {
  if (event.data.type === 'OFFLINE_MSG') {
    console.log('[SW] 收到離線訊息，準備進入佇列:', event.data.payload);
    // 這裡可以使用 Background Sync API 或單純通知頁面在連線後補發
  }
});

self.addEventListener('push', (event) => {
  let message = event.data ? event.data.text() : 'PCAI 系統通知';

  const options = {
    body: message,
    icon: '/icons/icon-192.png',
    badge: '/icons/icon-192.png',
    vibrate: [100, 50, 100],
    data: { url: 'https://jarvis.justdrink.com.tw/' }
  };

  event.waitUntil(
    self.registration.showNotification('J.I.I. 系統回報', options)
  );
});

// 點擊通知開啟網頁
self.addEventListener('notificationclick', (event) => {
  event.notification.close();
  event.waitUntil(
    clients.openWindow(event.notification.data.url)
  );
});
