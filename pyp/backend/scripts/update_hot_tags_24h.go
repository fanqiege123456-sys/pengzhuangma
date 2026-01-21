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

	// 24小时热门标签数据
	hotTags24h := []struct {
		Keyword  string
		Count24h int
	}{
		{"美食", 120},
		{"游戏", 110},
		{"旅行", 95},
		{"电影", 88},
		{"科技", 82},
		{"音乐", 76},
		{"宠物", 73},
		{"运动", 65},
		{"阅读", 58},
		{"摄影", 49},
	}

	// 更新24小时热门标签数据
	for _, tagData := range hotTags24h {
		// 查找标签
		var tag models.HotTag
		result := config.DB.Where("keyword = ?", tagData.Keyword).First(&tag)
		
		if result.Error != nil {
			fmt.Printf("标签不存�?%s\n", tagData.Keyword)
			continue
		}
		
		// 更新24小时搜索�?		tag.Count24h = tagData.Count24h
		tag.LastSearchAt = &[]time.Time{time.Now()}[0]
		
		if err := config.DB.Save(&tag).Error; err != nil {
			fmt.Printf("更新标签失败 %s: %v\n", tag.Keyword, err)
		} else {
			fmt.Printf("更新24小时热门标签成功 %s: %d\n", tag.Keyword, tag.Count24h)
		}
	}

	fmt.Println("24小时热门标签数据更新完成�?)
}