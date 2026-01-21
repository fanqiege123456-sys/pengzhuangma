//go:build scripts
// +build scripts

package main

import (
	"fmt"

	"collision-backend/config"
	"collision-backend/models"
)

func main() {
	// 初始化配置和数据库连�?	config.Init()

	// 查询所有热门标�?	var tags []models.HotTag
	config.DB.Find(&tags)

	fmt.Println("热门标签数据:")
	for _, tag := range tags {
		fmt.Printf("关键�? %-5s | 24h: %3d | 总搜�? %4d\n", tag.Keyword, tag.Count24h, tag.CountTotal)
	}

	// 专门查询24小时热门标签
	fmt.Println("\n24小时热门标签(按count_24h排序):")
	var hotTags24h []models.HotTag
	config.DB.Where("count_24h > 0").Order("count_24h DESC").Find(&hotTags24h)

	for _, tag := range hotTags24h {
		fmt.Printf("关键�? %-5s | 24h: %3d | 总搜�? %4d\n", tag.Keyword, tag.Count24h, tag.CountTotal)
	}
}