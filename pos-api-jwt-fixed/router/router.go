// === router/router.go ===
package router

import (
    "net/http"

    "github.com/gorilla/mux"
    "pos-api-jwt-fixed/internal/auth"
    "pos-api-jwt-fixed/internal/customers"
    "pos-api-jwt-fixed/internal/middleware"
    "pos-api-jwt-fixed/internal/orders"
    "pos-api-jwt-fixed/internal/products"
)

func NewRouter() http.Handler {
    r := mux.NewRouter()

    // 不需驗證的登入路由
    r.HandleFunc("/api/auth/login", auth.LoginHandler).Methods("POST")

    // API 群組（需 JWT 驗證）
    api := r.PathPrefix("/api").Subrouter()
    api.Use(middleware.JWTMiddleware)

    // 會員模組
    api.HandleFunc("/customers", customers.GetCustomersHandler).Methods("GET")
    api.HandleFunc("/customers/{id}", customers.GetCustomerByIDHandler).Methods("GET")
    api.HandleFunc("/customers", customers.CreateCustomerHandler).Methods("POST")
    api.HandleFunc("/customers/{id}", customers.UpdateCustomerHandler).Methods("PUT")
    api.HandleFunc("/customers/{id}", customers.DeleteCustomerHandler).Methods("DELETE")
    api.HandleFunc("/customers/sync", customers.SyncCustomersHandler).Methods("POST")

    // 商品模組
    api.HandleFunc("/products", products.GetProductsHandler).Methods("GET")
    api.HandleFunc("/products/{id}", products.GetProductByIDHandler).Methods("GET")
    api.HandleFunc("/products", products.CreateProductHandler).Methods("POST")
    api.HandleFunc("/products/{id}", products.UpdateProductHandler).Methods("PUT")
    api.HandleFunc("/products/{id}", products.DeleteProductHandler).Methods("DELETE")
    api.HandleFunc("/products/sync", products.SyncProductsHandler).Methods("POST")

    // 訂單模組
    api.HandleFunc("/orders", orders.GetOrdersHandler).Methods("GET")
    api.HandleFunc("/orders/{id}", orders.GetOrderByIDHandler).Methods("GET")
    api.HandleFunc("/orders", orders.CreateOrderHandler).Methods("POST")
    api.HandleFunc("/orders/{id}", orders.DeleteOrderHandler).Methods("DELETE")
    api.HandleFunc("/orders/{id}", orders.UpdateOrderHandler).Methods("PUT")

    return r
}

