package routes

import (
	"collision-backend/controllers"
	"collision-backend/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// CORS中间件已在main.go全局注册

	// 为前端添加兼容路由，处理带api/前缀的请求
	r.GET("/api/admin/api/email/logs", middlewares.JWTAuth(), middlewares.AdminAuth(), controllers.GetEmailLogs)

	// API路由组
	api := r.Group("/api")

	// 微信小程序用户路由（不需要认证）
	userController := &controllers.UserController{}
	collisionUserController := &controllers.CollisionController{}
	user := api.Group("/user")
	{
		user.POST("/login", userController.WechatLogin)

		// 需要认证的用户路由
		userAuth := user.Use(middlewares.JWTAuth())
		{
			userAuth.GET("/info", userController.GetUserInfo)
			userAuth.PUT("/profile", userController.UpdateProfile)
			userAuth.GET("/balance", userController.GetBalance)
		userAuth.GET("/consume-records", userController.GetConsumeRecords)   // 获取消费记录
		userAuth.GET("/recharge-records", userController.GetRechargeRecords) // 获取充值记录
		userAuth.PUT("/location", userController.UpdateUserLocation)         // 新增地址更新接口
		userAuth.GET("/collision-codes/:id", collisionUserController.GetMyCollisionCodeByID)
		userAuth.PUT("/collision-codes/:id", collisionUserController.UpdateMyCollisionCode)
	}
	}

	// 碰撞相关路由（需要用户认证）
	collision := api.Group("/collision").Use(middlewares.JWTAuth())
	{
		collision.POST("/submit", collisionUserController.SubmitCode)
		collision.POST("/batch-submit", collisionUserController.BatchSubmitCodes)
		collision.GET("/matches", collisionUserController.GetMatches)
		collision.GET("/matches/:id", collisionUserController.GetMatchDetail)
		collision.GET("/hot-codes", collisionUserController.GetHotCodes)
		collision.GET("/my-code", collisionUserController.GetMyCollisionCode)
		collision.GET("/my-codes", collisionUserController.GetMyCollisionCodes)                 // 获取所有碰撞码
		collision.POST("/my-codes/:id/renew", collisionUserController.RenewCollisionCode)       // 续费碰撞码
		collision.POST("/my-codes/:id/resubmit", collisionUserController.ResubmitCollisionCode) // 重新提交碰撞码
		collision.DELETE("/my-codes/:id", collisionUserController.DeleteMyCollisionCode)
		collision.POST("/search", collisionUserController.SearchCollisionCodes)
		collision.POST("/add-friend", collisionUserController.AddFriend)
		collision.POST("/send-friend-request", collisionUserController.SendFriendRequest)
		collision.POST("/force-add-friend", collisionUserController.ForceAddFriend)
		collision.POST("/haidilao", collisionUserController.Haidilao)
		collision.POST("/send-email", collisionUserController.SendEmailToMatchedUser) // 新增发送邮件给匹配用户的API
	}

	// 用户地址管理路由（需要用户认证）
	locationController := &controllers.LocationController{}
	locations := api.Group("/locations").Use(middlewares.JWTAuth())
	{
		locations.GET("", locationController.GetLocations)
		locations.POST("", locationController.CreateLocation)
		locations.PUT("/:id", locationController.UpdateLocation)
		locations.DELETE("/:id", locationController.DeleteLocation)
		locations.PUT("/:id/default", locationController.SetDefaultLocation)
	}

	// 管理员路由
	adminController := &controllers.AdminController{}
	admin := api.Group("/admin")
	{
		admin.POST("/login", adminController.Login)

		// 需要认证的路由
		adminAuth := admin.Use(middlewares.JWTAuth(), middlewares.AdminAuth())
		{
			adminAuth.GET("/list", adminController.GetAdmins)
			adminAuth.POST("/create", adminController.CreateAdmin)
			adminAuth.PUT("/:id/status", adminController.UpdateAdminStatus)
			adminAuth.DELETE("/:id", adminController.DeleteAdmin)
		}
	}

	// 用户管理路由（管理员用）
	adminUserController := &controllers.UserController{}
	users := api.Group("/users").Use(middlewares.JWTAuth(), middlewares.AdminAuth())
	{
		users.GET("", adminUserController.GetUsers)
		users.POST("", adminUserController.CreateUser)
		users.GET("/:id", adminUserController.GetUser)
		users.PUT("/:id", adminUserController.UpdateUser)
		users.DELETE("/:id", adminUserController.DeleteUser)
	}

	// ========== V3.0 新增路由 ==========

	// 热门标签（无需认证）
	api.GET("/hot-tags/24h", controllers.GetHotTags24h)
	api.GET("/hot-tags/all", controllers.GetHotTagsAll)
	api.POST("/hot-tags/click", controllers.ClickHotTag) // 新增标签点击API

	// 碰撞列表管理（需要认证）
	collisionListsAuth := api.Group("/collision-lists").Use(middlewares.JWTAuth())
	{
		collisionListsAuth.POST("", controllers.CreateCollisionList)
		collisionListsAuth.GET("", controllers.GetCollisionLists)
		collisionListsAuth.PUT("/:id", controllers.UpdateCollisionList)
		collisionListsAuth.DELETE("/:id", controllers.DeleteCollisionList)
	}

	// 碰撞结果（需要认证）
	collisionResultsAuth := api.Group("/collision-results").Use(middlewares.JWTAuth())
	{
		collisionResultsAuth.GET("", controllers.GetCollisionResults)
		collisionResultsAuth.GET("/:id/detail", controllers.GetCollisionResultDetail)
		collisionResultsAuth.PUT("/:id/remark", controllers.UpdateMatchRemark)
		collisionResultsAuth.POST("/:id/mark-known", controllers.MarkCollisionResultKnown)
		collisionResultsAuth.POST("/send-email", controllers.SendEmailToMatch)
		collisionResultsAuth.POST("/common-keywords", controllers.GetCommonKeywords) // 获取共同碰撞关键词
	}

	// 用户联系方式（需要认证）
	userContactAuth := api.Group("/user").Use(middlewares.JWTAuth())
	{
		userContactAuth.GET("/contacts", controllers.GetUserContacts)
		userContactAuth.POST("/email/bind", controllers.BindEmail)
		userContactAuth.POST("/email/verify", controllers.VerifyEmail)
		userContactAuth.PUT("/email/visibility", controllers.UpdateEmailVisibility)
		userContactAuth.POST("/phone/bind", controllers.BindPhone)
	}

	// 火花相关路由（需要用户认证）
	sparkController := &controllers.SparkController{}
	spark := api.Group("/collision-sparks").Use(middlewares.JWTAuth())
	{
		spark.GET("", sparkController.GetCollisionSparks)
	}

	// ========== V3.0 新增路由结束 ==========

	// 碰撞码管理路由
	collisionController := &controllers.CollisionController{}
	collisions := api.Group("/collisions").Use(middlewares.JWTAuth(), middlewares.AdminAuth())
	{
		collisions.GET("", collisionController.GetCodes)
		collisions.POST("", collisionController.CreateCode)
		collisions.PUT("/:id/status", collisionController.UpdateCodeStatus)
		collisions.PUT("/batch-approve", collisionController.BatchApproveCollisionCodes)
		collisions.PUT("/batch-reject", collisionController.BatchRejectCollisionCodes)
		collisions.PUT("/batch-approve-all", collisionController.BatchApproveAllCollisionCodes)
		collisions.DELETE("/:id", collisionController.DeleteCode)
		// 新增审核相关路由
		collisions.GET("/pending", collisionController.GetPendingCollisionCodes) // 获取待审核碰撞码
		collisions.PUT("/:id/approve", collisionController.ApproveCollisionCode) // 审核通过
		collisions.PUT("/:id/reject", collisionController.RejectCollisionCode)   // 审核拒绝
	}

	// 热门关键词管理路由
	keywordController := &controllers.KeywordController{}
	keywords := api.Group("/keywords").Use(middlewares.JWTAuth(), middlewares.AdminAuth())
	{
		keywords.GET("", keywordController.GetKeywords)
		keywords.POST("", keywordController.CreateKeyword)
		keywords.PUT("/:id/status", keywordController.UpdateKeywordStatus)
		keywords.DELETE("/:id", keywordController.DeleteKeyword)
	}

	// 违禁词管理路由
	forbiddenController := &controllers.ForbiddenKeywordController{}
	forbidden := api.Group("/forbidden-keywords").Use(middlewares.JWTAuth(), middlewares.AdminAuth())
	{
		forbidden.GET("", forbiddenController.GetForbiddenKeywords)
		forbidden.POST("", forbiddenController.CreateForbiddenKeyword)
		forbidden.DELETE("/:id", forbiddenController.DeleteForbiddenKeyword)
	}

	// 碰撞记录路由
	recordController := &controllers.RecordController{}
	records := api.Group("/records").Use(middlewares.JWTAuth(), middlewares.AdminAuth())
	{
		records.GET("", recordController.GetRecords)
	}

	// 仪表盘路由
	dashboardController := &controllers.DashboardController{}
	dashboard := api.Group("/dashboard").Use(middlewares.JWTAuth(), middlewares.AdminAuth())
	{
		dashboard.GET("/stats", dashboardController.GetStats)
		dashboard.GET("/hot-codes", dashboardController.GetHotCodes)
		dashboard.GET("/user-trend", dashboardController.GetUserRegistrationTrend)
		dashboard.GET("/success-rate", dashboardController.GetCollisionSuccessRate)
		// 新增审核设置路由
		dashboard.GET("/audit-setting", dashboardController.GetAuditSetting)    // 获取审核设置
		dashboard.PUT("/audit-setting", dashboardController.UpdateAuditSetting) // 更新审核设置
		dashboard.GET("/audit-stats", dashboardController.GetAuditStats)        // 获取审核统计数据
	}

	recharge := api.Group("/recharge").Use(middlewares.JWTAuth())
	{
		recharge.POST("/create", userController.CreateRechargeOrder)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}
