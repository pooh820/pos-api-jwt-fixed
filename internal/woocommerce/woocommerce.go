package woocommerce

import (
    "bytes" // ← 必須匯入
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
)

// WooCommerce API 基本參數
const (
    BaseURL        = "http://56.155.1.66/finery/wp-json/wc/v3"
    ConsumerKey    = "ck_812165fc15e5bbd1f30f96e027e94c7592a5438d"
    ConsumerSecret = "cs_0ee1b3dc51954092b9a20e19fb5c482e9159485c"
)

// WooCommerce 會員結構
type WooCustomer struct {
    ID        int    `json:"id"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Email     string `json:"email"`
    Username  string `json:"username"`
}

// WooCommerce 商品結構
type WooProduct struct {
    ID            int    `json:"id"`
    Name          string `json:"name"`
    RegularPrice  string `json:"regular_price"`
    StockQuantity int    `json:"stock_quantity"`
}

// 測試 WooCommerce API 連線
func TestConnection() error {
    url := fmt.Sprintf("%s/customers?consumer_key=%s&consumer_secret=%s", BaseURL, ConsumerKey, ConsumerSecret)

    resp, err := http.Get(url)
    if err != nil {
        return fmt.Errorf("無法連線 WooCommerce API: %v", err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("API 錯誤: %s", string(body))
    }

    fmt.Println("✅ WooCommerce 連線成功，回傳資料：")
    fmt.Println(string(body))

    return nil
}

// 取得 WooCommerce 會員資料
func GetCustomersFromWooCommerce() ([]WooCustomer, error) {
    url := fmt.Sprintf("%s/customers?consumer_key=%s&consumer_secret=%s", BaseURL, ConsumerKey, ConsumerSecret)

    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("無法連線 WooCommerce API: %v", err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API 錯誤: %s", string(body))
    }

    var customers []WooCustomer
    json.Unmarshal(body, &customers)

    return customers, nil
}

// 取得 WooCommerce 商品資料
func GetProductsFromWooCommerce() ([]WooProduct, error) {
    url := fmt.Sprintf("%s/products?consumer_key=%s&consumer_secret=%s", BaseURL, ConsumerKey, ConsumerSecret)

    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("無法連線 WooCommerce API: %v", err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API 錯誤: %s", string(body))
    }

    var products []WooProduct
    json.Unmarshal(body, &products)

    return products, nil
}

// 建立 WooCommerce 訂單
func CreateOrderInWooCommerce(orderData map[string]interface{}) (string, error) {
    url := fmt.Sprintf("%s/orders?consumer_key=%s&consumer_secret=%s", BaseURL, ConsumerKey, ConsumerSecret)

    bodyBytes, _ := json.Marshal(orderData)

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
    if err != nil {
        return "", fmt.Errorf("建立請求失敗: %v", err)
    }

    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("呼叫 WooCommerce API 失敗: %v", err)
    }
    defer resp.Body.Close()

    respBody, _ := io.ReadAll(resp.Body)

    if resp.StatusCode != http.StatusCreated {
        return "", fmt.Errorf("API 錯誤: %s", string(respBody))
    }

    var result struct {
        ID int `json:"id"`
    }
    json.Unmarshal(respBody, &result)

    if result.ID == 0 {
        return "", errors.New("WooCommerce 回傳訂單 ID 無效")
    }

    return fmt.Sprintf("%d", result.ID), nil
}

