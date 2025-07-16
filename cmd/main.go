package main

import (
    "fmt"
    "log"
    "net/http"

    "pos-api-jwt-fixed/internal/db"
    "pos-api-jwt-fixed/router"
    "pos-api-jwt-fixed/internal/config"
    "pos-api-jwt-fixed/internal/woocommerce"
)

func main() {
    // 讀取設定檔
    err := config.LoadConfig("internal/config/config.yaml")
    if err != nil {
        log.Fatalf("讀取設定檔失敗: %v", err)
    }

    fmt.Println("伺服器環境:", config.Cfg.Server.Env)
    fmt.Println("資料庫連線位址:", config.Cfg.Database.Host)

    // 初始化資料庫
    db.InitDB()

    if err := woocommerce.TestConnection(); err != nil {
    log.Printf("⚠️ WooCommerce API 測試失敗（不影響啟動）: %v", err)
    }

    // 初始化路由
    r := router.NewRouter()

    log.Println("✅ 伺服器啟動：http://localhost:8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatalf("❌ 伺服器啟動失敗：%v", err)
    }
}

