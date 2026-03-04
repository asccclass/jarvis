package main

import (
   "os"
   "fmt"
   "net/http"
   "github.com/joho/godotenv"
   "github.com/asccclass/sherryserver"
)

func main() {
   if err := godotenv.Load("envfile"); err != nil {
      fmt.Println(err.Error())
      return
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

   server, err := SherryServer.NewServer(":" + port, documentRoot, templateRoot)
   if err != nil {
      panic(err)
   }
   router := NewRouter(server, documentRoot)
   if router == nil {
      fmt.Println("router return nil")
      return
   }
   hub := newHub()
   go hub.run()
   router.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
      webhookHandler(hub, w, r)
   })
   router.HandleFunc("/subscribe", subscribe)
   server.Server.Handler = router  // server.CheckCROS(router)  // 需要自行implement, overwrite 預設的
   server.Start()
}
