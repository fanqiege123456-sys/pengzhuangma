//go:build scripts
// +build scripts

package main

import (
	"collision-backend/config"
	"collision-backend/models"
	"fmt"
	"os"
)

func main() {
	// 设置数据库配�?	nSetEnv("DB_HOST", "localhost")
	nSetEnv("DB_PORT", "3306")
	nSetEnv("DB_USER", "fanfan00")
	nSetEnv("DB_PASSWORD", "Xuaner.123")
	nSetEnv("DB_NAME", "collision_db")

	// 初始化配�?	config.Init()

	fmt.Println("�?配置初始化完�?)

	// 查询最�?0条邮件日�?	var emailLogs []models.EmailLog
	config.DB.Order("created_at DESC").Limit(10).Find(&emailLogs)

	fmt.Printf("\n📋 最�?0条邮件日�?\n")
	fmt.Printf("%3s | %6s | %-25s | %-50s | %-10s | %s\n", "ID", "UserID", "收件�?, "主题", "状�?, "发送时�?)
	fmt.Println("--------------------------------------------------------------------------------------------------------------------------------")

	for _, log := range emailLogs {
		status := log.Status
		sentAt := "-"
		if log.SentAt != nil {
			sentAt = log.SentAt.Format("2006-01-02 15:04:05")
		}

		fmt.Printf("%3d | %6d | %-25s | %-50s | %-10s | %s\n", 
			log.ID, log.UserID, log.ToEmail, log.Subject, status, sentAt)
	}

	fmt.Println("--------------------------------------------------------------------------------------------------------------------------------")
	fmt.Printf("📊 共找�?%d 条邮件日志\n", len(emailLogs))

	// 查询碰撞匹配记录
	var collisionResults []models.CollisionResult
	config.DB.Order("matched_at DESC").Limit(5).Find(&collisionResults)

	fmt.Printf("\n🔍 最�?条碰撞匹配记�?\n")
	fmt.Printf("%3s | %6s | %6s | %-25s | %s\n", "ID", "UserID", "匹配用户", "匹配邮箱", "匹配时间")
	fmt.Println("--------------------------------------------------------------------")

	for _, result := range collisionResults {
		fmt.Printf("%3d | %6d | %6d | %-25s | %s\n", 
			result.ID, result.UserID, result.MatchedUserID, result.MatchedEmail, result.MatchedAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Println("--------------------------------------------------------------------")
	fmt.Printf("📊 共找�?%d 条碰撞匹配记录\n", len(collisionResults))

	fmt.Println("\n🎉 检查完成！")
}

// nSetEnv 设置环境变量
func nSetEnv(key, value string) {
	os.Setenv(key, value)
}
