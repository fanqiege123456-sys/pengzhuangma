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

	// 执行SQL语句直接更新数据
	sql := `
	UPDATE hot_tags SET 
	    count_24h = CASE keyword 
	        WHEN '美食' THEN 120
	        WHEN '游戏' THEN 110
	        WHEN '旅行' THEN 95
	        WHEN '电影' THEN 88
	        WHEN '科技' THEN 82
	        WHEN '音乐' THEN 76
	        WHEN '宠物' THEN 73
	        WHEN '运动' THEN 65
	        WHEN '阅读' THEN 58
	        WHEN '摄影' THEN 49
	        ELSE count_24h
	    END,
	    count_total = CASE keyword 
	        WHEN '美食' THEN 1500
	        WHEN '游戏' THEN 1350
	        WHEN '旅行' THEN 1200
	        WHEN '电影' THEN 1050
	        WHEN '科技' THEN 950
	        WHEN '音乐' THEN 980
	        WHEN '宠物' THEN 870
	        WHEN '运动' THEN 890
	        WHEN '阅读' THEN 760
	        WHEN '摄影' THEN 680
	        ELSE count_total
	    END,
	    last_search_at = ?
	`

	result := config.DB.Exec(sql, time.Now())
	if result.Error != nil {
		fmt.Printf("执行SQL失败: %v\n", result.Error)
		return
	}

	fmt.Printf("更新�?%d 条记录\n", result.RowsAffected)

	// 验证更新结果
	fmt.Println("\n验证24小时热门标签数据:")
	type HotTag struct {
		Keyword    string
		Count24h   int
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

	// 测试24小时热门标签查询
	fmt.Println("\n测试24小时热门标签查询:")
	var hotTags24h []HotTag
	config.DB.Table("hot_tags").
		Select("keyword, count_24h, count_total").
		Where("count_24h > 0").
		Order("count_24h DESC").
		Limit(10).
		Find(&hotTags24h)

	if len(hotTags24h) > 0 {
		for _, tag := range hotTags24h {
			fmt.Printf("关键�? %-5s | 24h: %3d | 总搜�? %4d\n", tag.Keyword, tag.Count24h, tag.CountTotal)
		}
	} else {
		fmt.Println("没有找到24小时热门标签")
	}
}