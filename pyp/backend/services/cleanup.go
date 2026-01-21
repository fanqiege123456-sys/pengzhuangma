package services

import (
	"log"
	"time"

	"collision-backend/config"
	"collision-backend/models"
)

type CleanupService struct{}

// 开始定时清理任务
func (cs *CleanupService) StartCleanupTasks() {
	// 每10分钟清理一次过期碰撞码
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			cs.CleanupExpiredCodes()
		}
	}()

	// 每30分钟检查一次过期的匹配记录
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			cs.ProcessExpiredMatches()
		}
	}()

	// 每日重置24小时热门标签计数（每天0点执行）
	go func() {
		// 计算下一次执行时间（当天0点）
		now := time.Now()
		nextReset := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		duration := nextReset.Sub(now)
		
		// 初始延迟
		time.Sleep(duration)
		
		// 执行第一次重置
		cs.ResetHotTags24h()
		
		// 之后每天执行一次
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		
		for range ticker.C {
			cs.ResetHotTags24h()
		}
	}()

	log.Println("Cleanup service started")
}

// 清理过期的碰撞码和碰撞列表
func (cs *CleanupService) CleanupExpiredCodes() {
	now := time.Now()

	// 更新过期的碰撞码状态
	result := config.DB.Model(&models.CollisionCode{}).
		Where("status = ? AND expires_at < ?", "active", now).
		Update("status", "expired")

	if result.Error != nil {
		log.Printf("Error cleaning up expired codes: %v", result.Error)
		return
	}

	if result.RowsAffected > 0 {
		log.Printf("Cleaned up %d expired collision codes", result.RowsAffected)
	}

	// 同时更新过期的碰撞列表状态
	result2 := config.DB.Model(&models.CollisionList{}).
		Where("status = ? AND expire_at < ?", "active", now).
		Update("status", "expired")

	if result2.Error != nil {
		log.Printf("Error cleaning up expired collision lists: %v", result2.Error)
		return
	}

	if result2.RowsAffected > 0 {
		log.Printf("Cleaned up %d expired collision lists", result2.RowsAffected)
	}
}

// 处理过期的匹配记录
func (cs *CleanupService) ProcessExpiredMatches() {
	now := time.Now()

	// 查找已过期但状态仍为matched的记录
	var expiredRecords []models.CollisionRecord
	err := config.DB.Where("status = ? AND add_friend_deadline < ?", "matched", now).
		Preload("User1").
		Preload("User2").
		Find(&expiredRecords).Error

	if err != nil {
		log.Printf("Error fetching expired matches: %v", err)
		return
	}

	if len(expiredRecords) == 0 {
		return
	}

	log.Printf("Processing %d expired matches", len(expiredRecords))

	for _, record := range expiredRecords {
		cs.processExpiredMatch(record)
	}
}

// 处理单个过期匹配记录
func (cs *CleanupService) processExpiredMatch(record models.CollisionRecord) {
	// 检查双方的AllowPassiveAdd设置
	user1AllowsPassive := record.User1.AllowPassiveAdd
	user2AllowsPassive := record.User2.AllowPassiveAdd

	// 如果双方都允许被动添加，自动添加为好友
	if user1AllowsPassive && user2AllowsPassive {
		cs.autoAddFriends(record)
	} else {
		// 否则标记为错过
		cs.markAsMissed(record)
	}
}

// 自动添加好友
func (cs *CleanupService) autoAddFriends(record models.CollisionRecord) {
	tx := config.DB.Begin()

	// 检查是否已经是好友
	var existingFriend models.Friend
	err := tx.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		record.UserID1, record.UserID2, record.UserID2, record.UserID1).
		First(&existingFriend).Error

	if err == nil {
		// 已经是好友，只更新匹配记录状态
		tx.Model(&record).Update("status", "friend_added")
		tx.Commit()
		log.Printf("Match %d: Users %d and %d are already friends",
			record.ID, record.UserID1, record.UserID2)
		return
	}

	// 创建双向好友关系
	friend1 := models.Friend{
		UserID:   record.UserID1,
		FriendID: record.UserID2,
		Status:   "accepted",
	}

	friend2 := models.Friend{
		UserID:   record.UserID2,
		FriendID: record.UserID1,
		Status:   "accepted",
	}

	if err := tx.Create(&friend1).Error; err != nil {
		tx.Rollback()
		log.Printf("Error creating friend relationship 1 for match %d: %v", record.ID, err)
		return
	}

	if err := tx.Create(&friend2).Error; err != nil {
		tx.Rollback()
		log.Printf("Error creating friend relationship 2 for match %d: %v", record.ID, err)
		return
	}

	// 更新匹配记录状态
	if err := tx.Model(&record).Update("status", "friend_added").Error; err != nil {
		tx.Rollback()
		log.Printf("Error updating match status for match %d: %v", record.ID, err)
		return
	}

	tx.Commit()
	log.Printf("Auto-added friends for match %d: users %d and %d",
		record.ID, record.UserID1, record.UserID2)
}

// 标记为错过
func (cs *CleanupService) markAsMissed(record models.CollisionRecord) {
	err := config.DB.Model(&record).Update("status", "missed").Error
	if err != nil {
		log.Printf("Error marking match %d as missed: %v", record.ID, err)
		return
	}

	log.Printf("Marked match %d as missed: users %d and %d",
		record.ID, record.UserID1, record.UserID2)
}

// 重置24小时热门标签计数
func (cs *CleanupService) ResetHotTags24h() {
	log.Println("开始重置24小时热门标签计数...")
	
	result := config.DB.Model(&models.HotTag{}).Update("count_24h", 0)
	if result.Error != nil {
		log.Printf("重置24小时热门标签计数失败: %v", result.Error)
		return
	}
	
	log.Printf("成功重置 %d 个热门标签的24小时计数", result.RowsAffected)
}

// 手动触发清理（用于测试或立即清理）
func (cs *CleanupService) ManualCleanup() {
	log.Println("Starting manual cleanup...")
	cs.CleanupExpiredCodes()
	cs.ProcessExpiredMatches()
	log.Println("Manual cleanup completed")
}
