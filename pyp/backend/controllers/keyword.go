package controllers

import (
	"collision-backend/config"
	"collision-backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type KeywordController struct{}

type KeywordRequest struct {
	Keyword string `json:"keyword" binding:"required"`
	Status  string `json:"status" binding:"required,oneof=show hide blackhole"`
}

type KeywordStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=show hide blackhole"`
}

// 获取热门关键词列表
func (ctrl *KeywordController) GetKeywords(c *gin.Context) {
	var keywords []models.HotTag

	result := config.DB.Order("submit_count DESC, created_at DESC").Find(&keywords)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取关键词列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": keywords,
	})
}

// 创建热门关键词
func (ctrl *KeywordController) CreateKeyword(c *gin.Context) {
	var req KeywordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误",
		})
		return
	}

	// 检查关键词是否已存在
	var existingKeyword models.HotTag
	if err := config.DB.Where("keyword = ?", req.Keyword).First(&existingKeyword).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "关键词已存在",
		})
		return
	}

	keyword := models.HotTag{
		Keyword:     req.Keyword,
		Status:      req.Status,
		SubmitCount: 0,
	}

	if err := config.DB.Create(&keyword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建关键词失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "创建成功",
		"data": keyword,
	})
}

// 更新关键词状态
func (ctrl *KeywordController) UpdateKeywordStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的关键词ID",
		})
		return
	}

	var req KeywordStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误",
		})
		return
	}

	var keyword models.HotTag
	if err := config.DB.First(&keyword, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "关键词不存在",
		})
		return
	}

	keyword.Status = req.Status
	if err := config.DB.Save(&keyword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新状态失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
		"data": keyword,
	})
}

// 删除关键词
func (ctrl *KeywordController) DeleteKeyword(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的关键词ID",
		})
		return
	}

	var keyword models.HotTag
	if err := config.DB.First(&keyword, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "关键词不存在",
		})
		return
	}

	if err := config.DB.Delete(&keyword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "删除失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}
