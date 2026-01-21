package middlewares

import (
	"github.com/gin-gonic/gin"
)

// CORS中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求来源
		origin := c.Request.Header.Get("Origin")
		
		// 如果有来源，则允许该来源，否则允许所有
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
