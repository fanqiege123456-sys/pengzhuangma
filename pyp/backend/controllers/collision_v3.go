package controllers

import (
	"fmt"
	"html"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"collision-backend/config"
	"collision-backend/models"
	"collision-backend/services"

	"github.com/gin-gonic/gin"
)

// maskKeyword 关键词脱敏处理：替换50%的字符为*
func maskKeyword(keyword string) string {
	if len(keyword) <= 1 {
		return keyword
	}

	runes := []rune(keyword)
	maskCount := len(runes) / 2
	if maskCount == 0 {
		maskCount = 1
	}

	// 替换中间的50%字符为*
	start := len(runes) / 4
	for i := 0; i < maskCount && start+i < len(runes); i++ {
		runes[start+i] = '*'
	}

	return string(runes)
}

// GetHotTags24h 获取24小时热门标签（只展示前三位）
func GetHotTags24h(c *gin.Context) {
	var tags []models.HotTag

	// 先获取所有状态为show的标签
	config.DB.Where("status = ?", "show").Find(&tags)

	// 使用Go代码进行排序，确保按count_24h降序
	// 注意：数据库中存在count_24h和count24h两个字段，我们需要确保使用正确的字段
	for i := 0; i < len(tags); i++ {
		for j := i + 1; j < len(tags); j++ {
			// 比较两个标签的count_24h，将较大的放在前面
			if tags[i].Count24h < tags[j].Count24h {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}

	// 只取前3个
	if len(tags) > 3 {
		tags = tags[:3]
	}

	// 添加排名
	result := make([]gin.H, len(tags))
	for i, tag := range tags {
		result[i] = gin.H{
			"rank":    i + 1,
			"keyword": tag.Keyword,
			"count":   tag.Count24h,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// GetHotTagsAll 获取总榜热门标签
func GetHotTagsAll(c *gin.Context) {
	var tags []models.HotTag

	// 先获取所有状态为show的标签
	config.DB.Where("status = ?", "show").Find(&tags)

	// 使用Go代码进行排序，确保按count_total降序
	for i := 0; i < len(tags); i++ {
		for j := i + 1; j < len(tags); j++ {
			// 比较两个标签的count_total，将较大的放在前面
			if tags[i].CountTotal < tags[j].CountTotal {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}

	// 只取前3个
	if len(tags) > 3 {
		tags = tags[:3]
	}

	// 添加排名
	result := make([]gin.H, len(tags))
	for i, tag := range tags {
		result[i] = gin.H{
			"rank":    i + 1,
			"keyword": tag.Keyword,
			"count":   tag.CountTotal,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// CreateCollisionList 创建碰撞列表
func CreateCollisionList(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	var req struct {
		Keyword  string `json:"keyword" binding:"required"`
		Duration int    `json:"duration"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if req.Duration == 0 {
		req.Duration = 30
	}

	// 检查是否已存在
	var existingList models.CollisionList
	if err := config.DB.Where("user_id = ? AND keyword = ? AND status = 'active'", userID, req.Keyword).
		First(&existingList).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该关键词已在碰撞列表中"})
		return
	}

	// 检查积分余额
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败"})
		return
	}

	costPoints := req.Duration // 每天1积分
	if user.Coins < costPoints {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "积分不足"})
		return
	}

	// 扣除积分
	config.DB.Model(&user).Update("coins", user.Coins-costPoints)

	// 创建碰撞列表
	collisionList := models.CollisionList{
		UserID:     uint64(userID),
		Keyword:    req.Keyword,
		Duration:   req.Duration,
		CostPoints: costPoints,
		Status:     "active",
		ExpireAt:   time.Now().AddDate(0, 0, req.Duration),
	}
	config.DB.Create(&collisionList)

	// 更新热门标签统计
	updateHotTag(req.Keyword)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "添加成功",
		"data":    collisionList,
	})
}

// GetCollisionLists 获取我的碰撞列表
func GetCollisionLists(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	// 获取当前用户所有碰撞列表，包括已过期但状态未更新的
	var lists []models.CollisionList
	config.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&lists)

	// 处理返回数据，格式化日期并添加过期状态，同时返回完整数据库字段
	result := make([]gin.H, len(lists))
	now := time.Now()

	// 批量更新过期的碰撞列表状态
	var expiredListIDs []uint64

	for i, list := range lists {
		// 检查是否过期
		isExpired := list.ExpireAt.Before(now)
		currentStatus := list.Status

		// 记录需要更新的ID
		if isExpired && list.Status != "expired" {
			currentStatus = "expired"
			expiredListIDs = append(expiredListIDs, list.ID)
		}

		// 计算剩余时间
		timeLeft := list.ExpireAt.Sub(now)

		// 格式化所有时间字段
		formattedExpireAt := list.ExpireAt.Format("2006年01月02日 15:04:05")
		formattedCreatedAt := list.CreatedAt.Format("2006年01月02日 15:04:05")
		formattedUpdatedAt := list.UpdatedAt.Format("2006年01月02日 15:04:05")

		// 返回完整数据库字段，同时保留原有自定义字段
		result[i] = gin.H{
			// 完整数据库字段
			"id":          list.ID,
			"user_id":     list.UserID,
			"keyword":     list.Keyword,
			"duration":    list.Duration,
			"cost_points": list.CostPoints,
			"status":      currentStatus,
			"expire_at":   formattedExpireAt,
			"match_count": list.MatchCount,
			"created_at":  formattedCreatedAt,
			"updated_at":  formattedUpdatedAt,

			// 原有自定义字段（为了向后兼容）
			"is_expired":        isExpired,
			"time_left":         int(timeLeft.Hours()),
			"time_left_seconds": int(timeLeft.Seconds()),
		}
	}

	// 批量更新过期的碰撞列表状态
	if len(expiredListIDs) > 0 {
		config.DB.Model(&models.CollisionList{}).
			Where("id IN ?", expiredListIDs).
			Update("status", "expired")
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// UpdateCollisionList 更新碰撞列表
func UpdateCollisionList(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var list models.CollisionList
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&list).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "碰撞列表不存在"})
		return
	}

	var req struct {
		Status string `json:"status"`
		Extend int    `json:"extend"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 延长有效期
	if req.Extend > 0 {
		var user models.User
		if err := config.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败"})
			return
		}

		if user.Coins < req.Extend {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "积分不足"})
			return
		}

		// 扣除积分
		config.DB.Model(&user).Update("coins", user.Coins-req.Extend)

		// 延长过期时间
		list.ExpireAt = list.ExpireAt.AddDate(0, 0, req.Extend)
		list.Duration += req.Extend
		list.CostPoints += req.Extend
	}

	// 更新状态
	if req.Status != "" {
		list.Status = req.Status
	}

	config.DB.Save(&list)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    list,
	})
}

// DeleteCollisionList 删除碰撞列表
func DeleteCollisionList(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	result := config.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.CollisionList{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "碰撞列表不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// GetCollisionResults 获取碰撞结果
func GetCollisionResults(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	keyword := c.Query("keyword") // 新增：接受关键词参数
	startDate := time.Now().AddDate(0, 0, -days)

	// 先获取匹配结果，然后在Go代码中进行分组
	var allResults []models.CollisionResult
	query := config.DB.Where("user_id = ? AND matched_at >= ?", userID, startDate)

	// 如果提供了关键词，按关键词过滤
	if keyword != "" {
		query = query.Where("keyword = ?", keyword)
	}

	query.Order("matched_at DESC").Find(&allResults)

	// 在Go代码中按日期和关键词分组
	groupMap := make(map[string][]models.CollisionResult)
	for _, result := range allResults {
		// 使用完整的日期时间格式作为分组键，但只精确到天
		groupKey := result.MatchedAt.Format("2006-01-02") + "_" + result.Keyword
		groupMap[groupKey] = append(groupMap[groupKey], result)
	}

	// 收集所有分组键并排序
	var groupKeys []string
	for key := range groupMap {
		groupKeys = append(groupKeys, key)
	}

	// 按照时间倒序排列分组键
	sort.Strings(groupKeys)
	// 反转排序，使最新的日期显示在最前面
	for i, j := 0, len(groupKeys)-1; i < j; i, j = i+1, j-1 {
		groupKeys[i], groupKeys[j] = groupKeys[j], groupKeys[i]
	}

	// 构建结果
	result := make([]gin.H, 0, len(groupMap))
	for _, groupKey := range groupKeys {
		matches := groupMap[groupKey]

		// 解析分组键获取日期和关键词
		// 格式: 2026-01-02_keyword
		var dateStr, keyword string
		fmt.Sscanf(groupKey, "%[^_]_%s", &dateStr, &keyword)

		// 计算分组统计信息
		total := len(matches)
		knownCount := 0
		for _, m := range matches {
			if m.Remark != "" {
				knownCount++
			}
		}

		// 限制匹配结果数量为50个
		if len(matches) > 50 {
			matches = matches[:50]
		}

		// 构建匹配列表，返回完整字段
		matchList := make([]gin.H, len(matches))
		for j, m := range matches {
			matchList[j] = gin.H{
				// 完整数据库字段
				"id":                m.ID,
				"user_id":           m.UserID,
				"matched_user_id":   m.MatchedUserID,
				"collision_list_id": m.CollisionListID,
				"keyword":           m.Keyword,
				"matched_email":     m.MatchedEmail,
				"remark":            m.Remark,
				"is_known":          m.IsKnown,
				"email_sent":        m.EmailSent,
				"email_sent_at":     m.EmailSentAt,
				"matched_at":        m.MatchedAt.Format("2006年01月02日 15:04:05"),
				"created_at":        m.CreatedAt.Format("2006年01月02日 15:04:05"),
				"updated_at":        m.UpdatedAt.Format("2006年01月02日 15:04:05"),
			}
		}

		// 使用分组中第一条记录的完整时间作为分组日期
		// 因为分组是按时间倒序的，所以第一条记录是最新的
		groupDate := matches[0].MatchedAt
		formattedDate := groupDate.Format("2006年01月02日 15:04:05")

		result = append(result, gin.H{
			"id":         groupKey,
			"date":       formattedDate,
			"keyword":    maskKeyword(keyword),
			"keyword_raw": keyword,
			"total":      total,
			"knownCount": knownCount,
			"matches":    matchList,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// GetCollisionResultDetail 获取碰撞结果详情(分页)
func GetCollisionResultDetail(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	id := c.Param("id") // 格式: date_keyword
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// 解析ID，提取关键词
	var dateStr, keyword string
	fmt.Sscanf(id, "%[^_]_%s", &dateStr, &keyword)

	var matches []models.CollisionResult
	query := config.DB.Where("user_id = ?", userID)

	// 如果解析到关键词，按关键词过滤
	if keyword != "" {
		query = query.Where("keyword = ?", keyword)
	}

	query.Order("CASE WHEN remark = '' THEN 0 ELSE 1 END, matched_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&matches)

	// 收集所有被匹配用户的ID
	matchedUserIDs := make([]uint64, len(matches))
	for i, m := range matches {
		matchedUserIDs[i] = m.MatchedUserID
	}

	// 批量查询被匹配用户的邮箱显示设置
	var contacts []models.UserContact
	config.DB.Where("user_id IN ?", matchedUserIDs).Find(&contacts)

	// 创建用户ID到邮箱可见性的映射
	emailVisibleMap := make(map[uint64]bool)
	for _, contact := range contacts {
		emailVisibleMap[contact.UserID] = contact.EmailVisible
	}

	matchList := make([]gin.H, len(matches))
	for j, m := range matches {
		// 检查被匹配用户是否允许显示邮箱
		displayEmail := m.MatchedEmail
		emailVisible, exists := emailVisibleMap[m.MatchedUserID]
		// 如果没有设置或者设置为不可见，则隐藏邮箱
		if exists && !emailVisible {
			displayEmail = "对方已隐藏邮箱"
		}

		matchList[j] = gin.H{
			// 完整数据库字段
			"id":                m.ID,
			"user_id":           m.UserID,
			"matched_user_id":   m.MatchedUserID,
			"collision_list_id": m.CollisionListID,
			"keyword":           m.Keyword,
			"matched_email":     displayEmail, // 使用处理后的邮箱显示
			"remark":            m.Remark,
			"is_known":          m.IsKnown,
			"email_sent":        m.EmailSent,
			"email_sent_at":     m.EmailSentAt,
			"matched_at":        m.MatchedAt.Format("2006年01月02日 15:04:05"),
			"created_at":        m.CreatedAt.Format("2006年01月02日 15:04:05"),
			"updated_at":        m.UpdatedAt.Format("2006年01月02日 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    matchList,
	})
}

// MarkCollisionResultKnown 标记碰撞结果为已知
func MarkCollisionResultKnown(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	result := config.DB.Model(&models.CollisionResult{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_known", true)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// UpdateMatchRemark 更新匹配备注
func UpdateMatchRemark(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req struct {
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 限制备注长度
	if len([]rune(req.Remark)) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "备注最多10个字"})
		return
	}

	// 更新备注
	result := config.DB.Model(&models.CollisionResult{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("remark", req.Remark)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	// 读取更新后的记录，确保返回最新数据
	var updatedResult models.CollisionResult
	config.DB.First(&updatedResult, id)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "保存成功",
		"data": gin.H{
			"id":              updatedResult.ID,
			"user_id":         updatedResult.UserID,
			"matched_user_id": updatedResult.MatchedUserID,
			"keyword":         updatedResult.Keyword,
			"matched_email":   updatedResult.MatchedEmail,
			"remark":          updatedResult.Remark,
			"matched_at":      updatedResult.MatchedAt.Format("2006年01月02日 15:04:05"),
		},
	})
}

// GetCommonKeywords 获取两个用户共同碰撞的关键词
func GetCommonKeywords(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	// 获取对方用户ID
	var req struct {
		MatchedUserID uint64 `json:"matched_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 获取当前用户与对方用户碰撞的所有关键词
	var myKeywords []string
	config.DB.Model(&models.CollisionResult{}).
		Where("user_id = ? AND matched_user_id = ?", userID, req.MatchedUserID).
		Pluck("DISTINCT keyword", &myKeywords)

	// 获取对方用户与当前用户碰撞的所有关键词
	var theirKeywords []string
	config.DB.Model(&models.CollisionResult{}).
		Where("user_id = ? AND matched_user_id = ?", req.MatchedUserID, userID).
		Pluck("DISTINCT keyword", &theirKeywords)

	// 找出共同关键词
	keywordMap := make(map[string]bool)
	for _, keyword := range myKeywords {
		keywordMap[keyword] = true
	}

	var commonKeywords []string
	for _, keyword := range theirKeywords {
		if keywordMap[keyword] {
			commonKeywords = append(commonKeywords, keyword)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":            200,
		"message":         "success",
		"common_keywords": commonKeywords,
		"total":           len(commonKeywords),
	})
}

// 更新热门标签统计
func updateHotTag(keyword string) {
	var tag models.HotTag
	now := time.Now()

	if err := config.DB.Where("keyword = ?", keyword).First(&tag).Error; err != nil {
		// 不存在则创建，默认状态为hide（仅审核后首页展示）
		tag = models.HotTag{
			Keyword:      keyword,
			Count24h:     0,
			CountTotal:   0,
			Status:       "hide",
			SubmitCount:  0,
			LastSearchAt: &now,
		}
		config.DB.Create(&tag)
	} else {
		if tag.Status != "show" {
			return
		}

		// 检查是否超过24小时
		isWithin24h := tag.LastSearchAt != nil && now.Sub(*tag.LastSearchAt) <= 24*time.Hour

		updateData := map[string]interface{}{
			"count_total":    tag.CountTotal + 1,
			"submit_count":   tag.SubmitCount + 1,
			"last_search_at": now,
		}

		// 如果在24小时内，增加count_24h
		if isWithin24h {
			updateData["count_24h"] = tag.Count24h + 1
		} else {
			// 超过24小时，重置count_24h为1
			updateData["count_24h"] = 1
		}

		// 更新计数，包括SubmitCount和count_24h
		config.DB.Model(&tag).Updates(updateData)
	}
}

// SendEmailToMatch 发送邮件给匹配用户，扣除1积分
func SendEmailToMatch(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "not logged in"})
		return
	}

	var req struct {
		ResultID uint64 `json:"result_id" binding:"required"`
		Content  string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid params: " + err.Error()})
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "content required"})
		return
	}
	if len([]rune(req.Content)) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "content too long"})
		return
	}

	var collisionResult models.CollisionResult
	if err := config.DB.Where("id = ? AND user_id = ?", req.ResultID, userID).First(&collisionResult).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "match not found"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to load user"})
		return
	}
	if user.Coins < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "insufficient coins"})
		return
	}

	var matchedContact models.UserContact
	if err := config.DB.Where("user_id = ?", collisionResult.MatchedUserID).First(&matchedContact).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "matched user has no email"})
		return
	}
	if matchedContact.Email == "" || !matchedContact.EmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "matched user email not verified"})
		return
	}

	subject := "小程序匹配成功，用户给你发信息啦"
	htmlBody := fmt.Sprintf("<p>小程序匹配成功，用户给你发信息啦</p>\n<p>%s</p>", html.EscapeString(req.Content))

	emailService := services.NewSMTPEmailService(config.DB)
	if err := emailService.SendEmail(uint64(userID), matchedContact.Email, subject, htmlBody, "collision"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "send email failed: " + err.Error()})
		return
	}

	if err := config.DB.Model(&user).Update("coins", user.Coins-1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to update coins"})
		return
	}

	now := time.Now()
	_ = config.DB.Model(&collisionResult).Updates(map[string]interface{}{
		"email_sent":    true,
		"email_sent_at": now,
	}).Error

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "ok",
		"data": gin.H{
			"remaining_coins": user.Coins - 1,
		},
	})
}

// ClickHotTag 处理标签点击事件，增加标签点击次数
func ClickHotTag(c *gin.Context) {
	var req struct {
		Keyword string `json:"keyword" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 更新标签点击次数
	updateHotTag(req.Keyword)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}
