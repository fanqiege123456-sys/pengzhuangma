package controllers

import (
	"net/http"
	"strings"

	"collision-backend/config"
	"collision-backend/models"

	"github.com/gin-gonic/gin"
)

type ForbiddenKeywordController struct{}

type ForbiddenKeywordRequest struct {
	Keyword string `json:"keyword" binding:"required"`
}

// GetForbiddenKeywords 获取违禁词列表
func (ctrl *ForbiddenKeywordController) GetForbiddenKeywords(c *gin.Context) {
	var keywords []models.ForbiddenKeyword

	if err := config.DB.Order("created_at DESC").Find(&keywords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取违禁词列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": keywords,
	})
}

// CreateForbiddenKeyword 创建违禁词
func (ctrl *ForbiddenKeywordController) CreateForbiddenKeyword(c *gin.Context) {
	var req ForbiddenKeywordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误",
		})
		return
	}

	req.Keyword = strings.TrimSpace(req.Keyword)
	if req.Keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请输入违禁词",
		})
		return
	}

	var existing models.ForbiddenKeyword
	if err := config.DB.Where("keyword = ?", req.Keyword).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "违禁词已存在",
		})
		return
	}

	keyword := models.ForbiddenKeyword{
		Keyword: req.Keyword,
	}

	if err := config.DB.Create(&keyword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建违禁词失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "创建成功",
		"data": keyword,
	})
}

// DeleteForbiddenKeyword 删除违禁词
func (ctrl *ForbiddenKeywordController) DeleteForbiddenKeyword(c *gin.Context) {
	id := c.Param("id")

	var keyword models.ForbiddenKeyword
	if err := config.DB.First(&keyword, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "违禁词不存在",
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
