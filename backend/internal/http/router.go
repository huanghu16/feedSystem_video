package http

import (
	"feedSystem_video/backend/internal/account"
	"feedSystem_video/backend/internal/apierror"

	"github.com/gin-gonic/gin"
)

// SetupRouter 创建并配置 Gin 路由
func SetupRouter() *gin.Engine {
	r := gin.Default() // 创建 Gin 引擎

	// 健康检查路由
	r.GET("/healthz", func(c *gin.Context) {
		apierror.OK(c, gin.H{"status": "ok"})
	})

	// --- Account 模块 ---
	// 依赖注入链：Repo → Service → Handler
	// 每一层只依赖上一层的接口，不跨层调用
	repo := account.NewRepo()              // 创建数据仓库
	service := account.NewService(repo)    // 创建业务逻辑层
	handler := account.NewHandler(service) // 创建 HTTP 处理层

	// 路由组：/account 前缀
	accountGroup := r.Group("/account")
	{
		accountGroup.POST("/register", handler.Register)
		accountGroup.POST("/login", handler.Login)

	}
	return r
}
