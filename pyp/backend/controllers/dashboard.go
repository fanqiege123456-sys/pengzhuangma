package controllers

import (
	"collision-backend/config"
	"collision-backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardController struct{}

// 获取统计数据
func (ctrl *DashboardController) GetStats(c *gin.Context) {
	type StatsResponse struct {
		UserCount         int64 `json:"userCount"`
		TodayCodeCount    int64 `json:"todayCodeCount"`
		TodaySuccessCount int64 `json:"todaySuccessCount"`
		TodayRevenue      int   `json:"todayRevenue"`
	}

	var stats StatsResponse

	// 获取用户总数
	config.DB.Model(&models.User{}).Count(&stats.UserCount)

	// 获取今日数据
	today := time.Now().Format("2006-01-02")
	todayStart := today + " 00:00:00"
	todayEnd := today + " 23:59:59"

	// 今日碰撞码生成数
	config.DB.Model(&models.CollisionCode{}).
		Where("created_at BETWEEN ? AND ?", todayStart, todayEnd).
		Count(&stats.TodayCodeCount)

	// 今日成功匹配数
	config.DB.Model(&models.CollisionRecord{}).
		Where("status IN ? AND created_at BETWEEN ? AND ?", []string{"matched", "friend_added"}, todayStart, todayEnd).
		Count(&stats.TodaySuccessCount)

	// 今日收入（示例数据，实际应该从充值记录计算）
	var todayRecharge int64
	config.DB.Model(&models.RechargeRecord{}).
		Where("created_at BETWEEN ? AND ?", todayStart, todayEnd).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&todayRecharge)
	stats.TodayRevenue = int(todayRecharge / 100) // 转换为元

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": stats,
	})
}

// 获取热门碰撞码
func (ctrl *DashboardController) GetHotCodes(c *gin.Context) {
	var hotCodes []models.HotTag

	result := config.DB.Order("submit_count DESC").Limit(10).Find(&hotCodes)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取热门碰撞码失败",
		})
		return
	}

	// 转换为前端期望的格式
	type HotCodeResponse struct {
		Code      string `json:"code"`
		Count     int    `json:"count"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
	}

	var response []HotCodeResponse
	for _, keyword := range hotCodes {
		response = append(response, HotCodeResponse{
			Code:      keyword.Keyword,
			Count:     keyword.SubmitCount,
			Status:    keyword.Status,
			CreatedAt: keyword.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": response,
	})
}

// 获取用户注册趋势
func (ctrl *DashboardController) GetUserRegistrationTrend(c *gin.Context) {
	// 获取查询参数，默认为7天
	period := c.DefaultQuery("period", "7")
	days := 7
	if period == "30" {
		days = 30
	}

	// 生成日期序列
	type DateCount struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	var trend []DateCount

	// 计算起始日期
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)

	// 遍历日期，获取每天的注册用户数
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		nextDay := d.AddDate(0, 0, 1).Format("2006-01-02")

		var count int64
		config.DB.Model(&models.User{}).
			Where("created_at >= ? AND created_at < ?", dateStr, nextDay).
			Count(&count)

		trend = append(trend, DateCount{
			Date:  dateStr,
			Count: count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": trend,
	})
}

// 获取碰撞成功率统计
func (ctrl *DashboardController) GetCollisionSuccessRate(c *gin.Context) {
	// 获取查询参数，默认为7天
	period := c.DefaultQuery("period", "7")
	days := 7
	if period == "30" {
		days = 30
	}

	// 生成日期序列
	type DateSuccessRate struct {
		Date         string  `json:"date"`
		TotalCount   int64   `json:"totalCount"`
		SuccessCount int64   `json:"successCount"`
		SuccessRate  float64 `json:"successRate"`
	}

	var successRateData []DateSuccessRate

	// 计算起始日期
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)

	// 遍历日期，获取每天的碰撞成功率
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		nextDay := d.AddDate(0, 0, 1).Format("2006-01-02")

		// 当天总碰撞码数
		var totalCount int64
		config.DB.Model(&models.CollisionCode{}).
			Where("created_at >= ? AND created_at < ?", dateStr, nextDay).
			Count(&totalCount)

		// 当天成功匹配数
		var successCount int64
		config.DB.Model(&models.CollisionRecord{}).
			Where("status IN ? AND created_at >= ? AND created_at < ?", []string{"matched", "friend_added"}, dateStr, nextDay).
			Count(&successCount)

		// 计算成功率
		successRate := 0.0
		if totalCount > 0 {
			successRate = float64(successCount) / float64(totalCount) * 100
		}

		successRateData = append(successRateData, DateSuccessRate{
			Date:         dateStr,
			TotalCount:   totalCount,
			SuccessCount: successCount,
			SuccessRate:  successRate,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": successRateData,
	})
}

// 获取当前审核设置
func (ctrl *DashboardController) GetAuditSetting(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"enableCollisionAudit": config.Config.EnableCollisionAudit,
		},
	})
}

// 更新审核设置
func (ctrl *DashboardController) UpdateAuditSetting(c *gin.Context) {
	var req struct {
		EnableCollisionAudit *bool `json:"enableCollisionAudit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Invalid request data",
		})
		return
	}
	if req.EnableCollisionAudit == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Missing enableCollisionAudit",
		})
		return
	}

	// 更新全局配置
	config.Config.EnableCollisionAudit = *req.EnableCollisionAudit
	if !config.Config.EnableCollisionAudit {
		// 关闭审核时，直接放行待审核的碰撞码
		config.DB.Model(&models.CollisionCode{}).
			Where("audit_status = ?", "pending").
			Updates(map[string]interface{}{
				"audit_status": "approved",
			})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "Audit setting updated successfully",
		"data": gin.H{
			"enableCollisionAudit": *req.EnableCollisionAudit,
		},
	})
}

// 获取审核统计数据
func (ctrl *DashboardController) GetAuditStats(c *gin.Context) {
	type AuditStats struct {
		Pending  int64 `json:"pending"`  // 待审核数量
		Approved int64 `json:"approved"` // 已通过数量
		Rejected int64 `json:"rejected"` // 已拒绝数量
		Total    int64 `json:"total"`    // 总数量
	}

	var stats AuditStats

	// 获取待审核数量
	config.DB.Model(&models.CollisionCode{}).Where("audit_status = ?", "pending").Count(&stats.Pending)
	// 获取已通过数量
	config.DB.Model(&models.CollisionCode{}).Where("audit_status = ?", "approved").Count(&stats.Approved)
	// 获取已拒绝数量
	config.DB.Model(&models.CollisionCode{}).Where("audit_status = ?", "rejected").Count(&stats.Rejected)
	// 获取总数量
	config.DB.Model(&models.CollisionCode{}).Count(&stats.Total)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": stats,
	})
}
