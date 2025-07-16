package customers

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "strings"

    "github.com/gorilla/mux"
    "pos-api-jwt-fixed/internal/db"
    "pos-api-jwt-fixed/internal/woocommerce"
)

type Customer struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// 取得所有會員
func GetCustomersHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.DB.Query("SELECT id, name, email FROM customers")
    if err != nil {
        http.Error(w, "查詢失敗", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var customers []Customer
    for rows.Next() {
        var c Customer
        rows.Scan(&c.ID, &c.Name, &c.Email)
        customers = append(customers, c)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(customers)
}

// 取得單一會員
func GetCustomerByIDHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "ID 格式錯誤", http.StatusBadRequest)
        return
    }

    var c Customer
    err = db.DB.QueryRow("SELECT id, name, email FROM customers WHERE id = ?", id).Scan(&c.ID, &c.Name, &c.Email)
    if err == sql.ErrNoRows {
        http.Error(w, "找不到會員", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "查詢失敗", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(c)
}

// 建立會員
func CreateCustomerHandler(w http.ResponseWriter, r *http.Request) {
    var c Customer
    if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
        http.Error(w, "格式錯誤", http.StatusBadRequest)
        return
    }

    result, err := db.DB.Exec("INSERT INTO customers (name, email) VALUES (?, ?)", c.Name, c.Email)
    if err != nil {
        http.Error(w, "新增失敗", http.StatusInternalServerError)
        return
    }

    id, _ := result.LastInsertId()
    c.ID = int(id)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(c)
}

// 修改會員
func UpdateCustomerHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "ID 格式錯誤", http.StatusBadRequest)
        return
    }

    var c Customer
    if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
        http.Error(w, "格式錯誤", http.StatusBadRequest)
        return
    }

    _, err = db.DB.Exec("UPDATE customers SET name = ?, email = ? WHERE id = ?", c.Name, c.Email, id)
    if err != nil {
        http.Error(w, "更新失敗", http.StatusInternalServerError)
        return
    }

    c.ID = id
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(c)
}

// 刪除會員
func DeleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "ID 格式錯誤", http.StatusBadRequest)
        return
    }

    _, err = db.DB.Exec("DELETE FROM customers WHERE id = ?", id)
    if err != nil {
        http.Error(w, "刪除失敗", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// WooCommerce 會員同步 Handler
func SyncCustomersHandler(w http.ResponseWriter, r *http.Request) {
    wooCustomers, err := woocommerce.GetCustomersFromWooCommerce()
    if err != nil {
        http.Error(w, "同步 WooCommerce 會員失敗: "+err.Error(), http.StatusInternalServerError)
        return
    }

    var customers []Customer
    for _, wc := range wooCustomers {
           name := wc.FirstName + " " + wc.LastName
           name = strings.TrimSpace(name)
           if name == "" {
              name = wc.Username 
           }

           customer := Customer{
              ID:    wc.ID,
              Name:  name,
              Email: wc.Email,
           }

        _, dbErr := db.DB.Exec(`
            INSERT INTO customers (id, name, email)
            VALUES (?, ?, ?)
            ON DUPLICATE KEY UPDATE name=?, email=?`,
            customer.ID, customer.Name, customer.Email,
            customer.Name, customer.Email,
        )
        if dbErr != nil {
            fmt.Printf("⚠️ 資料庫寫入失敗：%v\n", dbErr)
            continue
        }

        customers = append(customers, customer)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(customers)
}


func trimSpace(s string) string {
    return string([]byte(s)) // 若您有 strings.TrimSpace 可以直接用那個
}

// ID 型別轉換函式
func convertID(rawID interface{}) (int, bool) {
    switch v := rawID.(type) {
    case float64:
        return int(v), true
    case int:
        return v, true
    case string:
        id, err := strconv.Atoi(v)
        if err == nil {
            return id, true
        }
        return 0, false
    default:
        return 0, false
    }
}

