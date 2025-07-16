package middleware

import (
    "net/http"
    "strings"
    "pos-api-jwt-fixed/internal/token"
)

func JWTMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            http.Error(w, "未提供有效的授權標頭", http.StatusUnauthorized)
            return
        }

        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
        _, err := token.ValidateJWT(tokenStr)
        if err != nil {
            http.Error(w, "驗證失敗: "+err.Error(), http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}
