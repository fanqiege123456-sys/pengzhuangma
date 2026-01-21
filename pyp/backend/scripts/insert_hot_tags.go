//go:build scripts
// +build scripts

package main

import (
	"fmt"
	"time"

	"collision-backend/config"
	"collision-backend/models"
)

func main() {
	// 初始化配置和数据库连�?	config.Init()

	// 热门标签数据
	hotTags := []models.HotTag{
		{
			Keyword:      "美食",
			Count24h:     120,
			CountTotal:   1500,
			LastSearchAt: &[]time.Time{time.Now()}[0],
		},
		{
			Keyword:      "旅行",
			Count24h:     95,
			CountTotal:   1200,
			LastSearchAt: &[]time.Time{time.Now()}[0],
		},
		{
			Keyword:      "电影",
			Count24h:     88,
			CountTotal:   1050,
			LastSearchAt: &[]time.Time{time.Now()}[0],
		},
		{
			Keyword:      "音乐",
			Count24h:     76,
			CountTotal:   980,
			LastSearchAt: &[]time.Time{time.Now()}[0],
		},
		{
			Keyword:      "运动",
			Count24h:     65,
			CountTotal:   890,
			LastSearchAt: &[]time.Time{time.Now()}[0],
		},
		{
			Keyword:      "阅读",
			Count24h:     58,
			CountTotal:   760,
			LastSearchAt: &[]time.Time{time.Now()}[0],
		},
		{
			Keyword:      "摄影",
			Count24h:     49,
			CountTotal:   680,
			LastSearchAt: &[]time.Time{time.Now()}[0],
		},
		{
			Keyword:      "游戏",
			Count24h:     110,
			CountTotal:   1350,
			LastSearchAt: &[]time.Time{time.Now()}[0],
		},
		{
			Keyword:      "科技",
			Count24h:     82,
			CountTotal:   950,
			LastSearchAt: &[]time.Time{time.Now()}[0],
		},
		{
			Keyword:      "宠物",
			Count24h:     73,
			CountTotal:   870,
			LastSearchAt: &[]time.Time{time.Now()}[0],
		},
	}

	// 插入热门标签数据
	for _, tag := range hotTags {
		// 先检查是否已存在该标�?		var existingTag models.HotTag
		result := config.DB.Where("keyword = ?", tag.Keyword).First(&existingTag)
		
		if result.Error != nil {
			// 不存在，创建新标�?			if err := config.DB.Create(&tag).Error; err != nil {
				fmt.Printf("插入标签失败 %s: %v\n", tag.Keyword, err)
			} else {
				fmt.Printf("插入标签成功 %s\n", tag.Keyword)
			}
		} else {
			// 已存在，更新统计数据
			existingTag.Count24h += tag.Count24h
			existingTag.CountTotal += tag.CountTotal
			existingTag.LastSearchAt = tag.LastSearchAt
			if err := config.DB.Save(&existingTag).Error; err != nil {
				fmt.Printf("更新标签失败 %s: %v\n", existingTag.Keyword, err)
			} else {
				fmt.Printf("更新标签成功 %s\n", existingTag.Keyword)
			}
		}
	}

	fmt.Println("热门标签数据插入完成�?)
}