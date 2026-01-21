package controllers

import (
	"net/http"
	"strconv"
	"time"

	"collision-backend/config"

	"github.com/gin-gonic/gin"
)

// SparkController 火花控制器
type SparkController struct{}

// GetCollisionSparks 获取碰撞动态列表
func (sc *SparkController) GetCollisionSparks(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	// 获取请求参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	keyword := c.Query("keyword") // 新增关键词筛选参数

	// 计算分页参数
	offset := (page - 1) * size
	startDate := time.Now().AddDate(0, 0, -days)

	// 定义碰撞动态结构体
	type CollisionSpark struct {
		ID            uint      `json:"id"`
		UserID        uint      `json:"user_id"`
		UserNickname  string    `json:"user_nickname"`
		Avatar        string    `json:"avatar"`
		Keyword       string    `json:"keyword"`
		CollisionTime string    `json:"collision_time"`
		MatchCount    int       `json:"match_count"`
		CreatedAt     time.Time `json:"created_at"`
	}

	// 查询碰撞动态数据
	var sparks []CollisionSpark
	var total int64

	// 构建查询，添加软删除检查和审核状态检查
	query := config.DB.Table("collision_codes").
		Joins("LEFT JOIN users ON collision_codes.user_id = users.id").
		Where("collision_codes.created_at >= ? AND collision_codes.status = 'active' AND collision_codes.audit_status = 'approved' AND collision_codes.deleted_at IS NULL", startDate)

	// 关键词筛选
	if keyword != "" {
		query = query.Where("collision_codes.tag LIKE ?", "%"+keyword+"%")
	}

	// 查询总数
	query.Count(&total)

	// 查询分页数据
	query.Select("collision_codes.id, collision_codes.user_id, users.nickname as user_nickname, users.avatar, collision_codes.tag as keyword, collision_codes.created_at, collision_codes.match_count").
		Order("collision_codes.created_at DESC").
		Offset(offset).
		Limit(size).
		Scan(&sparks)

	// 格式化碰撞时间为中文格式
	for i := range sparks {
		sparks[i].CollisionTime = sparks[i].CreatedAt.Format("2006年01月02日 15:04:05")
	}

	// 构建响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list": sparks,
			"pagination": gin.H{
				"page":  page,
				"size":  size,
				"total": total,
			},
		},
	})
}
