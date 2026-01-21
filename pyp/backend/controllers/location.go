package controllers

import (
	"collision-backend/config"
	"collision-backend/models"
	"collision-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LocationController struct{}

// 获取用户所有地址
func (lc *LocationController) GetLocations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var locations []models.UserLocation
	if err := config.DB.Where("user_id = ?", userID).Order("is_default DESC, created_at DESC").Find(&locations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to get locations"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(locations))
}

// 创建新地址
func (lc *LocationController) CreateLocation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	var req struct {
		Label     string `json:"label" binding:"required,oneof=home school work other"`
		Country   string `json:"country"`
		Province  string `json:"province"`
		City      string `json:"city"`
		District  string `json:"district"`
		IsDefault bool   `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request: "+err.Error()))
		return
	}

	// 如果设置为默认地址,先将其他地址设为非默认
	if req.IsDefault {
		config.DB.Model(&models.UserLocation{}).Where("user_id = ?", userID).Update("is_default", false)
	}

	location := models.UserLocation{
		UserID:    userID.(uint),
		Label:     req.Label,
		Country:   req.Country,
		Province:  req.Province,
		City:      req.City,
		District:  req.District,
		IsDefault: req.IsDefault,
	}

	if err := config.DB.Create(&location).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to create location"))
		return
	}

	// 如果是默认地址,更新用户表中的地址信息
	if req.IsDefault {
		config.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
			"country":  req.Country,
			"province": req.Province,
			"city":     req.City,
			"district": req.District,
		})
	}

	c.JSON(http.StatusOK, utils.Success(location))
}

// 更新地址
func (lc *LocationController) UpdateLocation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	locationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid location ID"))
		return
	}

	var location models.UserLocation
	if err := config.DB.Where("id = ? AND user_id = ?", locationID, userID).First(&location).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Location not found"))
		return
	}

	var req struct {
		Label     string `json:"label" binding:"omitempty,oneof=home school work other"`
		Country   string `json:"country"`
		Province  string `json:"province"`
		City      string `json:"city"`
		District  string `json:"district"`
		IsDefault bool   `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid request"))
		return
	}

	// 如果设置为默认地址,先将其他地址设为非默认
	if req.IsDefault && !location.IsDefault {
		config.DB.Model(&models.UserLocation{}).Where("user_id = ? AND id != ?", userID, locationID).Update("is_default", false)
	}

	updates := make(map[string]interface{})
	if req.Label != "" {
		updates["label"] = req.Label
	}
	if req.Country != "" {
		updates["country"] = req.Country
	}
	if req.Province != "" {
		updates["province"] = req.Province
	}
	if req.City != "" {
		updates["city"] = req.City
	}
	if req.District != "" {
		updates["district"] = req.District
	}
	updates["is_default"] = req.IsDefault

	if err := config.DB.Model(&location).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to update location"))
		return
	}

	// 如果是默认地址,更新用户表中的地址信息
	if req.IsDefault {
		config.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
			"country":  req.Country,
			"province": req.Province,
			"city":     req.City,
			"district": req.District,
		})
	}

	// 重新查询更新后的数据
	config.DB.First(&location, locationID)
	c.JSON(http.StatusOK, utils.Success(location))
}

// 删除地址
func (lc *LocationController) DeleteLocation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	locationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid location ID"))
		return
	}

	var location models.UserLocation
	if err := config.DB.Where("id = ? AND user_id = ?", locationID, userID).First(&location).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Location not found"))
		return
	}

	// 不允许删除默认地址,除非只剩一个地址
	if location.IsDefault {
		var count int64
		config.DB.Model(&models.UserLocation{}).Where("user_id = ?", userID).Count(&count)
		if count > 1 {
			c.JSON(http.StatusBadRequest, utils.Error(400, "Cannot delete default location. Please set another location as default first."))
			return
		}
	}

	if err := config.DB.Delete(&location).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(500, "Failed to delete location"))
		return
	}

	c.JSON(http.StatusOK, utils.Success(gin.H{"message": "Location deleted successfully"}))
}

// 设置默认地址
func (lc *LocationController) SetDefaultLocation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Error(401, "User not authenticated"))
		return
	}

	locationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(400, "Invalid location ID"))
		return
	}

	var location models.UserLocation
	if err := config.DB.Where("id = ? AND user_id = ?", locationID, userID).First(&location).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.Error(404, "Location not found"))
		return
	}

	// 将其他地址设为非默认，使用 map 确保 bool 零值能正确更新
	config.DB.Model(&models.UserLocation{}).Where("user_id = ?", userID).Updates(map[string]interface{}{"is_default": false})

	// 设置当前地址为默认，使用 map 确保 bool 值能正确更新
	config.DB.Model(&location).Updates(map[string]interface{}{"is_default": true})

	// 更新用户表中的地址信息
	config.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"country":  location.Country,
		"province": location.Province,
		"city":     location.City,
		"district": location.District,
	})

	// 重新获取更新后的 location 数据
	config.DB.First(&location, locationID)
	c.JSON(http.StatusOK, utils.Success(location))
}
