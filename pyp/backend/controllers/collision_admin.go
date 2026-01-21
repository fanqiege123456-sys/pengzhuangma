package controllers

import (
	"net/http"
	"strconv"
	"time"

	"collision-backend/config"
	"collision-backend/models"
	"collision-backend/utils"

	"github.com/gin-gonic/gin"
)

// Admin audit endpoints.
func (cc *CollisionController) GetPendingCollisionCodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")

	var codes []models.CollisionCode
	var total int64

	offset := (page - 1) * pageSize
	query := config.DB.Model(&models.CollisionCode{}).Where("audit_status = ?", "pending")
	if keyword != "" {
		query = query.Where("tag LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)
	query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&codes)
	markForbiddenKeywords(codes)

	var pendingCount int64
	config.DB.Model(&models.CollisionCode{}).Where("audit_status = ?", "pending").Count(&pendingCount)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)

	var todayApproved int64
	var todayRejected int64
	config.DB.Model(&models.CollisionCode{}).Where("audit_status = ? AND audit_at >= ? AND audit_at < ?", "approved", start, end).Count(&todayApproved)
	config.DB.Model(&models.CollisionCode{}).Where("audit_status = ? AND audit_at >= ? AND audit_at < ?", "rejected", start, end).Count(&todayRejected)

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"list": codes,
		"pagination": utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
		"pending_count": pendingCount,
		"today_approved": todayApproved,
		"today_rejected": todayRejected,
	}))
}

func (cc *CollisionController) ApproveCollisionCode(c *gin.Context) {
	adminID := c.GetUint("user_id")
	id := c.Param("id")

	var code models.CollisionCode
	if err := config.DB.First(&code, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Collision code not found"))
		return
	}

	now := time.Now()
	if err := config.DB.Model(&code).Updates(map[string]interface{}{
		"audit_status": "approved",
		"audit_by":     adminID,
		"audit_at":     &now,
		"reject_reason": "",
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to approve collision code"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{"message": "approved"}))
}

func (cc *CollisionController) RejectCollisionCode(c *gin.Context) {
	adminID := c.GetUint("user_id")
	id := c.Param("id")

	var req struct {
		RejectReason string `json:"reject_reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RejectReason == "" {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Missing reject_reason"))
		return
	}

	var code models.CollisionCode
	if err := config.DB.First(&code, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Collision code not found"))
		return
	}

	now := time.Now()
	tx := config.DB.Begin()
	if err := tx.Model(&code).Updates(map[string]interface{}{
		"audit_status": "rejected",
		"audit_by":     adminID,
		"audit_at":     &now,
		"reject_reason": req.RejectReason,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to reject collision code"))
		return
	}

	if code.CostCoins > 0 {
		var user models.User
		if err := tx.First(&user, code.UserID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to load user"))
			return
		}

		if err := tx.Model(&user).Update("coins", user.Coins+code.CostCoins).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to refund coins"))
			return
		}

		consumeRecord := models.ConsumeRecord{
			UserID: code.UserID,
			Coins:  code.CostCoins,
			Type:   "refund",
			Reason: "Collision code rejected: " + code.Tag,
		}
		if err := tx.Create(&consumeRecord).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create refund record"))
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to commit rejection"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{"message": "rejected"}))
}

func (cc *CollisionController) BatchApproveCollisionCodes(c *gin.Context) {
	adminID := c.GetUint("user_id")
	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid ids"))
		return
	}

	now := time.Now()
	if err := config.DB.Model(&models.CollisionCode{}).
		Where("id IN ?", req.IDs).
		Updates(map[string]interface{}{
			"audit_status": "approved",
			"audit_by":     adminID,
			"audit_at":     &now,
			"reject_reason": "",
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to approve codes"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{"message": "approved"}))
}

func (cc *CollisionController) BatchRejectCollisionCodes(c *gin.Context) {
	adminID := c.GetUint("user_id")
	var req struct {
		IDs          []uint `json:"ids"`
		RejectReason string `json:"reject_reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 || req.RejectReason == "" {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request"))
		return
	}

	var codes []models.CollisionCode
	if err := config.DB.Where("id IN ?", req.IDs).Find(&codes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to load codes"))
		return
	}

	now := time.Now()
	tx := config.DB.Begin()
	for _, code := range codes {
		if err := tx.Model(&models.CollisionCode{}).Where("id = ?", code.ID).Updates(map[string]interface{}{
			"audit_status": "rejected",
			"audit_by":     adminID,
			"audit_at":     &now,
			"reject_reason": req.RejectReason,
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to reject codes"))
			return
		}

		if code.CostCoins > 0 {
			var user models.User
			if err := tx.First(&user, code.UserID).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to load user"))
				return
			}
			if err := tx.Model(&user).Update("coins", user.Coins+code.CostCoins).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to refund coins"))
				return
			}
			consumeRecord := models.ConsumeRecord{
				UserID: code.UserID,
				Coins:  code.CostCoins,
				Type:   "refund",
				Reason: "Collision code rejected: " + code.Tag,
			}
			if err := tx.Create(&consumeRecord).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create refund record"))
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to commit rejection"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{"message": "rejected"}))
}

func (cc *CollisionController) BatchApproveAllCollisionCodes(c *gin.Context) {
	adminID := c.GetUint("user_id")
	now := time.Now()
	if err := config.DB.Model(&models.CollisionCode{}).
		Where("audit_status = ?", "pending").
		Updates(map[string]interface{}{
			"audit_status": "approved",
			"audit_by":     adminID,
			"audit_at":     &now,
			"reject_reason": "",
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to approve codes"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{"message": "approved"}))
}
