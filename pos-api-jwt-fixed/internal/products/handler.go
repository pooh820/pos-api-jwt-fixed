// === internal/products/handler.go ===
package products

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    "pos-api-jwt-fixed/internal/db"
    "pos-api-jwt-fixed/internal/woocommerce"
)

type Product struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Price string `json:"price"`
    Stock int    `json:"stock"`
}

// 取得所有商品
func GetProductsHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.DB.Query("SELECT id, name, price, stock FROM products")
    if err != nil {
        http.Error(w, "查詢失敗", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var products []Product
    for rows.Next() {
        var p Product
        rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
        products = append(products, p)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(products)
}

// 取得單一商品
func GetProductByIDHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "ID 格式錯誤", http.StatusBadRequest)
        return
    }

    var p Product
    err = db.DB.QueryRow("SELECT id, name, price, stock FROM products WHERE id = ?", id).Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
    if err == sql.ErrNoRows {
        http.Error(w, "找不到商品", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "查詢失敗", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(p)
}

// 建立商品
func CreateProductHandler(w http.ResponseWriter, r *http.Request) {
    var p Product
    if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
        http.Error(w, "格式錯誤", http.StatusBadRequest)
        return
    }

    result, err := db.DB.Exec("INSERT INTO products (name, price, stock) VALUES (?, ?, ?)", p.Name, p.Price, p.Stock)
    if err != nil {
        http.Error(w, "新增失敗", http.StatusInternalServerError)
        return
    }

    id, _ := result.LastInsertId()
    p.ID = int(id)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(p)
}

// 修改商品
func UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "ID 格式錯誤", http.StatusBadRequest)
        return
    }

    var p Product
    if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
        http.Error(w, "格式錯誤", http.StatusBadRequest)
        return
    }

    _, err = db.DB.Exec("UPDATE products SET name = ?, price = ?, stock = ? WHERE id = ?", p.Name, p.Price, p.Stock, id)
    if err != nil {
        http.Error(w, "更新失敗", http.StatusInternalServerError)
        return
    }

    p.ID = id
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(p)
}

// 刪除商品
func DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "ID 格式錯誤", http.StatusBadRequest)
        return
    }

    _, err = db.DB.Exec("DELETE FROM products WHERE id = ?", id)
    if err != nil {
        http.Error(w, "刪除失敗", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// WooCommerce 商品同步 Handler
func SyncProductsHandler(w http.ResponseWriter, r *http.Request) {
    wooProducts, err := woocommerce.GetProductsFromWooCommerce()
    if err != nil {
        http.Error(w, "同步 WooCommerce 商品失敗: "+err.Error(), http.StatusInternalServerError)
        return
    }

    var products []Product
    for _, wp := range wooProducts {
        id, ok := convertID(wp.ID)
        if !ok {
            fmt.Printf("⚠️ WooCommerce 商品 ID 格式錯誤，跳過：%v\n", wp.ID)
            continue
        }

        product := Product{
            ID:    id,
            Name:  wp.Name,
            Price: wp.RegularPrice,
            Stock: wp.StockQuantity,
        }

        _, dbErr := db.DB.Exec("INSERT INTO products (id, name, price, stock) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE name=?, price=?, stock=?", product.ID, product.Name, product.Price, product.Stock, product.Name, product.Price, product.Stock)
        if dbErr != nil {
            fmt.Printf("⚠️ 資料庫寫入失敗：%v\n", dbErr)
            continue
        }

        products = append(products, product)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(products)
}

// WooCommerce ID 轉換
func convertID(wooID int) (int, bool) {
    if wooID <= 0 {
        return 0, false
    }
    return wooID, true
}

