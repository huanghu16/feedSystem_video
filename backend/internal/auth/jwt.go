package auth

import (
	"crypto/rand"
	"encoding/hex"
	"feedSystem_video/internal/config"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 是 JWT 载荷（Payload），包含用户信息
// 这些字段会编码到 token 字符串中
type Claims struct {
	AccountID uint   `json:"account_id"`
	Username  string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateTokenPair 生成 access_token + refresh_token
// access_token：15分钟有效，放在请求头里
// refresh_token：7天有效，只用来换新的 access_token
func GenerateTokenPair(accountID uint, username string) (string, string, error) {
	// 生成 access_token
	accessClaims := Claims{
		AccountID: accountID,
		Username:  username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.C.JWT.AccessTTL) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
		SignedString([]byte(config.C.JWT.Secret))
	if err != nil {
		return "", "", fmt.Errorf("生成 access token 失败: %w", err)
	}
	// 生成 refresh_token
	// Refresh Token 是随机字符串，不存用户信息，只用来换 access token
	refreshToken := generateRandomToken(32)

	return accessToken, refreshToken, nil
}

// ParseToken 解析并验证 access token
// 返回 Claims（包含 account_id 和 username）
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法，防止算法替换攻击
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名算法: %v", token.Header["alg"])
		}
		return []byte(config.C.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("无效的 token")
}

// generateRandomToken 生成随机字符串（用于 refresh token）
func generateRandomToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}
