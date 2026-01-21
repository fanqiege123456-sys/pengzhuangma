package middlewares

import (
	"net/http"
	"strings"

	"collision-backend/utils"

	"github.com/gin-gonic/gin"
)

// JWT中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, utils.Error(401, "Token required"))
			c.Abort()
			return
		}

		// 移除Bearer前缀
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.Error(401, "Invalid token"))
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// 管理员权限中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		roleStr, ok := role.(string)
		if !exists || !ok || (roleStr != "admin" && roleStr != "super") {
			c.JSON(http.StatusForbidden, utils.Error(403, "Admin permission required"))
			c.Abort()
			return
		}
		c.Next()
	}
}
