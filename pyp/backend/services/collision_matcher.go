package services

import (
	"collision-backend/config"
	"collision-backend/models"
	"log"
	"time"

	"gorm.io/gorm"
)

// CollisionMatcher 碰撞匹配服务
type CollisionMatcher struct{}

// NewCollisionMatcher 创建碰撞匹配服务实例
func NewCollisionMatcher() *CollisionMatcher {
	return &CollisionMatcher{}
}

// MatchForCode 立即为指定碰撞码执行匹配
func (cm *CollisionMatcher) MatchForCode(code *models.CollisionCode) int {
	if code == nil {
		return 0
	}
	return cm.findAllMatches(code)
}

// RunMatcher 运行匹配逻辑（定期调用）
func (cm *CollisionMatcher) RunMatcher() {
	log.Println("========== 开始碰撞匹配任务 ==========")
	startTime := time.Now()

	// 清理无效的碰撞码(用户不存在的)
	cm.cleanInvalidCodes()

	// 获取所有有效的碰撞码（支持多对多匹配，不再用 is_matched 过滤）
	var activeCodes []models.CollisionCode
	if err := config.DB.
		Where("status != ?", "invalid").
		Preload("User").
		Find(&activeCodes).Error; err != nil {
		log.Printf("获取活跃碰撞码失败: %v", err)
		return
	}

	log.Printf("找到 %d 个活跃的碰撞码", len(activeCodes))

	matchCount := 0
	for _, code := range activeCodes {
		// 跳过用户数据未加载的碰撞码
		if code.User.ID == 0 {
			log.Printf("⚠️ 碰撞码#%d的用户数据未加载,跳过", code.ID)
			continue
		}

		// 为每个碰撞码寻找所有可能的匹配（多对多）
		matches := cm.findAllMatches(&code)
		matchCount += matches
	}

	elapsed := time.Since(startTime)
	log.Printf("========== 碰撞匹配任务完成 ==========")
	log.Printf("总耗时: %v, 新增匹配: %d, 活跃碰撞码: %d", elapsed, matchCount, len(activeCodes))
}

// cleanInvalidCodes 清理无效的碰撞码(用户不存在的)
func (cm *CollisionMatcher) cleanInvalidCodes() {
	var invalidCount int64

	// 查找所有用户不存在的碰撞码
	result := config.DB.Exec(`
		UPDATE collision_codes 
		SET status = 'invalid', is_matched = true 
		WHERE user_id NOT IN (SELECT id FROM users) 
		AND status = 'active' 
		AND is_matched = false
	`)

	invalidCount = result.RowsAffected
	if invalidCount > 0 {
		log.Printf("🧹 清理了 %d 个无效碰撞码(用户不存在)", invalidCount)
	}
}

// findAllMatches 为指定的碰撞码寻找所有可能的匹配（多对多）- 简化版：仅关键词相同即可匹配
func (cm *CollisionMatcher) findAllMatches(collisionCode *models.CollisionCode) int {
	matchCount := 0

	// 构建简化查询：仅相同关键词、活跃或过期状态、不是自己
	baseQuery := config.DB.Model(&models.CollisionCode{}).
		Where("collision_codes.tag = ? AND collision_codes.user_id != ?",
			collisionCode.Tag, collisionCode.UserID)

	// 查找所有符合条件的碰撞码（多对多）
	var matchedCodes []models.CollisionCode
	baseQuery.Preload("User").Find(&matchedCodes)

	// 为每个匹配创建记录（跳过已存在的匹配）
	for _, matchedCode := range matchedCodes {
		if cm.createMatchIfNotExists(collisionCode, &matchedCode) {
			matchCount++
		}
	}

	return matchCount
}

// createMatchIfNotExists 检查匹配是否已存在，不存在则创建 - 简化版：仅关键词相同即可匹配
func (cm *CollisionMatcher) createMatchIfNotExists(code1, code2 *models.CollisionCode) bool {
	// 检查是否已存在匹配结果（双向检查）
	var existingCount int64
	config.DB.Model(&models.CollisionResult{}).
		Where("(user_id = ? AND matched_user_id = ? AND keyword = ?) OR (user_id = ? AND matched_user_id = ? AND keyword = ?)",
			uint64(code1.UserID), uint64(code2.UserID), code1.Tag,
			uint64(code2.UserID), uint64(code1.UserID), code1.Tag).
		Count(&existingCount)

	if existingCount > 0 {
		return false // 已存在匹配，跳过
	}

	// 简化匹配类型：仅使用keyword即可
	matchType := "keyword" // 简化为仅关键词匹配

	return cm.createMatchRecord(code1, code2, matchType)
}

// findAndCreateMatch 为指定的碰撞码寻找匹配并创建记录（保留兼容，但不再使用）
func (cm *CollisionMatcher) findAndCreateMatch(collisionCode *models.CollisionCode) bool {
	log.Printf("为碰撞码 #%d (UserID:%d, Tag:%s) 寻找匹配...",
		collisionCode.ID, collisionCode.UserID, collisionCode.Tag)

	var matchedCode *models.CollisionCode
	var matchType string

	// 构建基础查询条件
	baseQuery := config.DB.Model(&models.CollisionCode{}).
		Where("collision_codes.tag = ? AND collision_codes.status != ? AND collision_codes.user_id != ? AND collision_codes.is_matched = ?",
			collisionCode.Tag, "invalid", collisionCode.UserID, false).
		Joins("LEFT JOIN users ON collision_codes.user_id = users.id").
		Where("users.location_visible = ?", true) // 只匹配地区可见的用户

	// 性别筛选（如果碰撞码指定了性别要求）
	if collisionCode.Gender > 0 {
		baseQuery = baseQuery.Where("users.gender = ?", collisionCode.Gender)
	}

	// 年龄筛选（对方的年龄必须在我的年龄范围内）
	if collisionCode.AgeMin > 0 && collisionCode.AgeMax > 0 {
		baseQuery = baseQuery.Where("users.age >= ? AND users.age <= ?", collisionCode.AgeMin, collisionCode.AgeMax)
	}

	// 双向匹配逻辑（精准匹配）：
	// 用户A: 搜索区域B + 个人地址A
	// 用户B: 搜索区域A + 个人地址B
	// 只有当 A搜索B区域 且 B搜索A区域 时才匹配成功
	//
	// 实现方式：
	// 1. 找到所有"搜索区域 = 当前用户个人地址"的碰撞码
	// 2. 检查这些碰撞码的用户个人地址 是否等于 当前碰撞码的搜索区域

	// 获取当前用户的个人地址
	currentUser := collisionCode.User
	if currentUser.ID == 0 {
		// 如果没有预加载，手动获取
		if err := config.DB.First(&currentUser, collisionCode.UserID).Error; err != nil {
			log.Printf("获取用户信息失败: %v", err)
			return false
		}
	}

	log.Printf("当前用户 User%d - 个人地址: %s/%s/%s/%s, 搜索区域: %s/%s/%s/%s",
		currentUser.ID, currentUser.Country, currentUser.Province, currentUser.City, currentUser.District,
		collisionCode.Country, collisionCode.Province, collisionCode.City, collisionCode.District)

	// 构建匹配查询：
	// 对方的搜索区域 = 我的个人地址
	// 对方的个人地址 = 我的搜索区域

	// 区县级精准匹配
	if collisionCode.District != "" && currentUser.District != "" {
		// 我搜索东城区，我在西城区
		// 对方搜索西城区，对方在东城区
		query := baseQuery.Where(
			"collision_codes.district = ? AND collision_codes.city = ? AND collision_codes.province = ? AND collision_codes.country = ?",
			currentUser.District, currentUser.City, currentUser.Province, currentUser.Country,
		).Where(
			"users.district = ? AND users.city = ? AND users.province = ? AND users.country = ?",
			collisionCode.District, collisionCode.City, collisionCode.Province, collisionCode.Country,
		)

		if err := query.Preload("User").First(&matchedCode).Error; err == nil {
			matchType = "district"
			log.Printf("✅ 区县精准匹配成功 - 碰撞码#%d (搜%s,在%s) <-> 碰撞码#%d (搜%s,在%s)",
				collisionCode.ID, collisionCode.District, currentUser.District,
				matchedCode.ID, matchedCode.District, matchedCode.User.District)
		}
	}

	// 城市级精准匹配
	if matchedCode == nil && collisionCode.City != "" && currentUser.City != "" && collisionCode.District == "" && currentUser.District == "" {
		query := baseQuery.Where(
			"collision_codes.city = ? AND collision_codes.province = ? AND collision_codes.country = ? AND collision_codes.district = ?",
			currentUser.City, currentUser.Province, currentUser.Country, "",
		).Where(
			"users.city = ? AND users.province = ? AND users.country = ? AND users.district = ?",
			collisionCode.City, collisionCode.Province, collisionCode.Country, "",
		)

		if err := query.Preload("User").First(&matchedCode).Error; err == nil {
			matchType = "city"
			log.Printf("✅ 城市精准匹配成功 - 碰撞码#%d (搜%s,在%s) <-> 碰撞码#%d (搜%s,在%s)",
				collisionCode.ID, collisionCode.City, currentUser.City,
				matchedCode.ID, matchedCode.City, matchedCode.User.City)
		}
	}

	// 省份级精准匹配
	if matchedCode == nil && collisionCode.Province != "" && currentUser.Province != "" && collisionCode.City == "" && currentUser.City == "" {
		query := baseQuery.Where(
			"collision_codes.province = ? AND collision_codes.country = ? AND collision_codes.city = ? AND collision_codes.district = ?",
			currentUser.Province, currentUser.Country, "", "",
		).Where(
			"users.province = ? AND users.country = ? AND users.city = ? AND users.district = ?",
			collisionCode.Province, collisionCode.Country, "", "",
		)

		if err := query.Preload("User").First(&matchedCode).Error; err == nil {
			matchType = "province"
			log.Printf("✅ 省份精准匹配成功 - 碰撞码#%d (搜%s,在%s) <-> 碰撞码#%d (搜%s,在%s)",
				collisionCode.ID, collisionCode.Province, currentUser.Province,
				matchedCode.ID, matchedCode.Province, matchedCode.User.Province)
		}
	}

	// 国家级精准匹配
	if matchedCode == nil && collisionCode.Country != "" && currentUser.Country != "" && collisionCode.Province == "" && currentUser.Province == "" {
		query := baseQuery.Where(
			"collision_codes.country = ? AND collision_codes.province = ? AND collision_codes.city = ? AND collision_codes.district = ?",
			currentUser.Country, "", "", "",
		).Where(
			"users.country = ? AND users.province = ? AND users.city = ? AND users.district = ?",
			collisionCode.Country, "", "", "",
		)

		if err := query.Preload("User").First(&matchedCode).Error; err == nil {
			matchType = "country"
			log.Printf("✅ 国家精准匹配成功 - 碰撞码#%d (搜%s,在%s) <-> 碰撞码#%d (搜%s,在%s)",
				collisionCode.ID, collisionCode.Country, currentUser.Country,
				matchedCode.ID, matchedCode.Country, matchedCode.User.Country)
		}
	}

	// 如果找到匹配，创建碰撞记录
	if matchedCode != nil {
		return cm.createMatchRecord(collisionCode, matchedCode, matchType)
	}

	log.Printf("碰撞码#%d 未找到匹配", collisionCode.ID)
	return false
}

// createMatchRecord 创建匹配记录并更新碰撞码状态
func (cm *CollisionMatcher) createMatchRecord(code1, code2 *models.CollisionCode, matchType string) bool {
	// 验证用户是否存在
	var user1, user2 models.User
	if err := config.DB.First(&user1, code1.UserID).Error; err != nil {
		log.Printf("❌ 用户User%d不存在,跳过匹配: %v", code1.UserID, err)
		return false
	}
	if err := config.DB.First(&user2, code2.UserID).Error; err != nil {
		log.Printf("❌ 用户User%d不存在,跳过匹配: %v", code2.UserID, err)
		return false
	}

	tx := config.DB.Begin()

	// 创建碰撞记录（双向：code1 -> code2 和 code2 -> code1）
	record1 := models.CollisionRecord{
		UserID1:           code1.UserID,
		UserID2:           code2.UserID,
		Tag:               code1.Tag,
		MatchType:         matchType,
		MatchCountry:      code1.Country,
		MatchProvince:     code1.Province,
		MatchCity:         code1.City,
		MatchDistrict:     code1.District,
		Status:            "matched",
		AddFriendDeadline: time.Now().Add(24 * time.Hour),
	}

	if err := tx.Create(&record1).Error; err != nil {
		tx.Rollback()
		log.Printf("创建碰撞记录失败 (User%d->User%d): %v", code1.UserID, code2.UserID, err)
		return false
	}

	// 创建反向记录
	record2 := models.CollisionRecord{
		UserID1:           code2.UserID,
		UserID2:           code1.UserID,
		Tag:               code1.Tag,
		MatchType:         matchType,
		MatchCountry:      code1.Country,
		MatchProvince:     code1.Province,
		MatchCity:         code1.City,
		MatchDistrict:     code1.District,
		Status:            "matched",
		AddFriendDeadline: time.Now().Add(24 * time.Hour),
	}

	if err := tx.Create(&record2).Error; err != nil {
		tx.Rollback()
		log.Printf("创建反向碰撞记录失败 (User%d->User%d): %v", code2.UserID, code1.UserID, err)
		return false
	}

	// 更新两个碰撞码的匹配计数和匹配状态
	if err := tx.Model(code1).Updates(map[string]interface{}{
		"match_count": gorm.Expr("match_count + 1"),
		"is_matched":  true,
	}).Error; err != nil {
		tx.Rollback()
		log.Printf("更新碰撞码#%d状态失败: %v", code1.ID, err)
		return false
	}

	if err := tx.Model(code2).Updates(map[string]interface{}{
		"match_count": gorm.Expr("match_count + 1"),
		"is_matched":  true,
	}).Error; err != nil {
		tx.Rollback()
		log.Printf("更新碰撞码#%d状态失败: %v", code2.ID, err)
		return false
	}

	// ========== V3.0 新增：写入 collision_results 表并发送邮件通知 ==========
	// 获取双方用户的联系方式
	var contact1, contact2 models.UserContact
	config.DB.Where("user_id = ?", code1.UserID).First(&contact1)
	config.DB.Where("user_id = ?", code2.UserID).First(&contact2)

	now := time.Now()

	// 写入 V3 碰撞结果表（用户1看到用户2）
	collisionResult1 := models.CollisionResult{
		UserID:          uint64(code1.UserID),
		MatchedUserID:   uint64(code2.UserID),
		CollisionListID: 0, // 由 CollisionCode 触发，无关联的 CollisionList
		Keyword:         code1.Tag,
		MatchedEmail:    contact2.Email,
		MatchedAt:       now,
	}
	tx.Create(&collisionResult1)

	// 写入 V3 碰撞结果表（用户2看到用户1）
	collisionResult2 := models.CollisionResult{
		UserID:          uint64(code2.UserID),
		MatchedUserID:   uint64(code1.UserID),
		CollisionListID: 0,
		Keyword:         code1.Tag,
		MatchedEmail:    contact1.Email,
		MatchedAt:       now,
	}
	tx.Create(&collisionResult2)

	tx.Commit()
	log.Printf("✅ 匹配成功！碰撞码#%d (User%d) <-> 碰撞码#%d (User%d), 类型: %s",
		code1.ID, code1.UserID, code2.ID, code2.UserID, matchType)

	// 更新碰撞列表的匹配数量
	// 1. 更新code1用户的碰撞列表
	config.DB.Model(&models.CollisionList{}).
		Where("user_id = ? AND keyword = ? AND status = 'active'", uint64(code1.UserID), code1.Tag).
		UpdateColumn("match_count", gorm.Expr("match_count + 1"))

	// 2. 更新code2用户的碰撞列表
	config.DB.Model(&models.CollisionList{}).
		Where("user_id = ? AND keyword = ? AND status = 'active'", uint64(code2.UserID), code1.Tag).
		UpdateColumn("match_count", gorm.Expr("match_count + 1"))

	// 不自动发送邮件，用户手动选择发送

	// 更新热门标签计数(基于碰撞次数)
	go cm.updateHotTagCount(code1.Tag)

	return true
}

// sendEmailNotifications 发送邮件通知给双方（V3.0 新增）
func (cm *CollisionMatcher) sendEmailNotifications(userID1, userID2 uint64, keyword string, contact1, contact2 models.UserContact) {
	emailService := NewSMTPEmailService(config.DB)

	// 发送给用户1（如果已验证邮箱），邮件中包含用户2的邮箱
	if contact1.Email != "" && contact1.EmailVerified {
		partnerEmail := ""
		if contact2.Email != "" && contact2.EmailVerified {
			partnerEmail = contact2.Email
		}
		if err := emailService.SendCollisionNotifyEmailWithPartnerCompat(userID1, contact1.Email, keyword, 1, partnerEmail); err != nil {
			log.Printf("📧 发送邮件给User%d失败: %v", userID1, err)
		} else {
			log.Printf("📧 已发送碰撞通知邮件给 User%d (%s)，包含对方邮箱: %s", userID1, contact1.Email, partnerEmail)
			// 更新邮件发送状态
			config.DB.Model(&models.CollisionResult{}).
				Where("user_id = ? AND matched_user_id = ? AND keyword = ?", userID1, userID2, keyword).
				Updates(map[string]interface{}{
					"email_sent":    true,
					"email_sent_at": time.Now(),
				})
		}
	}

	// 发送给用户2（如果已验证邮箱），邮件中包含用户1的邮箱
	if contact2.Email != "" && contact2.EmailVerified {
		partnerEmail := ""
		if contact1.Email != "" && contact1.EmailVerified {
			partnerEmail = contact1.Email
		}
		if err := emailService.SendCollisionNotifyEmailWithPartnerCompat(userID2, contact2.Email, keyword, 1, partnerEmail); err != nil {
			log.Printf("📧 发送邮件给User%d失败: %v", userID2, err)
		} else {
			log.Printf("📧 已发送碰撞通知邮件给 User%d (%s)，包含对方邮箱: %s", userID2, contact2.Email, partnerEmail)
			// 更新邮件发送状态
			config.DB.Model(&models.CollisionResult{}).
				Where("user_id = ? AND matched_user_id = ? AND keyword = ?", userID2, userID1, keyword).
				Updates(map[string]interface{}{
					"email_sent":    true,
					"email_sent_at": time.Now(),
				})
		}
	}
}

// updateHotTagCount 更新热门标签计数
func (cm *CollisionMatcher) updateHotTagCount(keyword string) {
	var tag models.HotTag
	now := time.Now()

	// 检查标签是否存在
	if err := config.DB.Where("keyword = ?", keyword).First(&tag).Error; err != nil {
		// 不存在则创建
		tag = models.HotTag{
			Keyword:      keyword,
			Count24h:     0,
			CountTotal:   0,
			Status:       "hide",
			SubmitCount:  0,
			LastSearchAt: &now,
		}
		config.DB.Create(&tag)
		return
	} else {
		if tag.Status != "show" {
			return
		}

		// 检查是否超过24小时
		isWithin24h := tag.LastSearchAt != nil && now.Sub(*tag.LastSearchAt) <= 24*time.Hour

		updateData := map[string]interface{}{
			"count_total":    tag.CountTotal + 1,
			"last_search_at": now,
		}

		// 如果在24小时内，增加count_24h
		if isWithin24h {
			updateData["count_24h"] = tag.Count24h + 1
		} else {
			// 超过24小时，重置count_24h为1
			updateData["count_24h"] = 1
		}

		// 更新计数
		config.DB.Model(&tag).Updates(updateData)
	}
}

// StartMatcherService 启动定期匹配服务
func (cm *CollisionMatcher) StartMatcherService(interval time.Duration) {
	log.Printf("启动碰撞匹配服务，间隔: %v", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 立即执行一次
	cm.RunMatcher()

	// 定期执行
	for range ticker.C {
		cm.RunMatcher()
	}
}
