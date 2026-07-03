package http

import (
	"feedSystem_video/internal/account"
	"feedSystem_video/internal/apierror"
	"feedSystem_video/internal/config"
	"feedSystem_video/internal/feed" //新增
	cors "feedSystem_video/internal/middleware/cors"
	"feedSystem_video/internal/middleware/jwt"
	"feedSystem_video/internal/middleware/ratelimit"
	"feedSystem_video/internal/notification"
	"feedSystem_video/internal/social" //新增
	"feedSystem_video/internal/video"

	"github.com/gin-gonic/gin"
)

// SetupRouter 创建并配置 Gin 路由
func SetupRouter() *gin.Engine {
	r := gin.Default() // 创建 Gin 引擎

	// 设置最大 multipart 内存为 256MB（支持大文件上传）
	r.MaxMultipartMemory = 256 << 20 // 256 MB

	// 全局中间件：CORS 跨域 + 限流（每秒 20 请求，突发上限 40）
	r.Use(cors.CORS())
	r.Use(ratelimit.Middleware(20, 40))

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
		accountGroup.POST("/getProfile", handler.GetProfile)

		// 需要JWT认证的接口
		accountAuthGroup := accountGroup.Use(jwt.JWTAuth())
		accountAuthGroup.POST("/uploadAvatar", handler.UploadAvatar)
		accountAuthGroup.POST("/changePassword", handler.ChangePassword)
		accountAuthGroup.POST("/updateBio", handler.UpdateBio)
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
		videoGroup.POST("/delete", videoHandler.DeleteVideo)
		videoGroup.POST("/deleteBatch", videoHandler.DeleteVideosBatch)
		videoGroup.POST("/uploadCover", videoHandler.UploadCover)
	}

	// 不需要 JWT 的接口（放在外面）
	r.POST("/video/listByAuthorID", videoHandler.ListByAuthor)
	// 不需要登录，即可搜索视频
	r.POST("/video/getDetail", videoHandler.GetDetail)
	// 记录视频播放
	r.POST("/video/recordPlay", videoHandler.RecordPlay)
	// 获取热门视频列表
	r.POST("/video/listHot", videoHandler.ListHotVideos)
	// 搜索视频
	r.POST("/video/search", videoHandler.SearchVideos) //新增

	// 静态文件服务（让上传的视频可以通过 URL 访问，路径从配置读取）
	r.Static(config.C.Storage.StaticPath, config.C.Storage.UploadDir)

	// --- Like 模块 ---
	likeGroup := r.Group("/like")
	{
		likeGroup.Use(jwt.JWTAuth())
		likeGroup.POST("/like", videoHandler.Like)
		likeGroup.POST("/unlike", videoHandler.Unlike)
		likeGroup.POST("/isLiked", videoHandler.IsLiked)
	}

	// --- Comment 模块 ---
	commentGroup := r.Group("/comment")
	{
		commentGroup.POST("/listAll", videoHandler.ListComments) // 不需要登录
		commentGroup.Use(jwt.JWTAuth())
		commentGroup.POST("/publish", videoHandler.PublishComment)
	}

	// --- Social 模块 ---
	socialRepo := social.NewRepo()
	socialService := social.NewService(socialRepo, repo)
	socialHandler := social.NewHandler(socialService)

	socialGroup := r.Group("/social")
	{
		socialGroup.Use(jwt.JWTAuth())
		socialGroup.POST("/follow", socialHandler.Follow)
		socialGroup.POST("/unfollow", socialHandler.Unfollow)
		socialGroup.POST("/isFollowing", socialHandler.IsFollowing)
		socialGroup.POST("/getAllFollowers", socialHandler.GetFollowers)
		socialGroup.POST("/getAllVloggers", socialHandler.GetVloggers)
		socialGroup.POST("/getCounts", socialHandler.GetCounts)
	}

	// --- Feed 模块 ---
	feedRepo := feed.NewRepo()
	feedService := feed.NewService(feedRepo)
	feedHandler := feed.NewHandler(feedService)

	feedGroup := r.Group("/feed")
	{
		feedGroup.POST("/listLatest", feedHandler.ListLatest)
	}

	// --- Notification 模块 ---
	notifRepo := notification.NewRepo()
	notifHandler := notification.NewHandler(notifRepo)

	notifGroup := r.Group("/notification")
	{
		notifGroup.Use(jwt.JWTAuth())
		notifGroup.POST("/unreadCount", notifHandler.UnreadCount)
		notifGroup.POST("/list", notifHandler.List)
	}

	return r
}
