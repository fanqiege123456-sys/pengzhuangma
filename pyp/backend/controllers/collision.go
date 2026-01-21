package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"collision-backend/config"
	"collision-backend/models"
	"collision-backend/services"
	"collision-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CollisionController struct{}

func defaultAuditStatus() string {
	if config.Config.EnableCollisionAudit {
		return "pending"
	}
	return "approved"
}

func markForbiddenKeywords(codes []models.CollisionCode) {
	if len(codes) == 0 {
		return
	}

	var forbiddenKeywords []models.ForbiddenKeyword
	if err := config.DB.Find(&forbiddenKeywords).Error; err != nil || len(forbiddenKeywords) == 0 {
		return
	}

	for i := range codes {
		for _, keyword := range forbiddenKeywords {
			if keyword.Keyword != "" && strings.Contains(codes[i].Tag, keyword.Keyword) {
				codes[i].IsForbidden = true
				break
			}
		}
	}
}

// è·åç¢°æç åè¡?
func (cc *CollisionController) GetCodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	statusFilter := c.Query("status")
	auditFilter := c.Query("audit_status")
	keyword := c.Query("keyword")

	var codes []models.CollisionCode
	var total int64

	offset := (page - 1) * pageSize

	query := config.DB.Model(&models.CollisionCode{})
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}
	if auditFilter != "" {
		query = query.Where("audit_status = ?", auditFilter)
	}
	if keyword != "" {
		query = query.Where("tag LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)
	query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&codes)
	markForbiddenKeywords(codes)

	var pendingCount int64
	config.DB.Model(&models.CollisionCode{}).Where("audit_status = ?", "pending").Count(&pendingCount)

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"list": codes,
		"pagination": utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
		"pending_count": pendingCount,
	}))
}

// åå»ºç¢°æç ?
func (cc *CollisionController) CreateCode(c *gin.Context) {
	var req struct {
		UserID    uint   `json:"user_id" binding:"required"`
		Tag       string `json:"tag" binding:"required"` // å´è¶£æ ç­¾
		EndDate   string `json:"end_date" binding:"required"`
		CostCoins int    `json:"cost_coins"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request data"))
		return
	}

	// è·åç¨æ·ä¿¡æ¯ï¼åæ¬å°åï¼?
	var user models.User
	if err := config.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "User not found"))
		return
	}

	// TODO: è§£ææ¥æå­ç¬¦ä¸?

	code := models.CollisionCode{
		UserID:   req.UserID,
		Tag:      req.Tag,
		Country:  user.Country,
		Province: user.Province,
		City:     user.City,
		District: user.District,
		Gender:   user.Gender,
		Status:   "active",
		// 直接设置为 pending 状态，确保首页显示需要审核
		// 但状态为 active，确保立即参与匹配
		AuditStatus: "pending",
		CostCoins:   req.CostCoins,
	}

	if err := config.DB.Create(&code).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create collision code"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(code))
}

// æ´æ°ç¢°æç ç¶æ?
func (cc *CollisionController) UpdateCodeStatus(c *gin.Context) {
	id := c.Param("id")
	var code models.CollisionCode

	if err := config.DB.First(&code, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Collision code not found"))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request data"))
		return
	}

	code.Status = req.Status

	if err := config.DB.Save(&code).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to update status"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(code))
}

// å é¤ç¢°æç ?
func (cc *CollisionController) DeleteCode(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Delete(&models.CollisionCode{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to delete collision code"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{"message": "Collision code deleted successfully"}))
}

// å°ç¨åºç¨æ·æäº¤ç¢°æç 
func (cc *CollisionController) SubmitCode(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var req struct {
		Tag       string `json:"tag" binding:"required"` // å´è¶£æ ç­¾
		Country   string `json:"country"`                // ææå¹éçå½å®?
		Province  string `json:"province"`               // ææå¹éççä»?
		City      string `json:"city"`                   // ææå¹éçåå¸?
		District  string `json:"district"`               // ææå¹éçåºå?
		Gender    int    `json:"gender"`                 // ææå¹éçæ§å« 0:ä¸é 1:ç?2:å¥?
		AgeMin    int    `json:"age_min"`                // æå°å¹´é¾ï¼é»è®¤20
		AgeMax    int    `json:"age_max"`                // æå¤§å¹´é¾ï¼é»è®¤30
		CostCoins int    `json:"cost_coins"`             // æ¶èéå¸æ°é?
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ç¢°æè¯·æ±åæ°éè¯¯: %v", err)
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request: "+err.Error()))
		return
	}

	// è®¾ç½®å¹´é¾èå´é»è®¤å?
	if req.AgeMin <= 0 {
		req.AgeMin = 20
	}
	if req.AgeMax <= 0 {
		req.AgeMax = 30
	}
	req.CostCoins = 10

	log.Printf("æ¶å°ç¢°æè¯·æ± - UserID: %v, Tag: %s, Location: %s/%s/%s/%s, Gender: %d, Age: %d-%d, CostCoins: %d",
		userID, req.Tag, req.Country, req.Province, req.City, req.District, req.Gender, req.AgeMin, req.AgeMax, req.CostCoins)

	// è·åå½åç¨æ·ä¿¡æ¯
	var currentUser models.User
	if err := config.DB.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "User not found"))
		return
	}

	// æ£æ¥ç¨æ·éå¸æ¯å¦è¶³å¤?
	if currentUser.Coins < req.CostCoins {
		c.JSON(http.StatusBadRequest, utils.Error(400, fmt.Sprintf("Insufficient coins: %d, need %d", currentUser.Coins, req.CostCoins)))
		return
	}

	// å¼å§äºå?
	tx := config.DB.Begin()

	// æ£é¤éå¸å¹¶è®°å½æ¶è´?
	newCoins := currentUser.Coins - req.CostCoins
	if err := tx.Model(&currentUser).Update("coins", newCoins).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to deduct coins"))
		return
	}

	// è®°å½æ¶è´¹è®°å½
	consumeRecord := models.ConsumeRecord{
		UserID: userID.(uint),
		Coins:  req.CostCoins,
		Type:   "collision",
		Reason: "åå¸ç¢°æç ? " + req.Tag,
	}
	if err := tx.Create(&consumeRecord).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create consume record"))
		return
	}

	// åå»ºç¢°æç ï¼24å°æ¶åè¿æï¼
	collisionCode := models.CollisionCode{
		UserID:   userID.(uint),
		Tag:      req.Tag,      // å´è¶£æ ç­¾
		Country:  req.Country,  // ä½¿ç¨è¯·æ±ä¸­çå°å
		Province: req.Province, // ä½¿ç¨è¯·æ±ä¸­çå°å
		City:     req.City,     // ä½¿ç¨è¯·æ±ä¸­çå°å
		District: req.District, // ä½¿ç¨è¯·æ±ä¸­çå°å
		Gender:   req.Gender,   // ææå¹éçæ§å«
		AgeMin:   req.AgeMin,   // æå°å¹´é¾?
		AgeMax:   req.AgeMax,   // æå¤§å¹´é¾?
		Status:   "active",
		// 直接设置为 pending 状态，确保首页显示需要审核
		// 但状态为 active，确保立即参与匹配
		AuditStatus: "pending",
		ExpiresAt:   time.Now().Add(24 * time.Hour), // 24å°æ¶åè¿æ?
		CostCoins:   req.CostCoins,
	}

	log.Printf("åå¤åå»ºç¢°æç ?- UserID: %d, Tag: %s, Gender: %d, Age: %d-%d, Location: %s/%s/%s/%s",
		collisionCode.UserID, collisionCode.Tag, collisionCode.Gender, collisionCode.AgeMin, collisionCode.AgeMax,
		collisionCode.Country, collisionCode.Province, collisionCode.City, collisionCode.District)

	if err := tx.Create(&collisionCode).Error; err != nil {
		tx.Rollback()
		log.Printf("åå»ºç¢°æç å¤±è´?- UserID: %v, Tag: %s, Error: %v", userID, req.Tag, err)
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create collision code: "+err.Error()))
		return
	}

	log.Printf("ç¢°æç åå»ºæå?- ID: %d, Tag: %s", collisionCode.ID, collisionCode.Tag)

	// æäº¤äºå¡
	tx.Commit()

	matcher := services.NewCollisionMatcher()
	matcher.MatchForCode(&collisionCode)

	// æ´æ°ç­é¨å³é®è¯ç»è®?å¦æææ ç­?
	if req.Tag != "" {
		var keyword models.HotTag
		err := config.DB.Where("keyword = ?", req.Tag).First(&keyword).Error
		if err == nil {
			// å³é®è¯å·²å­å¨,å¢å è®¡æ°
			config.DB.Model(&keyword).UpdateColumn("submit_count", gorm.Expr("submit_count + ?", 1))
		} else {
			// å³é®è¯ä¸å­å¨,åå»ºæ°ç(é»è®¤ä¸ºshowç¶æ?
			newKeyword := models.HotTag{
				Keyword:     req.Tag,
				Status:      "hide",
				SubmitCount: 1,
			}
			config.DB.Create(&newKeyword)
		}
	}

	// æ³¨æï¼ç¢°æå¹éä¸åç«å³è¿è¡ï¼èæ¯ç±åå°èæ¬å®æå¤ç?
	// è¿æ ·å¯ä»¥é¿åå¹¶åé®é¢ï¼å¹¶ä¸å¯ä»¥æ¹éå¤çæé«æç?

	// æ£æ¥æ¯å¦æåå²ç¢°æè®°å½å¯ä»¥æµ·åºæ?
	var historicalUsersCount int64
	config.DB.Model(&models.CollisionCode{}).
		Where("tag = ? AND user_id != ? AND expires_at > ?", req.Tag, userID, time.Now()).
		Joins("LEFT JOIN users ON collision_codes.user_id = users.id").
		Where("users.allow_haidilao = ?", true).
		Count(&historicalUsersCount)

	canHaidilao := historicalUsersCount > 0

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"message":        "ç¢°æç åå¸æåï¼ç­å¾å¹é...",
		"matched":        false,
		"code_id":        collisionCode.ID,
		"expires_at":     collisionCode.ExpiresAt,
		"can_haidilao":   canHaidilao,
		"haidilao_cost":  100, // æµ·åºææ¶è?00ç§¯å
		"haidilao_count": historicalUsersCount,
	}))
}

// è·åç¨æ·çå¹éè®°å½?
func (cc *CollisionController) GetMatches(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var records []models.CollisionRecord
	config.DB.Where("user_id1 = ? OR user_id2 = ?", userID, userID).
		Preload("User1").
		Preload("User2").
		Order("created_at DESC").
		Find(&records)

	// ä¸ºæ¯ä¸ªå¹éè®°å½æ·»å ç¶æååè®¡æ¶ä¿¡æ?
	var enrichedRecords []gin.H
	now := time.Now()

	for _, record := range records {
		// ç¡®å®å¯¹æ¹ç¨æ·
		var partner models.User
		if record.UserID1 == userID.(uint) {
			partner = record.User2
		} else {
			partner = record.User1
		}

		// è®¡ç®å©ä½æ¶é´
		timeLeft := record.AddFriendDeadline.Sub(now)
		var timeStatus string
		var canForceAdd bool

		if record.Status == "friend_added" {
			timeStatus = "already_friends"
			canForceAdd = false
		} else if record.Status == "missed" {
			timeStatus = "missed"
			canForceAdd = false
		} else if timeLeft > 0 {
			timeStatus = "active"
			canForceAdd = false
		} else {
			timeStatus = "expired"
			canForceAdd = partner.AllowPassiveAdd // åªæå¯¹æ¹åè®¸è¢«å¨æ·»å æè½å¼ºå¶å å¥½å?
		}

		enrichedRecord := gin.H{
			"id":                  record.ID,
			"tag":                 record.Tag,
			"match_type":          record.MatchType,
			"status":              record.Status,
			"time_status":         timeStatus,
			"created_at":          record.CreatedAt,
			"add_friend_deadline": record.AddFriendDeadline,
			"time_left_seconds":   int(timeLeft.Seconds()),
			"can_force_add":       canForceAdd,
			"partner": gin.H{
				"id":                partner.ID,
				"nickname":          partner.Nickname,
				"avatar":            partner.Avatar,
				"gender":            partner.Gender,
				"allow_passive_add": partner.AllowPassiveAdd,
			},
			"match_location": gin.H{
				"country":  record.MatchCountry,
				"province": record.MatchProvince,
				"city":     record.MatchCity,
				"district": record.MatchDistrict,
			},
		}

		enrichedRecords = append(enrichedRecords, enrichedRecord)
	}

	c.JSON(http.StatusOK, utils.Success(enrichedRecords))
}

// è·ååæ¡å¹éè¯¦æï¼ä»å¹éåæ¹å¯è§ï¼?
func (cc *CollisionController) GetMatchDetail(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	id := c.Param("id")
	var record models.CollisionRecord
	if err := config.DB.Preload("User1").Preload("User2").First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Match record not found"))
		return
	}

	// ä»åè®¸å¹éåæ¹æ¥ç?
	if record.UserID1 != userID.(uint) && record.UserID2 != userID.(uint) {
		c.JSON(http.StatusForbidden, utils.Error(403, "Not authorized for this match"))
		return
	}

	// ç¡®å®å¯¹æ¹
	var partner models.User
	if record.UserID1 == userID.(uint) {
		partner = record.User2
	} else {
		partner = record.User1
	}

	// è®¡ç®æ¶é´ä¸ç¶æ?
	now := time.Now()
	timeLeft := record.AddFriendDeadline.Sub(now)
	var timeStatus string
	var canForceAdd bool

	if record.Status == "friend_added" {
		timeStatus = "already_friends"
		canForceAdd = false
	} else if record.Status == "missed" {
		timeStatus = "missed"
		canForceAdd = false
	} else if timeLeft > 0 {
		timeStatus = "active"
		canForceAdd = false
	} else {
		timeStatus = "expired"
		canForceAdd = partner.AllowPassiveAdd
	}

	enriched := gin.H{
		"id":                  record.ID,
		"tag":                 record.Tag,
		"match_type":          record.MatchType,
		"status":              record.Status,
		"time_status":         timeStatus,
		"created_at":          record.CreatedAt,
		"add_friend_deadline": record.AddFriendDeadline,
		"time_left_seconds":   int(timeLeft.Seconds()),
		"can_force_add":       canForceAdd,
		"partner": gin.H{
			"id":                partner.ID,
			"nickname":          partner.Nickname,
			"avatar":            partner.Avatar,
			"gender":            partner.Gender,
			"age":               partner.Age,
			"wechat_no":         partner.WechatNo,
			"allow_passive_add": partner.AllowPassiveAdd,
		},
		"match_location": gin.H{
			"country":  record.MatchCountry,
			"province": record.MatchProvince,
			"city":     record.MatchCity,
			"district": record.MatchDistrict,
		},
	}

	c.JSON(http.StatusOK, utils.Success(enriched))
}

// å¼ºå¶æ·»å å¥½åï¼?4å°æ¶è¿æåä½¿ç¨ï¼
func (cc *CollisionController) ForceAddFriend(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var req struct {
		MatchID   uint `json:"match_id" binding:"required"`   // å¹éè®°å½ID
		CostCoins int  `json:"cost_coins" binding:"required"` // å¼ºå¶å å¥½åæ¶èéå¸?
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request"))
		return
	}

	// è·åå½åç¨æ·ä¿¡æ¯
	var currentUser models.User
	if err := config.DB.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "User not found"))
		return
	}

	// æ£æ¥éå¸æ¯å¦è¶³å¤?
	if currentUser.Coins < req.CostCoins {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Insufficient coins"))
		return
	}

	// è·åå¹éè®°å½
	var record models.CollisionRecord
	if err := config.DB.Preload("User1").Preload("User2").First(&record, req.MatchID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Match record not found"))
		return
	}

	// æ£æ¥ç¨æ·æ¯å¦æ¯å¹éçä¸æ?
	if record.UserID1 != userID.(uint) && record.UserID2 != userID.(uint) {
		c.JSON(http.StatusForbidden, utils.Error(403, "Not authorized for this match"))
		return
	}

	// æ£æ¥æ¯å¦å·²è¶è¿24å°æ¶æªæ­¢æ?
	if time.Now().Before(record.AddFriendDeadline) {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Still within 24-hour deadline, cannot force add"))
		return
	}

	// æ£æ¥è®°å½ç¶æ?
	if record.Status != "matched" {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Match status does not allow force add"))
		return
	}

	// ç¡®å®å¯¹æ¹ç¨æ·
	var targetUser models.User
	if record.UserID1 == userID.(uint) {
		targetUser = record.User2
	} else {
		targetUser = record.User1
	}

	// æ£æ¥å¯¹æ¹æ¯å¦åè®¸è¢«å¨æ·»å å¥½å?
	if !targetUser.AllowPassiveAdd {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Target user does not allow passive friend addition"))
		return
	}

	// å¼å§äºå?
	tx := config.DB.Begin()

	// æ£é¤éå¸
	if err := tx.Model(&currentUser).Update("coins", currentUser.Coins-req.CostCoins).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to deduct coins"))
		return
	}

	// è®°å½æ¶è´¹
	consumeRecord := models.ConsumeRecord{
		UserID: userID.(uint),
		Coins:  req.CostCoins,
		Type:   "force_add",
		Reason: "å¼ºå¶æ·»å å¥½å: " + targetUser.Nickname,
	}
	if err := tx.Create(&consumeRecord).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create consume record"))
		return
	}

	// æ£æ¥æ¯å¦å·²ç»æ¯å¥½å
	var existingFriend models.Friend
	err := tx.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, targetUser.ID, targetUser.ID, userID).First(&existingFriend).Error

	if err == nil {
		// å·²ç»æ¯å¥½åï¼ç´æ¥è¿åæå
		tx.Commit()
		c.JSON(http.StatusOK, utils.Success(gin.H{
			"message": "Already friends",
			"friend":  targetUser,
		}))
		return
	}

	// åå»ºå¥½åå³ç³»ï¼ååï¼
	friend1 := models.Friend{
		UserID:   userID.(uint),
		FriendID: targetUser.ID,
		Status:   "accepted", // å¼ºå¶æ·»å ç´æ¥ä¸ºacceptedç¶æ?
	}

	friend2 := models.Friend{
		UserID:   targetUser.ID,
		FriendID: userID.(uint),
		Status:   "accepted",
	}

	if err := tx.Create(&friend1).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create friend relationship"))
		return
	}

	if err := tx.Create(&friend2).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create friend relationship"))
		return
	}

	// æ´æ°å¹éè®°å½ç¶æ?
	if err := tx.Model(&record).Update("status", "friend_added").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to update match status"))
		return
	}

	// æäº¤äºå¡
	tx.Commit()

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"message":     "Successfully forced friend addition",
		"friend":      targetUser,
		"coins_spent": req.CostCoins,
	}))
}

// ä¸»å¨æ·»å å¥½åï¼å¹éæåä¸å?24 å°æ¶åï¼åæ¹å¯ä¸»å¨æ·»å ï¼
func (cc *CollisionController) AddFriend(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var req struct {
		MatchID uint `json:"match_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request"))
		return
	}

	var currentUser models.User
	if err := config.DB.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "User not found"))
		return
	}

	var record models.CollisionRecord
	if err := config.DB.Preload("User1").Preload("User2").First(&record, req.MatchID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Match record not found"))
		return
	}

	// éªè¯æé
	if record.UserID1 != userID.(uint) && record.UserID2 != userID.(uint) {
		c.JSON(http.StatusForbidden, utils.Error(403, "Not authorized for this match"))
		return
	}

	// éªè¯å¹éç¶æ?
	if record.Status != "matched" {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Match status does not allow add friend"))
		return
	}

	// éªè¯æ¯å¦å¨æéå
	if time.Now().After(record.AddFriendDeadline) {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Matching expired, use force add instead"))
		return
	}

	// ç¡®å®å¯¹æ¹
	var targetUser models.User
	if record.UserID1 == userID.(uint) {
		targetUser = record.User2
	} else {
		targetUser = record.User1
	}

	// äºå¡åå»ºå¥½åå³ç³»
	tx := config.DB.Begin()

	var existingFriend models.Friend
	err := tx.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, targetUser.ID, targetUser.ID, userID).First(&existingFriend).Error
	if err == nil {
		// å·²ç»æ¯å¥½å?
		tx.Commit()
		c.JSON(http.StatusOK, utils.Success(gin.H{"message": "Already friends", "friend": targetUser}))
		return
	} else if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to check friend status"))
		return
	}

	friend1 := models.Friend{UserID: userID.(uint), FriendID: targetUser.ID, Status: "accepted"}
	friend2 := models.Friend{UserID: targetUser.ID, FriendID: userID.(uint), Status: "accepted"}

	if err := tx.Create(&friend1).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create friend relationship"))
		return
	}
	if err := tx.Create(&friend2).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create friend relationship"))
		return
	}

	if err := tx.Model(&record).Update("status", "friend_added").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to update match status"))
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, utils.Success(gin.H{"message": "Friend added", "friend": targetUser}))
}

// è·åç­é¨ç¢°æç ?
func (cc *CollisionController) GetHotCodes(c *gin.Context) {
	// ä»ç­é¨å³é®è¯è¡¨è·åæ ç­¾ï¼åªæ¾ç¤ºç¶æä¸ºshowçï¼
	var hotKeywords []models.HotTag

	err := config.DB.Where("status = ?", "show").
		Order("submit_count DESC, created_at DESC").
		Limit(20).
		Find(&hotKeywords).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to get hot tags"))
		return
	}

	// è½¬æ¢ä¸ºåç«¯éè¦çæ ¼å¼
	var hotTags []struct {
		Tag        string `json:"tag"`
		MatchCount int    `json:"match_count"`
	}

	for _, keyword := range hotKeywords {
		hotTags = append(hotTags, struct {
			Tag        string `json:"tag"`
			MatchCount int    `json:"match_count"`
		}{
			Tag:        keyword.Keyword,
			MatchCount: keyword.SubmitCount,
		})
	}

	c.JSON(http.StatusOK, utils.Success(hotTags))
}

// æµ·åºæ?- ä»åå²ç¢°æè®°å½ä¸­éæºæä¸ä¸ªç¨æ?
func (cc *CollisionController) Haidilao(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var req struct {
		Tag       string `json:"tag" binding:"required"` // ç¢°æç æ ç­?
		CostCoins int    `json:"cost_coins"`             // æ¶èéå¸ï¼é»è®¤100ï¼?
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request"))
		return
	}

	// é»è®¤æ¶è?00ç§¯å
	if req.CostCoins == 0 {
		req.CostCoins = 100
	}

	// è·åå½åç¨æ·ä¿¡æ¯
	var currentUser models.User
	if err := config.DB.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "User not found"))
		return
	}

	// æ£æ¥éå¸æ¯å¦è¶³å¤?
	if currentUser.Coins < req.CostCoins {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Insufficient coins"))
		return
	}

	// æ¥æ¾åå²ä¸ä½¿ç¨è¿è¯¥æ ç­¾ä¸åè®¸è¢«æµ·åºæçç¨æ?
	var historicalCodes []models.CollisionCode
	err := config.DB.Where("tag = ? AND user_id != ?", req.Tag, userID).
		Joins("LEFT JOIN users ON collision_codes.user_id = users.id").
		Where("users.allow_haidilao = ?", true).
		Preload("User").
		Find(&historicalCodes).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to search historical records"))
		return
	}

	if len(historicalCodes) == 0 {
		c.JSON(http.StatusNotFound, utils.Error(404, "No users available for Haidilao"))
		return
	}

	// å»éç¨æ·IDï¼ä¸ä¸ªç¨æ·å¯è½æå¤ä¸ªç¢°æç ï¼
	userMap := make(map[uint]models.User)
	for _, code := range historicalCodes {
		if code.User.ID != 0 && code.User.AllowHaidilao {
			userMap[code.User.ID] = code.User
		}
	}

	if len(userMap) == 0 {
		c.JSON(http.StatusNotFound, utils.Error(404, "No users available for Haidilao"))
		return
	}

	// å°ç¨æ·è½¬æ¢ä¸ºåçä»¥ä¾¿éæºéæ©
	var users []models.User
	for _, user := range userMap {
		users = append(users, user)
	}

	// éæºéæ©ä¸ä¸ªç¨æ?
	randomIndex := time.Now().UnixNano() % int64(len(users))
	selectedUser := users[randomIndex]

	// å¼å§äºå?
	tx := config.DB.Begin()

	// æ£é¤éå¸
	if err := tx.Model(&currentUser).Update("coins", currentUser.Coins-req.CostCoins).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to deduct coins"))
		return
	}

	// è®°å½æ¶è´¹
	consumeRecord := models.ConsumeRecord{
		UserID: userID.(uint),
		Coins:  req.CostCoins,
		Type:   "haidilao",
		Reason: "æµ·åºæç¨æ? " + selectedUser.Nickname + " (æ ç­¾: " + req.Tag + ")",
	}
	if err := tx.Create(&consumeRecord).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create consume record"))
		return
	}

	// æ£æ¥æ¯å¦å·²ç»æ¯å¥½å
	var existingFriend models.Friend
	err = tx.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, selectedUser.ID, selectedUser.ID, userID).First(&existingFriend).Error

	if err == nil {
		// å·²ç»æ¯å¥½å?
		tx.Commit()
		c.JSON(http.StatusOK, utils.Success(gin.H{
			"message":         "Haidilao succeeded, already friends",
			"friend":          selectedUser,
			"coins_spent":     req.CostCoins,
			"already_friends": true,
		}))
		return
	}

	// åå»ºååå¥½åå³ç³»
	friend1 := models.Friend{
		UserID:   userID.(uint),
		FriendID: selectedUser.ID,
		Status:   "accepted", // æµ·åºæç´æ¥æä¸ºå¥½å?
	}

	friend2 := models.Friend{
		UserID:   selectedUser.ID,
		FriendID: userID.(uint),
		Status:   "accepted",
	}

	if err := tx.Create(&friend1).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create friend relationship"))
		return
	}

	if err := tx.Create(&friend2).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create friend relationship"))
		return
	}

	// åå»ºç¢°æè®°å½ï¼éè¿æµ·åºææ¹å¼ï¼
	record := models.CollisionRecord{
		UserID1:           userID.(uint),
		UserID2:           selectedUser.ID,
		Tag:               req.Tag,
		MatchType:         "haidilao",
		Status:            "friend_added",
		AddFriendDeadline: time.Now(), // å·²ç»å å¥½åï¼æªæ­¢æ¶é´è®¾ä¸ºå½å
	}

	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create collision record"))
		return
	}

	// æäº¤äºå¡
	tx.Commit()

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"message":         "æµ·åºææåï¼",
		"friend":          selectedUser,
		"coins_spent":     req.CostCoins,
		"match_id":        record.ID,
		"already_friends": false,
	}))
}

// è·åç¨æ·å½åçç¢°æç ä¿¡æ¯
func (cc *CollisionController) GetMyCollisionCode(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	// æ¥æ¾ç¨æ·ææ°çç¢°æç ?
	var collisionCode models.CollisionCode
	err := config.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&collisionCode).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, utils.Success(gin.H{
				"has_code": false,
				"message":  "No collision code",
			}))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to get collision code"))
		return
	}

	// æ£æ¥è¯¥æ ç­¾æ¯å¦æ¯é»æ´ç¶æ?
	var keyword models.HotTag
	isBlackhole := false
	if collisionCode.Tag != "" {
		err = config.DB.Where("keyword = ?", collisionCode.Tag).First(&keyword).Error
		if err == nil && keyword.Status == "blackhole" {
			isBlackhole = true
		}
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"has_code":     true,
		"code":         collisionCode,
		"is_blackhole": isBlackhole,
		"keyword_info": keyword,
	}))
}

// æç´¢ç¢°æç ï¼æ ¹æ®å³é®è¯ï¼
func (cc *CollisionController) SearchCollisionCodes(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var req struct {
		Keyword string `json:"keyword" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "è¯·è¾å¥å³é®è¯"))
		return
	}

	// æ¥æ¾å¹éçç¢°æç ï¼æé¤èªå·±ï¼
	var collisionCodes []models.CollisionCode
	err := config.DB.Where("tag = ? AND user_id != ? AND status != 'blackhole' AND status != 'invalid'",
		req.Keyword, userID).
		Preload("User").
		Order("created_at DESC").
		Limit(50). // éå¶è¿å50æ?
		Find(&collisionCodes).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "æç´¢å¤±è´¥"))
		return
	}

	// æå»ºè¿åæ°æ®ï¼åå«ç¨æ·åºæ¬ä¿¡æ?
	var results []gin.H
	for _, code := range collisionCodes {
		// éèå®æ´å¾®ä¿¡å·ï¼åªæ¾ç¤ºé¨å?
		wechatNo := code.User.WechatNo
		if len(wechatNo) > 4 {
			wechatNo = wechatNo[:2] + "***" + wechatNo[len(wechatNo)-2:]
		}

		results = append(results, gin.H{
			"id":        code.ID,
			"user_id":   code.UserID,
			"nickname":  code.User.Nickname,
			"avatar":    code.User.Avatar,
			"gender":    code.User.Gender,
			"age":       code.User.Age,
			"tag":       code.Tag,
			"location":  fmt.Sprintf("%s %s %s", code.Province, code.City, code.District),
			"wechat_no": wechatNo, // é¨åéèçå¾®ä¿¡å·
		})
	}

	c.JSON(http.StatusOK, utils.Success(results))
}

// åéå¥½åè¯·æ±ï¼éè¿æç´¢åç´æ¥æ·»å ï¼
func (cc *CollisionController) SendFriendRequest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var req struct {
		FriendID uint `json:"friend_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "è¯·æ±åæ°éè¯¯"))
		return
	}

	// ä¸è½æ·»å èªå·±ä¸ºå¥½å?
	if req.FriendID == userID.(uint) {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Cannot add yourself"))
		return
	}

	// æ£æ¥å¯¹æ¹æ¯å¦å­å?
	var targetUser models.User
	if err := config.DB.First(&targetUser, req.FriendID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "User not found"))
		return
	}

	// æ£æ¥æ¯å¦å·²ç»æ¯å¥½å
	var existingFriend models.Friend
	err := config.DB.Where(
		"(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, req.FriendID, req.FriendID, userID,
	).First(&existingFriend).Error

	if err == nil {
		c.JSON(http.StatusOK, utils.Success(gin.H{
			"message": "ä½ ä»¬å·²ç»æ¯å¥½åäº",
			"friend":  targetUser,
		}))
		return
	}

	// å¼å§äºå?
	tx := config.DB.Begin()

	// åå»ºååå¥½åå³ç³»
	friend1 := models.Friend{
		UserID:   userID.(uint),
		FriendID: req.FriendID,
		Status:   "accepted", // ç´æ¥æä¸ºå¥½å
	}

	friend2 := models.Friend{
		UserID:   req.FriendID,
		FriendID: userID.(uint),
		Status:   "accepted",
	}

	if err := tx.Create(&friend1).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "æ·»å å¥½åå¤±è´¥"))
		return
	}

	if err := tx.Create(&friend2).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "æ·»å å¥½åå¤±è´¥"))
		return
	}

	// æäº¤äºå¡
	tx.Commit()

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"message": "æ·»å å¥½åæå",
		"friend":  targetUser,
	}))
}

// æ¹éæäº¤ç¢°æç ?
func (cc *CollisionController) BatchSubmitCodes(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var req struct {
		Codes []struct {
			Country   string `json:"country" binding:"required"`
			Province  string `json:"province" binding:"required"`
			City      string `json:"city" binding:"required"`
			District  string `json:"district"`
			Tag       string `json:"tag"`
			Gender    int    `json:"gender"`
			AgeMin    int    `json:"age_min"`
			AgeMax    int    `json:"age_max"`
			CostCoins int    `json:"cost_coins"`
		} `json:"codes" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "è¯·æ±åæ°éè¯¯"))
		return
	}

	// æ£æ¥æ°ééå?
	if len(req.Codes) > 50 {
		c.JSON(http.StatusBadRequest, utils.Error(400, "åæ¬¡æå¤æäº?0ä¸ªç¢°æç "))
		return
	}

	// è·åç¨æ·ä¿¡æ¯
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "User not found"))
		return
	}

	// ¼ÆËã×ÜÏûºÄ£¨Ã¿Ìõ¹Ì¶¨10£©
	perCost := 10
	totalCost := perCost * len(req.Codes)

	// ¼ì²éÓà¶î

	if user.Coins < totalCost {
		c.JSON(http.StatusBadRequest, utils.Error(400, fmt.Sprintf(
			"ç¢°æå¸ä¸è¶? å½åä½é¢: %d, éè¦? %d",
			user.Coins,
			totalCost,
		)))
		return
	}

	// å¼å§äºå?
	tx := config.DB.Begin()

	// æ£é¤ç¢°æå¸?
	if err := tx.Model(&user).UpdateColumn("coins", gorm.Expr("coins - ?", totalCost)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to deduct coins"))
		return
	}

	// æ¹éåå»ºç¢°æç ?
	successCount := 0
	failedCodes := []string{}

	for i, codeReq := range req.Codes {
		collisionCode := models.CollisionCode{
			UserID:   userID.(uint),
			Tag:      codeReq.Tag,
			Country:  codeReq.Country,
			Province: codeReq.Province,
			City:     codeReq.City,
			District: codeReq.District,
			Gender:   codeReq.Gender,
			AgeMin:   codeReq.AgeMin,
			AgeMax:   codeReq.AgeMax,
			Status:   "active",
			// 直接设置为 pending 状态，确保首页显示需要审核
			// 但状态为 active，确保立即参与匹配
			AuditStatus: "pending",
			ExpiresAt:   time.Now().Add(24 * time.Hour), // 24å°æ¶åè¿æ?
			CostCoins:   perCost,
		}

		if err := tx.Create(&collisionCode).Error; err != nil {
			log.Printf("æ¹éåå»ºç¢°æç å¤±è´?- Index: %d, Error: %v", i, err)
			failedCodes = append(failedCodes, fmt.Sprintf("#%d", i+1))
			continue
		}

		successCount++

		// æ´æ°ç­é¨å³é®è¯ç»è®?å¦æææ ç­?
		if codeReq.Tag != "" {
			var keyword models.HotTag
			err := tx.Where("keyword = ?", codeReq.Tag).First(&keyword).Error
			if err == nil {
				// å³é®è¯å·²å­å¨,å¢å è®¡æ°
				tx.Model(&keyword).UpdateColumn("submit_count", gorm.Expr("submit_count + ?", 1))
			} else {
				// å³é®è¯ä¸å­å¨,åå»ºæ°ç
				newKeyword := models.HotTag{
					Keyword:     codeReq.Tag,
					Status:      "hide",
					SubmitCount: 1,
				}
				tx.Create(&newKeyword)
			}
		}
	}

	// å¦æå¨é¨å¤±è´¥,åæ»äºå¡
	if successCount == 0 {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "æ¹éæäº¤å¤±è´¥"))
		return
	}

	// æäº¤äºå¡
	tx.Commit()

	// æå»ºååºæ¶æ¯
	message := fmt.Sprintf("æåæäº¤%dä¸ªç¢°æç ", successCount)
	if len(failedCodes) > 0 {
		message += fmt.Sprintf(", %d failed", len(failedCodes))
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"message":       message,
		"success_count": successCount,
		"failed_count":  len(failedCodes),
		"total_cost":    totalCost,
		"new_balance":   user.Coins - totalCost,
	}))
}
