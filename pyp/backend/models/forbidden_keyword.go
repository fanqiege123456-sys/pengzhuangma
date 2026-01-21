package models

import "time"

// ForbiddenKeyword 违禁词
type ForbiddenKeyword struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	Keyword   string    `json:"keyword" gorm:"size:100;uniqueIndex;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ForbiddenKeyword) TableName() string {
	return "forbidden_keywords"
}
