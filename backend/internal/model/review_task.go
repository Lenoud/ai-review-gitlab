package model

import (
	"time"

	"gorm.io/datatypes"
)

type ReviewTask struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	ProjectID     uint           `gorm:"column:project_id;not null;index" json:"projectId"`
	EventType     string         `gorm:"column:event_type;size:32;not null;index" json:"eventType"`
	DedupeKey     string         `gorm:"column:dedupe_key;size:255;not null;uniqueIndex" json:"dedupeKey"`
	PayloadJSON   datatypes.JSON `gorm:"column:payload_json;type:json" json:"payloadJson"`
	Status        string         `gorm:"size:32;not null;index" json:"status"`
	Priority      int            `gorm:"not null;default:0;index" json:"priority"`
	Attempts      int            `gorm:"not null;default:0" json:"attempts"`
	MaxAttempts   int            `gorm:"column:max_attempts;not null;default:3" json:"maxAttempts"`
	NextRunAt     time.Time      `gorm:"column:next_run_at;not null;index" json:"nextRunAt"`
	LockedBy      string         `gorm:"column:locked_by;size:128" json:"lockedBy"`
	LockedAt      *time.Time     `gorm:"column:locked_at" json:"lockedAt"`
	StartedAt     *time.Time     `gorm:"column:started_at" json:"startedAt"`
	FinishedAt    *time.Time     `gorm:"column:finished_at" json:"finishedAt"`
	ErrorMessage  string         `gorm:"column:error_message;size:1024" json:"errorMessage"`
	ResultLogType string         `gorm:"column:result_log_type;size:32" json:"resultLogType"`
	ResultLogID   uint           `gorm:"column:result_log_id" json:"resultLogId"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

func (ReviewTask) TableName() string {
	return "review_task"
}

type ReviewTaskAttempt struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	TaskID       uint       `gorm:"column:task_id;not null;index" json:"taskId"`
	AttemptNo    int        `gorm:"column:attempt_no;not null" json:"attemptNo"`
	Status       string     `gorm:"size:32;not null;index" json:"status"`
	StartedAt    time.Time  `gorm:"column:started_at;not null" json:"startedAt"`
	FinishedAt   *time.Time `gorm:"column:finished_at" json:"finishedAt"`
	DurationMs   int64      `gorm:"column:duration_ms;not null;default:0" json:"durationMs"`
	ErrorMessage string     `gorm:"column:error_message;size:1024" json:"errorMessage"`
	ErrorStack   string     `gorm:"column:error_stack;type:text" json:"errorStack"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (ReviewTaskAttempt) TableName() string {
	return "review_task_attempt"
}
