package auth

import (
    "encoding/json"
    "net/http"
    "pos-api-jwt-fixed/internal/token"
)

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Token string `json:"token"`
}

type StaffProfile struct {
    Username string `json:"username"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil || req.Username != "admin" || req.Password != "123456" {
        http.Error(w, "帳號或密碼錯誤", http.StatusUnauthorized)
        return
    }

    tokenStr, err := token.GenerateJWT(req.Username)
    if err != nil {
        http.Error(w, "JWT 簽發失敗", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(LoginResponse{Token: tokenStr})
}

// 員工基本資料查詢 API
func GetStaffProfileHandler(w http.ResponseWriter, r *http.Request) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        http.Error(w, "未提供 Token", http.StatusUnauthorized)
        return
    }

    tokenStr := token.ExtractTokenFromHeader(authHeader)
    username, err := token.ValidateJWT(tokenStr)
    if err != nil {
        http.Error(w, "Token 驗證失敗", http.StatusUnauthorized)
        return
    }

    profile := StaffProfile{Username: username}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(profile)
}
