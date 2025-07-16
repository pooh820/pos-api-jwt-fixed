package token

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("your_secret_key") // 建議改為環境變數

// 簽發 JWT Token
func GenerateJWT(username string) (string, error) {
    claims := jwt.MapClaims{}
    claims["sub"] = username
    claims["exp"] = time.Now().Add(24 * time.Hour).Unix()

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtKey)
}

// 驗證 JWT Token 並取出用戶名稱
func ValidateJWT(tokenStr string) (string, error) {
    token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
        // 驗證簽章演算法
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return jwtKey, nil
    })

    if err != nil {
        return "", err
    }

    // 驗證 claims 格式並取出 sub 欄位
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        if sub, ok := claims["sub"].(string); ok {
            return sub, nil
        }
        return "", errors.New("sub 欄位缺失或格式錯誤")
    }

    return "", errors.New("無效的 Token")
}

// 從 Header 取出純 token
func ExtractTokenFromHeader(header string) string {
    const prefix = "Bearer "
    if len(header) > len(prefix) && header[:len(prefix)] == prefix {
        return header[len(prefix):]
    }
    return header
}

