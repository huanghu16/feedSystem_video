package middleware

import (
	"feedSystem_video/internal/apierror"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
// 前后端分离架构必需，允许前端（如 localhost:5173）访问后端 API
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// 预检请求直接返回 200
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}

// Recovery panic 恢复中间件（补充 gin.Default 的 Recovery，增加 JSON 响应）
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				apierror.FailServer(c, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}
