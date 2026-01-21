//go:build scripts
// +build scripts

package main

import (
	"fmt"
	"time"

	"collision-backend/config"
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
		// 使用更可靠的更新方式，直接指定要更新的字�?		result := config.DB.Model(&struct{}{}).Table("hot_tags").
			Where("keyword = ?", tagData.Keyword).
			Updates(map[string]interface{}{
				"count_24h":      tagData.Count24h,
				"last_search_at": time.Now(),
			})
		
		if result.Error != nil {
			fmt.Printf("更新标签失败 %s: %v\n", tagData.Keyword, result.Error)
		} else {
			if result.RowsAffected > 0 {
				fmt.Printf("更新24小时热门标签成功 %s: %d\n", tagData.Keyword, tagData.Count24h)
			} else {
				fmt.Printf("标签不存�?%s\n", tagData.Keyword)
			}
		}
	}

	// 验证更新结果
	fmt.Println("\n验证24小时热门标签数据:")
	type HotTag struct {
		Keyword   string
		Count24h  int
		CountTotal int
	}

	var tags []HotTag
	config.DB.Table("hot_tags").
		Select("keyword, count_24h, count_total").
		Order("count_24h DESC").
		Find(&tags)

	for _, tag := range tags {
		fmt.Printf("关键�? %-5s | 24h: %3d | 总搜�? %4d\n", tag.Keyword, tag.Count24h, tag.CountTotal)
	}
}