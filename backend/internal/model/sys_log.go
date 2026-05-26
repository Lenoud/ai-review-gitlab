package model

import "time"

type SysLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Level      string    `gorm:"size:32;index" json:"level"`
	Module     string    `gorm:"size:64;index" json:"module"`
	Action     string    `gorm:"size:128;index" json:"action"`
	Message    string    `gorm:"size:255" json:"message"`
	Detail     string    `gorm:"type:text" json:"detail"`
	ErrorStack string    `gorm:"column:error_stack;type:mediumtext" json:"errorStack"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (SysLog) TableName() string {
	return "sys_log"
}
