package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB     *gorm.DB
	Redis  *redis.Client
	Config AppConfig
)

type AppConfig struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	RedisHost    string
	RedisPort    string
	RedisDB      int
	JWTSecret    string
	ServerPort   string
	WechatAppID  string
	WechatSecret string
	// 阿里云邮件服务配置
	AliyunDMAccessKey    string
	AliyunDMAccessSecret string
	AliyunDMAccount      string
	AliyunDMAccountName  string
	AliyunDMRegion       string
	// SMTP邮件配置（用于阿里企业邮箱）
	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  string
	SMTPFromAlias string
	SMTPReplyTo   string
	// 审核配置
	EnableCollisionAudit bool // 是否开启碰撞码审核
}

func GetConfig() *AppConfig {
	return &Config
}

func Init() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 初始化配置
	Config = AppConfig{
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "3306"),
		DBUser:       getEnv("DB_USER", "root"),
		DBPassword:   getEnv("DB_PASSWORD", ""),
		DBName:       getEnv("DB_NAME", "collision_db"),
		RedisHost:    getEnv("REDIS_HOST", "localhost"),
		RedisPort:    getEnv("REDIS_PORT", "6379"),
		JWTSecret:    getEnv("JWT_SECRET", "default_secret"),
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		WechatAppID:  getEnv("WECHAT_APPID", ""),
		WechatSecret: getEnv("WECHAT_SECRET", ""),
		// 阿里云邮件配置
		AliyunDMAccessKey:    getEnv("ALIYUN_DM_ACCESS_KEY", ""),
		AliyunDMAccessSecret: getEnv("ALIYUN_DM_ACCESS_SECRET", ""),
		AliyunDMAccount:      getEnv("ALIYUN_DM_ACCOUNT", "noreply@yourdomain.com"),
		AliyunDMAccountName:  getEnv("ALIYUN_DM_ACCOUNT_NAME", "标签碰撞"),
		AliyunDMRegion:       getEnv("ALIYUN_DM_REGION", "cn-hangzhou"),
		// SMTP邮件配置
		SMTPHost:      getEnv("SMTP_HOST", "smtp.qiye.aliyun.com"),
		SMTPPort:      getEnvInt("SMTP_PORT", 25),
		SMTPUsername:  getEnv("SMTP_USERNAME", ""),
		SMTPPassword:  getEnv("SMTP_PASSWORD", ""),
		SMTPFromAlias: getEnv("SMTP_FROM_ALIAS", "标签碰撞"),
		SMTPReplyTo:   getEnv("SMTP_REPLY_TO", ""),
		// 审核配置，默认关闭审核
		EnableCollisionAudit: getEnvBool("ENABLE_COLLISION_AUDIT", false),
	}

	// 初始化数据库
	initDB()

	// 初始化Redis
	initRedis()
}

func initDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s",
		Config.DBUser,
		Config.DBPassword,
		Config.DBHost,
		Config.DBPort,
		Config.DBName,
	)

	var err error
	// 配置GORM
	gormConfig := &gorm.Config{}

	// 打开数据库连接
	DB, err = gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get database connection pool:", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)                  // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)                 // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour)        // 连接最大生命周期
	sqlDB.SetConnMaxIdleTime(30 * time.Minute) // 连接最大空闲时间

	log.Println("Database connected successfully with connection pool configured")
}

func initRedis() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     Config.RedisHost + ":" + Config.RedisPort,
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})

	log.Println("Redis connected successfully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intVal
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolVal
}
