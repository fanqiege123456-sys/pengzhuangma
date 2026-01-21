package controllers

import (
	"net/http"
	"strconv"

	"collision-backend/config"
	"collision-backend/models"
	"collision-backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AdminController struct{}

// 管理员登录
func (ac *AdminController) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request data"))
		return
	}

	var admin models.Admin
	if err := config.DB.Where("username = ? AND status = ?", req.Username, "active").First(&admin).Error; err != nil {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "Invalid credentials"))
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "Invalid credentials"))
		return
	}

	// 生成JWT token
	token, err := utils.GenerateToken(admin.ID, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to generate token"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"token": token,
		"admin": gin.H{
			"id":       admin.ID,
			"username": admin.Username,
			"email":    admin.Email,
			"role":     admin.Role,
		},
	}))
}

// 获取管理员列表
func (ac *AdminController) GetAdmins(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var admins []models.Admin
	var total int64

	offset := (page - 1) * pageSize

	config.DB.Model(&models.Admin{}).Count(&total)
	config.DB.Select("id, username, email, role, status, created_at, updated_at").
		Offset(offset).Limit(pageSize).Find(&admins)

	c.JSON(http.StatusOK, utils.Success(utils.PageData{
		List: admins,
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}))
}

// 创建管理员
func (ac *AdminController) CreateAdmin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request data"))
		return
	}

	// 检查用户名是否存在
	var count int64
	config.DB.Model(&models.Admin{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, utils.Error(409, "Username already exists"))
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to hash password"))
		return
	}

	admin := models.Admin{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Role:     req.Role,
		Status:   "active",
	}

	if err := config.DB.Create(&admin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create admin"))
		return
	}

	// 清除密码字段
	admin.Password = ""

	c.JSON(http.StatusOK, utils.Success(admin))
}

// 更新管理员状态
func (ac *AdminController) UpdateAdminStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request data"))
		return
	}

	var admin models.Admin
	if err := config.DB.First(&admin, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Admin not found"))
		return
	}

	// 不允许修改超级管理员(admin用户)的状态
	if admin.Username == "admin" {
		c.JSON(http.StatusForbidden, utils.Error(403, "Cannot modify super admin status"))
		return
	}

	// 更新状态
	if err := config.DB.Model(&admin).Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to update admin status"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"message": "Status updated successfully",
		"admin":   admin,
	}))
}

// 删除管理员
func (ac *AdminController) DeleteAdmin(c *gin.Context) {
	id := c.Param("id")

	var admin models.Admin
	if err := config.DB.First(&admin, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Admin not found"))
		return
	}

	// 不允许删除超级管理员
	if admin.Username == "admin" {
		c.JSON(http.StatusForbidden, utils.Error(403, "Cannot delete super admin"))
		return
	}

	// 检查当前超级管理员数量(role='super')
	var superAdminCount int64
	config.DB.Model(&models.Admin{}).Where("role = ? AND status = ?", "super", "active").Count(&superAdminCount)

	// 如果删除的是超级管理员,且剩余超级管理员少于3个,则禁止删除
	if admin.Role == "super" && superAdminCount <= 3 {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Cannot delete super admin when count is 3 or less"))
		return
	}

	// 软删除(将状态改为deleted)
	if err := config.DB.Model(&admin).Update("status", "deleted").Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to delete admin"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{
		"message": "Admin deleted successfully",
	}))
}
