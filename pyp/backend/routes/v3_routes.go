package routes

import (
	"collision-backend/controllers"
	"collision-backend/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterV3Routes(r *gin.Engine) {
	// 只保留admin相关路由，移除重复的API路由
	admin := r.Group("/admin/api").Use(middlewares.JWTAuth(), middlewares.AdminAuth())
	{
		admin.GET("/email/config", controllers.GetEmailConfig)
		admin.POST("/email/config", controllers.SaveEmailConfig)
		admin.GET("/email/logs", controllers.GetEmailLogs)
	}
}
