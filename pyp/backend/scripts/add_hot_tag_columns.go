//go:build scripts
// +build scripts

package main

import (
	"fmt"

	"collision-backend/config"
)

func main() {
	// 初始化配置和数据库连�?	config.Init()

	// 执行SQL语句添加缺失的字�?	sqlStatements := []string{
		// 添加count_24h字段
		"ALTER TABLE hot_tags ADD COLUMN count_24h INT DEFAULT 0",
		// 添加count_total字段（如果不存在�?		"ALTER TABLE hot_tags ADD COLUMN count_total INT DEFAULT 0",
		// 添加last_search_at字段
		"ALTER TABLE hot_tags ADD COLUMN last_search_at DATETIME",
		// 添加索引
		"ALTER TABLE hot_tags ADD INDEX idx_count_24h (count_24h)",
		"ALTER TABLE hot_tags ADD INDEX idx_count_total (count_total)",
	}

	for _, sql := range sqlStatements {
		result := config.DB.Exec(sql)
		if result.Error != nil {
			// 忽略重复字段错误
			if result.Error.Error() != "Error 1060 (42S21): Duplicate column name 'count_24h'" &&
			   result.Error.Error() != "Error 1060 (42S21): Duplicate column name 'count_total'" &&
			   result.Error.Error() != "Error 1060 (42S21): Duplicate column name 'last_search_at'" &&
			   result.Error.Error() != "Error 1061 (42000): Duplicate key name 'idx_count_24h'" &&
			   result.Error.Error() != "Error 1061 (42000): Duplicate key name 'idx_count_total'" {
				fmt.Printf("执行SQL失败: %s\n错误: %v\n", sql, result.Error)
			} else {
				fmt.Printf("字段/索引已存在，跳过: %s\n", sql)
			}
		} else {
			fmt.Printf("执行SQL成功: %s\n", sql)
		}
	}

	fmt.Println("\n字段添加完成�?)
}