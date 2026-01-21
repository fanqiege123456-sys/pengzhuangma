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

	// 针对每个关键词执行单独的UPDATE语句
	for _, tagData := range hotTags24h {
		// 先查询当前�?		type TagInfo struct {
			Keyword   string
			Count24h  int
		}

		var tagInfo TagInfo
		config.DB.Table("hot_tags").
			Select("keyword, count_24h").
			Where("keyword = ?", tagData.Keyword).
			First(&tagInfo)

		fmt.Printf("更新�?%s: count_24h=%d\n", tagInfo.Keyword, tagInfo.Count24h)

		// 执行更新
		sql := "UPDATE hot_tags SET count_24h = ?, last_search_at = ? WHERE keyword = ?"
		result := config.DB.Exec(sql, tagData.Count24h, time.Now(), tagData.Keyword)
		if result.Error != nil {
			fmt.Printf("更新失败 %s: %v\n", tagData.Keyword, result.Error)
			continue
		}

		fmt.Printf("更新 %s: 影响 %d 行\n", tagData.Keyword, result.RowsAffected)

		// 验证更新结果
		var updatedTag TagInfo
		config.DB.Table("hot_tags").
			Select("keyword, count_24h").
			Where("keyword = ?", tagData.Keyword).
			First(&updatedTag)

		fmt.Printf("更新�?%s: count_24h=%d\n", updatedTag.Keyword, updatedTag.Count24h)
	}

	// 最后测�?4小时热门标签查询
	fmt.Println("\n最终测�?4小时热门标签查询:")
	type HotTag struct {
		Keyword   string
		Count24h  int
		CountTotal int
	}

	var tags []HotTag
	config.DB.Table("hot_tags").
		Select("keyword, count_24h, count_total").
		Where("count_24h > 0").
		Order("count_24h DESC").
		Find(&tags)

	if len(tags) > 0 {
		for _, tag := range tags {
			fmt.Printf("关键�? %-5s | 24h: %3d | 总搜�? %4d\n", tag.Keyword, tag.Count24h, tag.CountTotal)
		}
	} else {
		fmt.Println("没有找到24小时热门标签")
	}
}