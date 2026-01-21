package controllers

import (
	"collision-backend/config"
	"collision-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RecordController struct{}

// 获取碰撞记录列表
func (ctrl *RecordController) GetRecords(c *gin.Context) {
	var records []models.CollisionRecord

	// 预加载相关的用户信息
	result := config.DB.Preload("User1").Preload("User2").
		Order("add_friend_deadline DESC").Find(&records)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取碰撞记录失败",
		})
		return
	}

	// 转换数据格式以匹配前端期望的结构
	type RecordResponse struct {
		ID                uint         `json:"id"`
		Tag               string       `json:"tag"`        // 兴趣标签
		MatchType         string       `json:"match_type"` // 匹配类型
		User1             *models.User `json:"user1"`
		User2             *models.User `json:"user2"`
		Status            string       `json:"status"`
		AddFriendDeadline string       `json:"add_friend_deadline"`
		EmailSent         bool         `json:"email_sent"`    // 邮件是否已发送
		EmailStatus       string       `json:"email_status"`  // 邮件发送状态
		EmailSentAt       string       `json:"email_sent_at"` // 邮件发送时间
		CreatedAt         string       `json:"created_at"`
		// 添加匹配用户的详细信息
		User1Email string `json:"user1_email"` // 用户1的邮箱
		User2Email string `json:"user2_email"` // 用户2的邮箱
	}

	var response []RecordResponse
	for _, record := range records {
		// 获取用户1和用户2的最新邮件日志
		var user1EmailLog models.EmailLog
		var user2EmailLog models.EmailLog

		// 查询用户1的最新邮件日志
		config.DB.Where("user_id = ? AND subject LIKE ?", record.UserID1, "%"+record.Tag+"%").
			Order("created_at DESC").First(&user1EmailLog)

		// 查询用户2的最新邮件日志
		config.DB.Where("user_id = ? AND subject LIKE ?", record.UserID2, "%"+record.Tag+"%").
			Order("created_at DESC").First(&user2EmailLog)

		// 合并邮件状态信息
		emailSent := false
		emailStatus := ""
		emailSentAt := ""

		// 优先使用用户1的邮件状态
		if user1EmailLog.ID > 0 {
			emailSent = user1EmailLog.Status == "sent"
			emailStatus = user1EmailLog.Status
			if user1EmailLog.SentAt != nil {
				emailSentAt = user1EmailLog.SentAt.Format("2006-01-02 15:04:05")
			}
		}

		// 如果用户1没有邮件日志，使用用户2的
		if emailSentAt == "" && user2EmailLog.ID > 0 {
			emailSent = user2EmailLog.Status == "sent"
			emailStatus = user2EmailLog.Status
			if user2EmailLog.SentAt != nil {
				emailSentAt = user2EmailLog.SentAt.Format("2006-01-02 15:04:05")
			}
		}

		// 获取用户联系方式
		var user1Contact models.UserContact
		var user2Contact models.UserContact
		config.DB.Where("user_id = ?", record.UserID1).First(&user1Contact)
		config.DB.Where("user_id = ?", record.UserID2).First(&user2Contact)

		response = append(response, RecordResponse{
			ID:                record.ID,
			Tag:               record.Tag,       // 更新为使用Tag
			MatchType:         record.MatchType, // 添加匹配类型
			User1:             &record.User1,
			User2:             &record.User2,
			Status:            record.Status,
			AddFriendDeadline: record.AddFriendDeadline.Format("2006-01-02 15:04:05"),
			EmailSent:         emailSent,
			EmailStatus:       emailStatus,
			EmailSentAt:       emailSentAt,
			CreatedAt:         record.CreatedAt.Format("2006-01-02 15:04:05"),
			User1Email:        user1Contact.Email,
			User2Email:        user2Contact.Email,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": response,
	})
}
