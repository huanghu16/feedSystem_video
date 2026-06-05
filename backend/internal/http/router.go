package http

import (
	"feedSystem_video/backend/internal/account"
	"feedSystem_video/backend/internal/apierror"
	"feedSystem_video/backend/internal/middleware/jwt"
	"feedSystem_video/backend/internal/video"

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

	// --- Video 模块 ---
	videoRepo := video.NewRepo()
	videoService := video.NewService(videoRepo, repo) // repo 是 account 的 repo
	videoHandler := video.NewHandler(videoService)

	videoGroup := r.Group("/video")
	{
		// 需要 JWT 的接口
		videoGroup.Use(jwt.JWTAuth())
		videoGroup.POST("/publish", videoHandler.Publish)
		videoGroup.POST("/uploadVideo", videoHandler.UploadVideo)
	}

	// 不需要 JWT 的接口（放在外面）
	r.POST("/video/listByAuthorID", videoHandler.ListByAuthor)

	// 静态文件服务（让上传的视频可以通过 URL 访问）
	r.Static("/static", "./uploads")

	return r
}
