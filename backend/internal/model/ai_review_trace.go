package model

import "time"

type AIReviewTrace struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ReviewEventType string    `gorm:"column:review_event_type;size:32;not null;uniqueIndex:uk_ai_review_trace_event;index" json:"reviewEventType"`
	ReviewEventID   uint      `gorm:"column:review_event_id;not null;uniqueIndex:uk_ai_review_trace_event;index" json:"reviewEventId"`
	Prompt          string    `gorm:"type:longtext" json:"prompt"`
	Response        string    `gorm:"type:longtext" json:"response"`
	Provider        string    `gorm:"size:64;not null;index" json:"provider"`
	ModelCode       string    `gorm:"column:model_code;size:128;not null;index" json:"modelCode"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (AIReviewTrace) TableName() string {
	return "ai_review_trace"
}
