package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"collision-backend/config"
	"collision-backend/models"
	"collision-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct{}

// 获取用户列表
func (uc *UserController) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var users []models.User
	var total int64

	offset := (page - 1) * pageSize

	config.DB.Model(&models.User{}).Count(&total)
	config.DB.Offset(offset).Limit(pageSize).Find(&users)

	// 关联查询用户联系方式
	for i := range users {
		var contact models.UserContact
		if err := config.DB.Where("user_id = ?", users[i].ID).First(&contact).Error; err == nil {
			// 如果有联系方式记录，更新用户信息
			users[i].Phone = contact.Phone
			users[i].Email = contact.Email
		}
	}

	c.JSON(http.StatusOK, utils.Success(utils.PageData{
		List: users,
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}))
}

// 创建用户
func (uc *UserController) CreateUser(c *gin.Context) {
	var createData struct {
		WechatNo      string `json:"wechat_no" binding:"required"`
		Phone         string `json:"phone"`
		Email         string `json:"email"`
		Nickname      string `json:"nickname"`
		Coins         int    `json:"coins"`
		AllowForceAdd bool   `json:"allow_force_add"`
	}

	if err := c.ShouldBindJSON(&createData); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorWithMsg(utils.ValidationErrorCode, "请求参数错误: "+err.Error()))
		return
	}

	// 生成随机OpenID（管理后台创建用户时使用）
	openID := "admin_create_" + utils.GenerateRandomString(16)

	// 开始事务
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建用户
	user := models.User{
		OpenID:        openID,
		Nickname:      createData.Nickname,
		WechatNo:      createData.WechatNo,
		Phone:         createData.Phone,
		Coins:         createData.Coins,
		AllowForceAdd: createData.AllowForceAdd,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.ErrorWithMsg(utils.DatabaseErrorCode, "创建用户失败: "+err.Error()))
		return
	}

	// 创建或更新用户联系方式
	contact := models.UserContact{
		UserID:        uint64(user.ID),
		Phone:         createData.Phone,
		Email:         createData.Email,
		EmailVerified: true,
		PhoneVerified: true,
		EmailVisible:  true,
	}

	if err := tx.Save(&contact).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.ErrorWithMsg(utils.DatabaseErrorCode, "创建用户联系方式失败: "+err.Error()))
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorWithMsg(utils.DatabaseErrorCode, "提交事务失败: "+err.Error()))
		return
	}

	// 更新用户信息（包含联系方式）
	user.Phone = createData.Phone
	user.Email = createData.Email

	fmt.Printf("管理员创建用户成功: ID=%d, WechatNo=%s, Nickname=%s, Phone=%s, Email=%s\n",
		user.ID, user.WechatNo, user.Nickname, user.Phone, user.Email)
	c.JSON(http.StatusOK, utils.SuccessWithMsg(user, "创建用户成功"))
}

// 获取用户详情
func (uc *UserController) GetUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorWithDefaultMsg(utils.NotFoundCode))
		return
	}

	// 关联查询用户联系方式
	var contact models.UserContact
	if err := config.DB.Where("user_id = ?", user.ID).First(&contact).Error; err == nil {
		// 如果有联系方式记录，更新用户信息
		user.Phone = contact.Phone
		user.Email = contact.Email
	}

	c.JSON(http.StatusOK, utils.Success(user))
}

// 更新用户信息
func (uc *UserController) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "User not found"))
		return
	}

	var updateData struct {
		Phone         string `json:"phone"`
		Email         string `json:"email"`
		Coins         *int   `json:"coins"`
		AllowForceAdd *bool  `json:"allow_force_add"`
		Nickname      string `json:"nickname"`
		Age           *int   `json:"age"`
		Gender        *int   `json:"gender"`
		WechatNo      string `json:"wechat_no"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request data"))
		return
	}

	// 开始事务
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 使用 map 来更新，确保 bool 类型的零值也能正确更新
	updates := map[string]interface{}{
		"phone":     updateData.Phone,
		"wechat_no": updateData.WechatNo,
	}

	// 添加其他可选字段
	if updateData.Nickname != "" {
		updates["nickname"] = updateData.Nickname
	}
	if updateData.Coins != nil {
		updates["coins"] = *updateData.Coins
	}
	if updateData.Age != nil {
		updates["age"] = *updateData.Age
	}
	if updateData.Gender != nil {
		updates["gender"] = *updateData.Gender
	}
	if updateData.AllowForceAdd != nil {
		updates["allow_force_add"] = *updateData.AllowForceAdd
	}

	if err := tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to update user"))
		return
	}

	// 更新用户联系方式
	var contact models.UserContact
	result := tx.Where("user_id = ?", user.ID).First(&contact)
	if result.Error == nil {
		// 如果存在联系方式，更新
		contact.Phone = updateData.Phone
		contact.Email = updateData.Email
		if err := tx.Save(&contact).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to update user contact"))
			return
		}
	} else {
		// 如果不存在联系方式，创建
		contact = models.UserContact{
			UserID:        uint64(user.ID),
			Phone:         updateData.Phone,
			Email:         updateData.Email,
			EmailVerified: true,
			PhoneVerified: true,
			EmailVisible:  true,
		}
		if err := tx.Create(&contact).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create user contact"))
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to commit transaction"))
		return
	}

	// 重新获取更新后的用户数据
	config.DB.First(&user, id)
	// 获取完整联系方式
	var updatedContact models.UserContact
	if err := config.DB.Where("user_id = ?", user.ID).First(&updatedContact).Error; err == nil {
		user.Phone = updatedContact.Phone
		user.Email = updatedContact.Email
	}
	c.JSON(http.StatusOK, utils.Success(user))
}

// 删除用户
func (uc *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to delete user"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{"message": "User deleted successfully"}))
}

// 微信小程序登录
func (uc *UserController) WechatLogin(c *gin.Context) {
	var req struct {
		Code     string `json:"code" binding:"required"`
		UserInfo *struct {
			NickName  string `json:"nickName"`
			AvatarUrl string `json:"avatarUrl"`
			Gender    int    `json:"gender"`
			Country   string `json:"country"`
			Province  string `json:"province"`
			City      string `json:"city"`
			Language  string `json:"language"`
		} `json:"userInfo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request"))
		return
	}

	fmt.Printf("微信登录请求 - Code: %s, UserInfo: %+v\n", req.Code, req.UserInfo)

	// 调用微信API验证code，获取openid和session_key
	var openID string

	if config.Config.WechatAppID != "" && config.Config.WechatSecret != "" {
		// 使用真实的微信API
		wechatResp, err := uc.getWechatSession(req.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to verify wechat code: "+err.Error()))
			return
		}
		openID = wechatResp.OpenID
		fmt.Printf("微信API返回 - OpenID: %s\n", openID)
	} else {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "wechat config missing"))
		return
	}
// 查找或创建用户
	var user models.User
	if err := config.DB.Where("open_id = ?", openID).First(&user).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to query user"))
			return
		}
		if req.UserInfo == nil {
			c.JSON(http.StatusBadRequest, utils.Error(400, "Need wechat profile authorization"))
			return
		}
		// 用户不存在，创建新用户
		nickname := "微信用户"
		avatar := "https://mmbiz.qpic.cn/mmbiz/icTdbqWNOwNRna42FI242Lcia07jQodd2FJGIYQfG0LAJGFxM4FbnQP6yfMxBgJ0F3YRqJCJ1aPAK2dQagdusBZg/0"

		if req.UserInfo != nil {
			nickname = req.UserInfo.NickName
			avatar = req.UserInfo.AvatarUrl
		}

		user = models.User{
			OpenID:   openID,
			Nickname: nickname,
			Avatar:   avatar,
			WechatNo: "wx" + utils.GenerateRandomString(8),
			Coins:    1000, // 新用户赠送1000积分
		}

		// 如果提供了用户信息，保存地理位置
		if req.UserInfo != nil {
			user.Country = req.UserInfo.Country
			user.Province = req.UserInfo.Province
			user.City = req.UserInfo.City
			// Gender: 0未知, 1男, 2女
			if req.UserInfo.Gender == 1 {
				user.Gender = 1
			} else if req.UserInfo.Gender == 2 {
				user.Gender = 2
			}
		}

		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create user"))
			return
		}

		fmt.Printf("创建新用户成功: ID=%d, OpenID=%s, Nickname=%s\n", user.ID, user.OpenID, user.Nickname)
	} else {
		// 用户已存在，更新用户信息（如果提供了）
		fmt.Printf("用户已存在: ID=%d, OpenID=%s, Nickname=%s\n", user.ID, user.OpenID, user.Nickname)

		if req.UserInfo != nil {
			updates := make(map[string]interface{})
			updates["nickname"] = req.UserInfo.NickName
			updates["avatar"] = req.UserInfo.AvatarUrl
			updates["country"] = req.UserInfo.Country
			updates["province"] = req.UserInfo.Province
			updates["city"] = req.UserInfo.City

			if req.UserInfo.Gender == 1 {
				updates["gender"] = 1
			} else if req.UserInfo.Gender == 2 {
				updates["gender"] = 2
			}

			if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
				fmt.Printf("更新用户信息失败: %v\n", err)
			} else {
				// 重新加载用户信息
				config.DB.First(&user, user.ID)
				fmt.Printf("更新用户信息成功: %+v\n", user)
			}
		}
	}

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, "user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to generate token"))
		return
	}

	fmt.Printf("登录成功，生成token: %s\n", token)

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"token": token,
		"user":  user,
	}))
}

// 微信登录会话信息
type WechatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// 调用微信API获取session信息
func (uc *UserController) getWechatSession(code string) (*WechatSessionResponse, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		config.Config.WechatAppID,
		config.Config.WechatSecret,
		code,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result WechatSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error: %d %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// 获取当前用户信息
func (uc *UserController) GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "User not found"))
		return
	}

	// 关联查询用户联系方式
	var contact models.UserContact
	if err := config.DB.Where("user_id = ?", user.ID).First(&contact).Error; err == nil {
		// 如果有联系方式记录，更新用户信息
		user.Phone = contact.Phone
		user.Email = contact.Email
	}

	c.JSON(http.StatusOK, utils.Success(user))
}

// 更新用户资料
func (uc *UserController) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var req struct {
		Nickname      string `json:"nickname"`
		Avatar        string `json:"avatar"`
		Gender        *int   `json:"gender"`
		Age           int    `json:"age"`
		Bio           string `json:"bio"`
		WechatNo      string `json:"wechat_no"`
		AllowForceAdd *bool  `json:"allow_force_add"`
		AllowHaidilao *bool  `json:"allow_haidilao"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request"))
		return
	}

	// 使用 map 来更新，这样可以正确处理 bool 类型的 false 值
	updates := make(map[string]interface{})

	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	// 允许设置年龄为0（表示未设置）
	updates["age"] = req.Age
	// 总是更新bio字段，允许用户清空
	updates["bio"] = req.Bio
	if req.WechatNo != "" {
		updates["wechat_no"] = req.WechatNo
	}
	// 对于 bool 指针，只要不为 nil 就更新（可以正确处理 true 和 false）
	if req.AllowForceAdd != nil {
		updates["allow_force_add"] = *req.AllowForceAdd
	}
	if req.AllowHaidilao != nil {
		updates["allow_haidilao"] = *req.AllowHaidilao
	}

	// 如果没有需要更新的字段
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, utils.Error(400, "No fields to update"))
		return
	}

	// 使用 Updates 更新指定字段
	if err := config.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to update profile"))
		return
	}

	// 重新查询更新后的用户信息
	var user models.User
	config.DB.First(&user, userID)
	c.JSON(http.StatusOK, utils.Success(user))
}

// 获取用户余额
func (uc *UserController) GetBalance(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var user models.User
	if err := config.DB.Select("coins").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "User not found"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"balance": user.Coins,
	}))
}

// GetConsumeRecords 获取用户消费记录（分页）
func (uc *UserController) GetConsumeRecords(c *gin.Context) {
	// 从JWT获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	// 分页参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// 兼容前端可能传递的0值或负数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询消费记录
	var records []models.ConsumeRecord
	var total int64

	// 获取总数
	config.DB.Model(&models.ConsumeRecord{}).Where("user_id = ?", userID).Count(&total)

	// 获取分页数据
	if err := config.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to get consume records"))
		return
	}

	// 格式化数据为前端期望的格式
	formattedRecords := make([]gin.H, len(records))
	for i, record := range records {
		formattedRecords[i] = gin.H{
			"id":                  record.ID,
			"type":                record.Type,
			"type_display":        getTypeDisplay(record.Type),
			"description":         record.Reason,
			"amount":              record.Coins,
			"coins":               record.Coins, // 兼容前端期望的coins字段
			"reason":              record.Reason, // 兼容前端期望的reason字段
			"created_at":          record.CreatedAt,
			"created_at_display":  record.CreatedAt.Format("01-02 15:04"),
		}
	}

	// 返回包含分页信息的完整数据结构
	response := gin.H{
		"records":   formattedRecords,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     (int(total) + pageSize - 1) / pageSize,
	}

	c.JSON(http.StatusOK, utils.Success(response))
}

// CreateRechargeOrder 创建充值订单（本地开发模拟）
func (uc *UserController) CreateRechargeOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var req struct {
		Amount int `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid amount"))
		return
	}

	orderNo := fmt.Sprintf("RC%s%04d", time.Now().Format("20060102150405"), rand.Intn(10000))
	coins := req.Amount

	tx := config.DB.Begin()

	var user models.User
	if err := tx.First(&user, userID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, utils.Error(404, "User not found"))
		return
	}

	user.Coins += coins
	user.TotalRecharge += coins
	if err := tx.Model(&user).Updates(map[string]interface{}{
		"coins":          user.Coins,
		"total_recharge": user.TotalRecharge,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to update balance"))
		return
	}

	record := models.RechargeRecord{
		UserID:  userID.(uint),
		Amount:  req.Amount,
		Coins:   coins,
		OrderNo: orderNo,
		Status:  "success",
		PayType: "mock",
	}
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create recharge record"))
		return
	}

	consumeRecord := models.ConsumeRecord{
		UserID: userID.(uint),
		Coins:  coins,
		Type:   "recharge",
		Reason: "充值",
	}
	if err := tx.Create(&consumeRecord).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create consume record"))
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to commit transaction"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"order_id":  orderNo,
		"prepay_id": "mock_" + orderNo,
	}))
}

// getTypeDisplay 获取消费类型的显示文本
func getTypeDisplay(consumeType string) string {
	typeMap := map[string]string{
		"collision":        "碰撞提交",
		"collision_submit": "碰撞提交", 
		"renew_collision":  "续期碰撞",
		"force_add":        "强制添加",
		"match_reward":     "匹配奖励",
		"recharge":         "充值",
		"refund":           "退款",
		"system":           "系统调整",
		"haidilao":         "海底捞",
		"send_email":       "发送邮件",
	}
	
	if display, exists := typeMap[consumeType]; exists {
		return display
	}
	return consumeType
}

// GetRechargeRecords 获取用户充值记录（分页）
func (uc *UserController) GetRechargeRecords(c *gin.Context) {
	// 从JWT获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	// 分页参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// 兼容前端可能传递的0值或负数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询充值记录
	var records []models.RechargeRecord
	var total int64

	// 获取总数
	config.DB.Model(&models.RechargeRecord{}).Where("user_id = ?", userID).Count(&total)

	// 获取分页数据
	if err := config.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to get recharge records"))
		return
	}

	// 构建返回数据 - 确保格式与消费记录一致
	response := gin.H{
		"total":     total,                                  // 总记录数
		"items":     records,                                // 前端可能期望用items而不是records
		"records":   records,                                // 同时保留records，兼容其他前端
		"page":      page,                                   // 当前页码
		"page_size": pageSize,                               // 每页大小
		"size":      pageSize,                               // 同时保留size，兼容其他前端
		"current":   page,                                   // 兼容某些前端框架的current字段
		"pages":     (int(total) + pageSize - 1) / pageSize, // 总页数
	}

	c.JSON(http.StatusOK, utils.Success(response))
}

// 更新用户地址信息
func (uc *UserController) UpdateUserLocation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var req struct {
		Country         string `json:"country"`
		Province        string `json:"province"`
		City            string `json:"city"`
		District        string `json:"district"`
		AllowUpperLevel *bool  `json:"allow_upper_level"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request"))
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "User not found"))
		return
	}

	// 使用 map 来更新，确保 bool 类型的零值也能正确更新
	updates := map[string]interface{}{
		"country":  req.Country,
		"province": req.Province,
		"city":     req.City,
		"district": req.District,
	}

	if req.AllowUpperLevel != nil {
		updates["allow_upper_level"] = *req.AllowUpperLevel
	}

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to update user location"))
		return
	}

	// 重新获取更新后的用户数据
	config.DB.First(&user, userID)
	c.JSON(http.StatusOK, utils.Success(user))
}


