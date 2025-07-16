package orders

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    "time"
    "fmt"

    "github.com/gorilla/mux"
    "pos-api-jwt-fixed/internal/db"
)

type Order struct {
    ID            int         `json:"id"`
    CustomerID    int         `json:"customer_id"`
    TotalAmount   float64     `json:"total_amount"`
    PaymentMethod string      `json:"payment_method"`
    CreatedAt     time.Time   `json:"created_at"`
    Items         []OrderItem `json:"items"`
}

type OrderItem struct {
    ID        int     `json:"id"`
    ProductID int     `json:"product_id"`
    Quantity  int     `json:"quantity"`
    UnitPrice float64 `json:"unit_price"`
}

// 建立訂單
func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
    var order Order
    if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
        http.Error(w, "格式錯誤", http.StatusBadRequest)
        return
    }

    tx, err := db.DB.Begin()
    if err != nil {
        http.Error(w, "資料庫交易啟動失敗", http.StatusInternalServerError)
        return
    }

    // 計算總金額
    var totalAmount float64 = 0
    for i, item := range order.Items {
        var unitPrice float64
        err := tx.QueryRow("SELECT price FROM products WHERE id = ?", item.ProductID).Scan(&unitPrice)
        if err != nil {
            tx.Rollback()
            http.Error(w, fmt.Sprintf("查詢商品 %d 價格失敗", item.ProductID), http.StatusInternalServerError)
            return
        }
        order.Items[i].UnitPrice = unitPrice
        totalAmount += unitPrice * float64(item.Quantity)
    }

    // 寫入 orders
    result, err := tx.Exec("INSERT INTO orders (customer_id, total_amount, payment_method) VALUES (?, ?, ?)",
        order.CustomerID, totalAmount, "cash") // 此處先固定為 cash
    if err != nil {
        tx.Rollback()
        http.Error(w, "訂單建立失敗", http.StatusInternalServerError)
        return
    }

    orderID, _ := result.LastInsertId()
    order.ID = int(orderID)

    // 寫入 order_items
    for _, item := range order.Items {
        _, err := tx.Exec("INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES (?, ?, ?, ?)",
            orderID, item.ProductID, item.Quantity, item.UnitPrice)
        if err != nil {
            tx.Rollback()
            http.Error(w, "訂單明細建立失敗", http.StatusInternalServerError)
            return
        }
    }

    if err := tx.Commit(); err != nil {
        http.Error(w, "訂單儲存失敗", http.StatusInternalServerError)
        return
    }

    order.TotalAmount = totalAmount
    order.PaymentMethod = "cash"

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}

// 查詢訂單列表
func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.DB.Query("SELECT id, customer_id, total_amount, payment_method, created_at FROM orders")
    if err != nil {
        http.Error(w, "查詢訂單失敗", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var orders []Order
    for rows.Next() {
        var o Order
        if err := rows.Scan(&o.ID, &o.CustomerID, &o.TotalAmount, &o.PaymentMethod, &o.CreatedAt); err == nil {
            orders = append(orders, o)
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(orders)
}

// 查詢單一訂單
func GetOrderByIDHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "ID 格式錯誤", http.StatusBadRequest)
        return
    }

    var o Order
    err = db.DB.QueryRow("SELECT id, customer_id, total_amount, payment_method, created_at FROM orders WHERE id = ?", id).
        Scan(&o.ID, &o.CustomerID, &o.TotalAmount, &o.PaymentMethod, &o.CreatedAt)
    if err == sql.ErrNoRows {
        http.Error(w, "找不到訂單", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "查詢訂單失敗", http.StatusInternalServerError)
        return
    }

    rows, err := db.DB.Query("SELECT id, product_id, quantity, unit_price FROM order_items WHERE order_id = ?", id)
    if err != nil {
        http.Error(w, "查詢訂單明細失敗", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    for rows.Next() {
        var item OrderItem
        if err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.UnitPrice); err == nil {
            o.Items = append(o.Items, item)
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(o)
}

// 刪除訂單
func DeleteOrderHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "ID 格式錯誤", http.StatusBadRequest)
        return
    }

    tx, err := db.DB.Begin()
    if err != nil {
        http.Error(w, "資料庫交易啟動失敗", http.StatusInternalServerError)
        return
    }

    _, err = tx.Exec("DELETE FROM order_items WHERE order_id = ?", id)
    if err != nil {
        tx.Rollback()
        http.Error(w, "刪除訂單明細失敗", http.StatusInternalServerError)
        return
    }

    _, err = tx.Exec("DELETE FROM orders WHERE id = ?", id)
    if err != nil {
        tx.Rollback()
        http.Error(w, "刪除訂單失敗", http.StatusInternalServerError)
        return
    }

    if err := tx.Commit(); err != nil {
        http.Error(w, "刪除訂單失敗", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// 修改訂單
func UpdateOrderHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "ID 格式錯誤", http.StatusBadRequest)
        return
    }

    var order Order
    if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
        http.Error(w, "格式錯誤", http.StatusBadRequest)
        return
    }

    tx, err := db.DB.Begin()
    if err != nil {
        http.Error(w, "資料庫交易啟動失敗", http.StatusInternalServerError)
        return
    }

    // 更新主訂單表（目前只有 customer 欄位）
    _, err = tx.Exec("UPDATE orders SET customer_id = ? WHERE id = ?", order.CustomerID, id)
    if err != nil {
        tx.Rollback()
        http.Error(w, "訂單主資料更新失敗", http.StatusInternalServerError)
        return
    }

    // 刪除原有的 order_items
    _, err = tx.Exec("DELETE FROM order_items WHERE order_id = ?", id)
    if err != nil {
        tx.Rollback()
        http.Error(w, "訂單明細刪除失敗", http.StatusInternalServerError)
        return
    }

    // 新增新的 order_items
    for _, item := range order.Items {
        _, err := tx.Exec("INSERT INTO order_items (order_id, product_id, quantity) VALUES (?, ?, ?)", id, item.ProductID, item.Quantity)
        if err != nil {
            tx.Rollback()
            http.Error(w, "訂單明細新增失敗", http.StatusInternalServerError)
            return
        }
    }

    if err := tx.Commit(); err != nil {
        http.Error(w, "訂單更新失敗", http.StatusInternalServerError)
        return
    }

    order.ID = id
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}

