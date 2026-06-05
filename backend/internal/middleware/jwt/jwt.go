package jwt

import (
	"feedSystem_video/internal/apierror"
	"feedSystem_video/internal/auth"
	"strings"

	"github.com/gin-gonic/gin"
)

// AccountIDKey 是 Gin Context 中存储用户 ID 的 key
const AccountIDKey = "account_id"

// JWTAuth 强制 JWT 验证中间件
// 用于需要登录的接口（如发布视频、点赞）
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			apierror.FailAuth(c, "缺少认证信息")
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(token)
		if err != nil {
			apierror.FailAuth(c, "认证已过期或无效")
			c.Abort()
			return
		}

		// 把用户 ID 和用户名存入 Context，后续 handler 可以直接用
		c.Set(AccountIDKey, claims.AccountID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// SoftJWTAuth 可选 JWT 验证中间件
// 用于 Feed 流等接口：登录用户看个性化内容，未登录用户看默认内容
func SoftJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			// 没有 token 也放行，只是不设置用户信息
			c.Next()
			return
		}

		claims, err := auth.ParseToken(token)
		if err != nil {
			// token 无效也放行（和没登录一样）
			c.Next()
			return
		}

		c.Set(AccountIDKey, claims.AccountID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// extractToken 从请求头中提取 Bearer Token
// 格式：Authorization: Bearer <token>
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}
