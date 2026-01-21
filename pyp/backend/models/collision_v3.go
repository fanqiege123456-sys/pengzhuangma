package models

import (
	"encoding/json"
	"time"
)

// CollisionList 碰撞列表
type CollisionList struct {
	ID         uint64    `json:"id" gorm:"primaryKey"`
	UserID     uint64    `json:"user_id" gorm:"index;not null"`
	Keyword    string    `json:"keyword" gorm:"size:100;not null"`
	Duration   int       `json:"duration" gorm:"default:30"`
	CostPoints int       `json:"cost_points" gorm:"default:30"`
	Status     string    `json:"status" gorm:"size:20;default:active"` // active, inactive, expired
	ExpireAt   time.Time `json:"expire_at" gorm:"index"`
	MatchCount int       `json:"match_count" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (CollisionList) TableName() string {
	return "collision_lists"
}

// CollisionResult 碰撞结果
type CollisionResult struct {
	ID              uint64     `json:"id" gorm:"primaryKey"`
	UserID          uint64     `json:"user_id" gorm:"index;not null"`
	MatchedUserID   uint64     `json:"matched_user_id" gorm:"index;not null"`
	CollisionListID uint64     `json:"collision_list_id" gorm:"index;not null"`
	Keyword         string     `json:"keyword" gorm:"size:100;not null"`
	MatchedEmail    string     `json:"matched_email" gorm:"size:255"`
	Remark          string     `json:"remark" gorm:"size:20;default:''"`
	IsKnown         bool       `json:"is_known" gorm:"default:false"`
	EmailSent       bool       `json:"email_sent" gorm:"default:false"`
	EmailSentAt     *time.Time `json:"email_sent_at"`
	MatchedAt       time.Time  `json:"matched_at" gorm:"index"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (CollisionResult) TableName() string {
	return "collision_results"
}

// UserContact 用户联系方式
type UserContact struct {
	ID                uint64     `json:"id" gorm:"primaryKey"`
	UserID            uint64     `json:"user_id" gorm:"uniqueIndex;not null"`
	Email             string     `json:"email" gorm:"size:191;index"`
	EmailVerified     bool       `json:"email_verified" gorm:"default:false"`
	EmailVisible      bool       `json:"email_visible" gorm:"default:true"` // 邮箱是否在碰撞结果中显示
	EmailVerifyCode   string     `json:"-" gorm:"size:10"`
	EmailVerifyExpire *time.Time `json:"-"`
	Phone             string     `json:"phone" gorm:"size:20;index"`
	PhoneVerified     bool       `json:"phone_verified" gorm:"default:false"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (UserContact) TableName() string {
	return "user_contacts"
}

// HotTag 热门标签
type HotTag struct {
	ID           uint64     `json:"id" gorm:"primaryKey"`
	Keyword      string     `json:"keyword" gorm:"size:100;uniqueIndex;not null"`
	Count24h     int        `json:"count_24h" gorm:"column:count_24h;default:0;index"`
	CountTotal   int        `json:"count_total" gorm:"column:count_total;default:0;index"`
	Status       string     `json:"status" gorm:"size:20;default:show"` // show, hide, blackhole
	SubmitCount  int        `json:"submit_count" gorm:"default:0"`
	LastSearchAt *time.Time `json:"last_search_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (HotTag) TableName() string {
	return "hot_tags"
}

// EmailLog 邮件发送记录
type EmailLog struct {
	ID        uint64     `json:"id" gorm:"primaryKey"`
	UserID    uint64     `json:"user_id" gorm:"index;not null"`
	ToEmail   string     `json:"to_email" gorm:"size:255;not null"`
	Subject   string     `json:"subject" gorm:"size:255;not null"`
	Content   string     `json:"content" gorm:"type:text"`
	Type      string     `json:"type" gorm:"size:20;default:system;index"`    // verify, collision, system
	Status    string     `json:"status" gorm:"size:20;default:pending;index"` // pending, sent, failed
	ErrorMsg  string     `json:"error_msg" gorm:"type:text"`
	SentAt    *time.Time `json:"sent_at"`
	CreatedAt time.Time  `json:"created_at"`
}

func (EmailLog) TableName() string {
	return "email_logs"
}

// CollisionResultGroup 按日期分组的碰撞结果
type CollisionResultGroup struct {
	ID         string            `json:"id"`
	Date       string            `json:"date"`
	Keyword    string            `json:"keyword"`
	Total      int               `json:"total"`
	KnownCount int               `json:"knownCount"`
	Matches    []CollisionResult `json:"matches"`
}

// SystemConfig 系统配置表
type SystemConfig struct {
	ID          uint64    `json:"id" gorm:"primaryKey"`
	ConfigKey   string    `json:"config_key" gorm:"uniqueIndex;size:100;not null"`
	ConfigValue string    `json:"config_value" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

// GetValue 从 JSON 配置中获取值
func (c *SystemConfig) GetValue(key string) string {
	if c.ConfigValue == "" {
		return ""
	}
	// 简单 JSON 解析
	var data map[string]string
	if err := json.Unmarshal([]byte(c.ConfigValue), &data); err != nil {
		return ""
	}
	return data[key]
}

// SetValues 设置配置值
func (c *SystemConfig) SetValues(values map[string]string) {
	data, _ := json.Marshal(values)
	c.ConfigValue = string(data)
}
