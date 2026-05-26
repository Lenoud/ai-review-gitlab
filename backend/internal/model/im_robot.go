package model

import "time"

type IMRobot struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Platform   string    `gorm:"size:32;not null;index" json:"platform"`
	Name       string    `gorm:"size:128;not null;index" json:"name"`
	WebhookURL string    `gorm:"column:webhook_url;size:1024;not null" json:"webhookUrl"`
	Secret     string    `gorm:"size:512;not null;default:''" json:"secret"`
	Enabled    bool      `gorm:"not null;default:true;index" json:"enabled"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (IMRobot) TableName() string {
	return "im_robot"
}
