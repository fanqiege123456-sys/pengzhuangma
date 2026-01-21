package controllers

import (
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"collision-backend/config"
	"collision-backend/models"
	"collision-backend/services"
	"collision-backend/utils"

	"github.com/gin-gonic/gin"
)

// User collision code helpers.
func (cc *CollisionController) GetMyCollisionCodes(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var codes []models.CollisionCode
	var total int64
	config.DB.Model(&models.CollisionCode{}).Where("user_id = ?", userID).Count(&total)
	config.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&codes)

	now := time.Now()
	result := make([]gin.H, 0, len(codes))
	for _, code := range codes {
		isExpired := code.ExpiresAt.Before(now) || code.Status == "expired"
		collisionStatus := "进行中"
		if code.IsMatched || code.MatchCount > 0 {
			collisionStatus = "已碰撞"
		}

		timeLeftText := "已过期"
		timeLeftSeconds := int64(0)
		if !isExpired {
			timeLeft := code.ExpiresAt.Sub(now)
			if timeLeft < 0 {
				timeLeft = 0
			}
			timeLeftSeconds = int64(timeLeft.Seconds())
			totalMinutes := int(timeLeft.Minutes())
			hours := totalMinutes / 60
			minutes := totalMinutes % 60
			switch {
			case hours > 0 && minutes > 0:
				timeLeftText = fmt.Sprintf("%d小时%d分钟", hours, minutes)
			case hours > 0:
				timeLeftText = fmt.Sprintf("%d小时", hours)
			case minutes > 0:
				timeLeftText = fmt.Sprintf("%d分钟", minutes)
			default:
				timeLeftText = "少于1分钟"
			}
		}

		result = append(result, gin.H{
			"id":               code.ID,
			"tag":              code.Tag,
			"country":          code.Country,
			"province":         code.Province,
			"city":             code.City,
			"district":         code.District,
			"gender":           code.Gender,
			"age_min":          code.AgeMin,
			"age_max":          code.AgeMax,
			"status":           code.Status,
			"audit_status":     code.AuditStatus,
			"cost_coins":       code.CostCoins,
			"expires_at":       code.ExpiresAt,
			"created_at":       code.CreatedAt,
			"updated_at":       code.UpdatedAt,
			"match_count":      code.MatchCount,
			"is_matched":       code.IsMatched,
			"is_expired":       isExpired,
			"collision_status": collisionStatus,
			"time_left":        timeLeftText,
			"time_left_seconds": timeLeftSeconds,
		})
	}
c.JSON(http.StatusOK, utils.Success(gin.H{
		"codes": result,
		"total": total,
	}))
}

func (cc *CollisionController) GetMyCollisionCodeByID(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	id := c.Param("id")
	var code models.CollisionCode
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&code).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Collision code not found"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(code))
}

func (cc *CollisionController) UpdateMyCollisionCode(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	id := c.Param("id")
	var req struct {
		Tag       string `json:"tag"`
		Days      int    `json:"days"`
		CostCoins int    `json:"cost_coins"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request"))
		return
	}

	var code models.CollisionCode
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&code).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Collision code not found"))
		return
	}

	updates := map[string]interface{}{}
	if req.Tag != "" && req.Tag != code.Tag {
		updates["tag"] = req.Tag
		updates["audit_status"] = defaultAuditStatus()
		updates["reject_reason"] = ""
		updates["audit_at"] = nil
		updates["audit_by"] = 0
	}
	if req.Days > 0 {
		updates["expires_at"] = time.Now().Add(time.Duration(req.Days) * 24 * time.Hour)
		updates["status"] = "active"
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, utils.Error(400, "No changes"))
		return
	}

	tx := config.DB.Begin()
	if req.CostCoins > 0 {
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to load user"))
			return
		}
		if user.Coins < req.CostCoins {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, utils.Error(400, "Insufficient coins"))
			return
		}
		if err := tx.Model(&user).Update("coins", user.Coins-req.CostCoins).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to deduct coins"))
			return
		}
		consumeRecord := models.ConsumeRecord{
			UserID: userID.(uint),
			Coins:  req.CostCoins,
			Type:   "renew_collision",
			Reason: "Update collision code: " + code.Tag,
		}
		if err := tx.Create(&consumeRecord).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create consume record"))
			return
		}
	}

	if err := tx.Model(&code).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to update collision code"))
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to commit update"))
		return
	}

	if _, ok := updates["tag"]; ok {
		matcher := services.NewCollisionMatcher()
		matcher.MatchForCode(&code)
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{"message": "updated"}))
}

func (cc *CollisionController) RenewCollisionCode(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	id := c.Param("id")
	var code models.CollisionCode
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&code).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Collision code not found"))
		return
	}

	const costCoins = 10
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to load user"))
		return
	}
	if user.Coins < costCoins {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Insufficient coins"))
		return
	}

	now := time.Now()
	newExpire := now.Add(24 * time.Hour)
	if code.ExpiresAt.After(now) {
		newExpire = code.ExpiresAt.Add(24 * time.Hour)
	}

	tx := config.DB.Begin()
	if err := tx.Model(&user).Update("coins", user.Coins-costCoins).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to deduct coins"))
		return
	}

	consumeRecord := models.ConsumeRecord{
		UserID: userID.(uint),
		Coins:  costCoins,
		Type:   "renew_collision",
		Reason: "Renew collision code: " + code.Tag,
	}
	if err := tx.Create(&consumeRecord).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create consume record"))
		return
	}

	if err := tx.Model(&code).Updates(map[string]interface{}{
		"expires_at": newExpire,
		"status":     "active",
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to renew collision code"))
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to commit renewal"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{"message": "renewed"}))
}

func (cc *CollisionController) ResubmitCollisionCode(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	id := c.Param("id")
	var code models.CollisionCode
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&code).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Collision code not found"))
		return
	}

	const costCoins = 10
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to load user"))
		return
	}
	if user.Coins < costCoins {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Insufficient coins"))
		return
	}

	tx := config.DB.Begin()
	if err := tx.Model(&user).Update("coins", user.Coins-costCoins).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to deduct coins"))
		return
	}

	consumeRecord := models.ConsumeRecord{
		UserID: userID.(uint),
		Coins:  costCoins,
		Type:   "collision_submit",
		Reason: "Resubmit collision code: " + code.Tag,
	}
	if err := tx.Create(&consumeRecord).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create consume record"))
		return
	}

	if err := tx.Model(&code).Updates(map[string]interface{}{
		"status":       "active",
		"audit_status": defaultAuditStatus(),
		"audit_at":     nil,
		"audit_by":     0,
		"reject_reason": "",
		"expires_at":   time.Now().Add(24 * time.Hour),
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to resubmit collision code"))
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to commit resubmit"))
		return
	}

	matcher := services.NewCollisionMatcher()
	matcher.MatchForCode(&code)

	c.JSON(http.StatusOK, utils.Success(gin.H{"message": "resubmitted"}))
}

func (cc *CollisionController) DeleteMyCollisionCode(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	id := c.Param("id")
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.CollisionCode{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to delete collision code"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{"message": "deleted"}))
}

func (cc *CollisionController) SendEmailToMatchedUser(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	var req struct {
		MatchedUserID uint64 `json:"matched_user_id"`
		Keyword       string `json:"keyword"`
		Message       string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.MatchedUserID == 0 || req.Keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var collisionResult models.CollisionResult
	if err := config.DB.Where("user_id = ? AND matched_user_id = ? AND keyword = ?", userID, req.MatchedUserID, req.Keyword).
		Order("created_at DESC").
		First(&collisionResult).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "碰撞记录不存在"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败"})
		return
	}
	if user.Coins < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "积分不足"})
		return
	}

	var matchedContact models.UserContact
	if err := config.DB.Where("user_id = ?", collisionResult.MatchedUserID).First(&matchedContact).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "对方未绑定邮箱"})
		return
	}
	if matchedContact.Email == "" || !matchedContact.EmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "对方邮箱未验证"})
		return
	}

	subject := "小程序匹配成功，用户给你发信息啦"
	content := strings.TrimSpace(req.Message)
	if content == "" {
		content = "您好，我是通过碰撞交友认识您的，很高兴认识您！"
	}
	if len([]rune(content)) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "邮件内容最多500个字"})
		return
	}

	htmlBody := "<p>小程序匹配成功，用户给你发信息啦：</p><p>" + html.EscapeString(content) + "</p>"
	emailService := services.NewSMTPEmailService(config.DB)
	if err := emailService.SendEmail(uint64(userID), matchedContact.Email, subject, htmlBody, "collision"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "邮件发送失败: " + err.Error()})
		return
	}

	if err := config.DB.Model(&user).Update("coins", user.Coins-1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "积分扣除失败"})
		return
	}

	_ = config.DB.Model(&collisionResult).Updates(map[string]interface{}{
		"email_sent":    true,
		"email_sent_at": time.Now(),
	}).Error

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "邮件发送成功",
		"data": gin.H{
			"remaining_coins": user.Coins - 1,
		},
	})
}
