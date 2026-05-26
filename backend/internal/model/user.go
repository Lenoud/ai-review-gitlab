package model

import "time"

type SysUser struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Nickname     string    `gorm:"size:64" json:"nickname"`
	Remark       string    `gorm:"size:255" json:"remark"`
	Email        string    `gorm:"size:128" json:"email"`
	Avatar       string    `gorm:"size:512" json:"avatar"`
	Status       string    `gorm:"size:32;not null;default:enabled" json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (SysUser) TableName() string {
	return "sys_user"
}
