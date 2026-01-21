package main

import (
	"log"
	"net/http"
	"time"

	"collision-backend/config"
	"collision-backend/middlewares"
	"collision-backend/models"
	"collision-backend/routes"
	"collision-backend/services"
	"collision-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 初始化配置
	config.Init()

	// 初始化JWT
	utils.InitJWT(config.Config.JWTSecret)

	// 数据库迁移 - 包含所有模型，GORM 会自动添加缺失的字段
	err := config.DB.AutoMigrate(
		// 基础模型
		&models.User{},
		&models.CollisionCode{},
		&models.HotTag{},
		&models.CollisionRecord{},
		&models.Friend{},
		&models.FriendCondition{},
		&models.RechargeRecord{},
		&models.ConsumeRecord{},
		&models.Admin{},
		&models.UserLocation{},
		// V3.0 新增模型
		&models.CollisionList{},
		&models.CollisionResult{},
		&models.UserContact{},
		&models.HotTag{}, // 确保HotTag模型被迁移
		&models.UserContact{},
		&models.HotTag{},
		&models.EmailLog{},
		&models.SystemConfig{},
		&models.ForbiddenKeyword{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 创建默认管理员
	createDefaultAdmin()

	// 启动后台服务（清理服务、匹配服务）
	go startBackgroundServices()

	// 设置Gin模式
	gin.SetMode(gin.DebugMode)

	// 创建路由
	r := gin.Default()

	// 全局应用CORS中间件（在所有路由之前）
	r.Use(middlewares.CORS())

	// 设置路由
	routes.SetupRoutes(r)
	routes.RegisterV3Routes(r)

	// 启动WebSocket服务器，监听8001端口
	go func() {
		// WebSocket升级器
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有跨域请求
			},
		}

		// WebSocket处理函数
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// 升级HTTP连接为WebSocket
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Printf("WebSocket升级失败: %v", err)
				return
			}
			defer conn.Close()

			log.Printf("新的WebSocket连接")

			// 保持连接，接收消息
			for {
				// 读取消息
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Printf("WebSocket读取失败: %v", err)
					break
				}

				// 打印收到的消息
				log.Printf("收到WebSocket消息: %s", message)

				// 回复消息
				err = conn.WriteMessage(websocket.TextMessage, []byte("连接成功"))
				if err != nil {
					log.Printf("WebSocket写入失败: %v", err)
					break
				}
			}
		})

		// 启动WebSocket服务器
		log.Printf("WebSocket服务器启动在8001端口")
		if err := http.ListenAndServe("0.0.0.0:8001", nil); err != nil {
			log.Fatal("WebSocket服务器启动失败:", err)
		}
	}()

	// 启动HTTP服务器，绑定到0.0.0.0确保手机能访问
	log.Printf("HTTP服务器启动在端口 %s", config.Config.ServerPort)
	if err := r.Run("0.0.0.0:" + config.Config.ServerPort); err != nil {
		log.Fatal("HTTP服务器启动失败:", err)
	}
}

func createDefaultAdmin() {
	var count int64
	config.DB.Model(&models.Admin{}).Count(&count)

	if count == 0 {
		// 创建默认管理员
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		admin := models.Admin{
			Username: "admin",
			Password: string(hashedPassword),
			Email:    "admin@example.com",
			Role:     "admin",
			Status:   "active",
		}

		config.DB.Create(&admin)
		log.Println("Default admin created - username: admin, password: admin123")
	}

	// 启动清理服务
	cleanupService := &services.CleanupService{}
	cleanupService.StartCleanupTasks()
}

// startBackgroundServices 启动后台服务
func startBackgroundServices() {
	log.Println("启动后台服务...")

	// 1. 启动清理服务
	cleanupService := &services.CleanupService{}
	go cleanupService.StartCleanupTasks()

	// 2. 启动碰撞匹配服务（每5分钟执行一次）
	// V3.0 已集成邮件通知功能：由用户手动选择发送邮件
	collisionMatcher := services.NewCollisionMatcher()
	go collisionMatcher.StartMatcherService(5 * time.Minute)

	log.Println("后台服务启动完成")
}
