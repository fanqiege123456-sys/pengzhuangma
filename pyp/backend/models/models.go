package models

import (
	"time"

	"gorm.io/gorm"
)

// 用户表
type User struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	OpenID        string         `gorm:"uniqueIndex;size:50" json:"openid"`
	UnionID       string         `gorm:"size:50" json:"unionid"`
	Nickname      string         `gorm:"size:100" json:"nickname"`
	Avatar        string         `gorm:"size:255" json:"avatar"`
	WechatNo      string         `gorm:"size:50" json:"wechat_no"`
	Phone         string         `gorm:"size:20" json:"phone"`
	Email         string         `gorm:"size:100" json:"email"`   // 邮箱
	Gender        int            `gorm:"default:0" json:"gender"` // 0:未知 1:男 2:女
	Age           int            `gorm:"default:0" json:"age"`    // 年龄
	Bio           string         `gorm:"size:500" json:"bio"`     // 个人简介
	Coins         int            `gorm:"default:0" json:"coins"`
	TotalRecharge int            `gorm:"default:0" json:"total_recharge"`

	// 地址信息
	Country  string `gorm:"size:50" json:"country"`  // 国家
	Province string `gorm:"size:50" json:"province"` // 省份
	City     string `gorm:"size:50" json:"city"`     // 城市
	District string `gorm:"size:50" json:"district"` // 区县

	// 地区可见性设置（用于碰撞匹配）
	// 如果 LocationVisible=true，则其他用户搜索该地区时可以匹配到我
	// 如果为 false，则即使地址匹配也不会被搜索到
	LocationVisible bool `gorm:"default:true" json:"location_visible"` // 地区是否可见（允许被搜索）

	// 碰撞设置
	AllowUpperLevel bool `gorm:"default:false" json:"allow_upper_level"` // 允许上级地区碰撞（已废弃，改为精准匹配）
	AllowPassiveAdd bool `gorm:"default:false" json:"allow_passive_add"` // 允许被动添加好友
	AllowForceAdd   bool `gorm:"default:false" json:"allow_force_add"`   // 允许被强制添加好友
	AllowHaidilao   bool `gorm:"default:false" json:"allow_haidilao"`    // 允许被海底捞
}

// 碰撞码表
type CollisionCode struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// 兴趣标签/碰撞码
	Tag string `gorm:"size:50;index" json:"tag"` // 兴趣标签（新字段）

	// 发布者地址信息（用于匹配）
	Country  string `gorm:"size:50;index" json:"country"`
	Province string `gorm:"size:50;index" json:"province"`
	City     string `gorm:"size:50;index" json:"city"`
	District string `gorm:"size:50;index" json:"district"`

	// 发布者性别
	Gender int `gorm:"index" json:"gender"` // 发布者性别

	// 年龄范围筛选
	AgeMin int `gorm:"default:20" json:"age_min"` // 最小年龄
	AgeMax int `gorm:"default:30" json:"age_max"` // 最大年龄

	// 碰撞设置
	Status    string    `gorm:"default:active" json:"status"` // active, expired
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`      // 24小时后过期
	CostCoins int       `gorm:"default:0" json:"cost_coins"`  // 发布消耗金币

	// 审核信息
	AuditStatus  string     `gorm:"default:pending" json:"audit_status"` // pending, approved, rejected
	AuditBy      uint       `json:"audit_by"`                            // 审核人ID
	AuditAt      *time.Time `json:"audit_at"`                            // 审核时间（指针类型，避免零值问题）
	RejectReason string     `gorm:"size:200" json:"reject_reason"`       // 拒绝原因

	// 统计信息
	MatchCount int  `gorm:"default:0" json:"match_count"`    // 匹配次数
	IsMatched  bool `gorm:"default:false" json:"is_matched"` // 是否已匹配

	// 管理端展示字段（不入库）
	IsForbidden bool `gorm:"-" json:"is_forbidden"` // 是否命中违禁词
}

// 碰撞记录表
type CollisionRecord struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 匹配双方
	UserID1 uint `gorm:"not null;index" json:"user_id1"`
	UserID2 uint `gorm:"not null;index" json:"user_id2"`
	User1   User `gorm:"foreignKey:UserID1" json:"user1,omitempty"`
	User2   User `gorm:"foreignKey:UserID2" json:"user2,omitempty"`

	// 匹配信息
	Tag       string `gorm:"size:50;not null" json:"tag"`        // 匹配的兴趣标签
	MatchType string `gorm:"size:20;not null" json:"match_type"` // district, city, province, country

	// 匹配时的地理信息
	MatchCountry  string `gorm:"size:50" json:"match_country"`
	MatchProvince string `gorm:"size:50" json:"match_province"`
	MatchCity     string `gorm:"size:50" json:"match_city"`
	MatchDistrict string `gorm:"size:50" json:"match_district"`

	// 状态管理
	Status            string    `gorm:"size:20;default:matched" json:"status"` // matched, friend_added, missed
	AddFriendDeadline time.Time `json:"add_friend_deadline"`                   // 加好友截止时间
	// 邮件发送状态
	EmailSent   bool       `gorm:"default:false" json:"email_sent"`        // 邮件是否已发送
	EmailSentAt *time.Time `json:"email_sent_at"`                          // 邮件发送时间
	EmailStatus string     `gorm:"size:20;default:''" json:"email_status"` // 邮件发送状态：success, failed, pending
}

// 好友表
type Friend struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	FriendID  uint           `gorm:"not null" json:"friend_id"`
	Status    string         `gorm:"size:20;default:pending" json:"status"` // pending, accepted, blocked
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Friend    User           `gorm:"foreignKey:FriendID" json:"friend,omitempty"`
}

// 好友条件表
type FriendCondition struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Gender    int            `json:"gender"`                   // 期望性别
	MinAge    int            `json:"min_age"`                  // 最小年龄
	MaxAge    int            `json:"max_age"`                  // 最大年龄
	Location  string         `gorm:"size:100" json:"location"` // 地区要求
}

// 充值记录表
type RechargeRecord struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Amount    int            `gorm:"not null" json:"amount"`                // 充值金额（分）
	Coins     int            `gorm:"not null" json:"coins"`                 // 获得金币数
	OrderNo   string         `gorm:"size:50;uniqueIndex" json:"order_no"`   // 订单号
	Status    string         `gorm:"size:20;default:pending" json:"status"` // pending, success, failed
	PayType   string         `gorm:"size:20" json:"pay_type"`               // wechat, alipay
}

// 消费记录表
type ConsumeRecord struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Coins     int            `gorm:"not null" json:"coins"`        // 消费金币数
	Type      string         `gorm:"size:20;not null" json:"type"` // collision, force_add, etc.
	Reason    string         `gorm:"size:100" json:"reason"`       // 消费原因描述
}

// 管理员表
type Admin struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Email     string         `gorm:"size:100" json:"email"`
	Nickname  string         `gorm:"size:100" json:"nickname"`
	Role      string         `gorm:"size:20;default:admin" json:"role"`
	Status    string         `gorm:"size:20;default:active" json:"status"`
}

// 用户地址表 - 支持多地址管理
type UserLocation struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uint `gorm:"not null;index" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	Label    string `gorm:"size:20" json:"label"`    // 地址标签: home(老家), school(学校), work(工作地), other(其他)
	Country  string `gorm:"size:50" json:"country"`  // 国家
	Province string `gorm:"size:50" json:"province"` // 省份
	City     string `gorm:"size:50" json:"city"`     // 城市
	District string `gorm:"size:50" json:"district"` // 区县

	IsDefault bool `gorm:"default:false" json:"is_default"` // 是否为默认地址
}
