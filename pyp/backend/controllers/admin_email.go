package controllers

import (
	"net/http"
	"strconv"

	"collision-backend/config"
	"collision-backend/models"

	"github.com/gin-gonic/gin"
)

func requireAdminRole(c *gin.Context) bool {
	roleValue, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "管理员权限不足"})
		return false
	}

	role, ok := roleValue.(string)
	if !ok || (role != "admin" && role != "super") {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "管理员权限不足"})
		return false
	}

	return true
}

// 后台邮件配置管理接口
// GET /admin/api/email/config
func GetEmailConfig(c *gin.Context) {
	var cfg models.SystemConfig

	// 查找邮件配置
	if err := config.DB.Where("config_key = ?", "email_config").First(&cfg).Error; err != nil {
		// 配置不存在，返回默认值
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"access_key":   "",
				"account":      "",
				"account_name": "",
				"region":       "cn-hangzhou",
				"configured":   false,
			},
		})
		return
	}

	// 解析配置
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"access_key":   cfg.GetValue("access_key"),
			"account":      cfg.GetValue("account"),
			"account_name": cfg.GetValue("account_name"),
			"region":       cfg.GetValue("region"),
			"configured":   cfg.GetValue("access_secret") != "",
		},
	})
}

// POST /admin/api/email/config
func SaveEmailConfig(c *gin.Context) {
	// 检查管理员权限
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	if !requireAdminRole(c) {
		return
	}

	var body struct {
		AccessKey    string `json:"access_key" binding:"required"`
		AccessSecret string `json:"access_secret" binding:"required"`
		Account      string `json:"account" binding:"required"`
		AccountName  string `json:"account_name"`
		Region       string `json:"region"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if body.AccountName == "" {
		body.AccountName = "标签碰撞"
	}
	if body.Region == "" {
		body.Region = "cn-hangzhou"
	}

	// 查找或创建配置
	var cfg models.SystemConfig
	if err := config.DB.Where("config_key = ?", "email_config").First(&cfg).Error; err != nil {
		cfg = models.SystemConfig{
			ConfigKey: "email_config",
		}
	}

	// 设置配置值
	cfg.SetValues(map[string]string{
		"access_key":    body.AccessKey,
		"access_secret": body.AccessSecret,
		"account":       body.Account,
		"account_name":  body.AccountName,
		"region":        body.Region,
	})

	if err := config.DB.Save(&cfg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "保存成功",
	})
}

// GET /admin/api/email/logs
// 获取邮件日志列表
func GetEmailLogs(c *gin.Context) {
	// 检查管理员权限
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	if !requireAdminRole(c) {
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	status := c.Query("status")
	emailType := c.Query("type")

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 构建查询
	query := config.DB.Model(&models.EmailLog{}).Order("created_at DESC")

	// 添加搜索条件
	if keyword != "" {
		query = query.Where("subject LIKE ? OR to_email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if emailType != "" {
		query = query.Where("type = ?", emailType)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取列表数据
	var emailLogs []models.EmailLog
	query.Offset(offset).Limit(pageSize).Find(&emailLogs)

	// 返回数据
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list": emailLogs,
			"pagination": gin.H{
				"page":     page,
				"page_size": pageSize,
				"total":    total,
			},
		},
	})
}
