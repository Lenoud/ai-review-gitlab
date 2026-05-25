package model

import (
	"time"

	"gorm.io/datatypes"
)

type PushReviewLog struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	ProjectID           uint           `gorm:"column:project_id;not null;index" json:"projectId"`
	ProjectName         string         `gorm:"column:project_name;size:128;not null;index" json:"projectName"`
	Author              string         `gorm:"size:128;index" json:"author"`
	AuthorIdentity      string         `gorm:"column:author_identity;size:128;index" json:"authorIdentity"`
	AuthorDisplayName   string         `gorm:"column:author_display_name;size:128" json:"authorDisplayName"`
	Branch              string         `gorm:"size:255;index" json:"branch"`
	CommitMessages      string         `gorm:"column:commit_messages;type:text" json:"commitMessages"`
	Commits             datatypes.JSON `gorm:"type:json" json:"commits"`
	Score               int            `gorm:"not null;default:0;index" json:"score"`
	Additions           int            `gorm:"not null;default:0" json:"additions"`
	Deletions           int            `gorm:"not null;default:0" json:"deletions"`
	LastCommitURL       string         `gorm:"column:last_commit_url;size:1024" json:"lastCommitUrl"`
	ReviewResult        string         `gorm:"column:review_result;type:text" json:"reviewResult"`
	ShareToken          string         `gorm:"column:share_token;size:128;index" json:"shareToken"`
	ShareTokenExpiresAt int64          `gorm:"column:share_token_expires_at" json:"shareTokenExpiresAt"`
	CreatedAt           time.Time      `json:"createdAt"`
	UpdatedAt           time.Time      `json:"updatedAt"`
}

func (PushReviewLog) TableName() string {
	return "push_review_log"
}

type MergeRequestReviewLog struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	ProjectID           uint      `gorm:"column:project_id;not null;index" json:"projectId"`
	ProjectName         string    `gorm:"column:project_name;size:128;not null;index" json:"projectName"`
	Author              string    `gorm:"size:128;index" json:"author"`
	AuthorIdentity      string    `gorm:"column:author_identity;size:128;index" json:"authorIdentity"`
	AuthorDisplayName   string    `gorm:"column:author_display_name;size:128" json:"authorDisplayName"`
	SourceBranch        string    `gorm:"column:source_branch;size:255;index" json:"sourceBranch"`
	TargetBranch        string    `gorm:"column:target_branch;size:255;index" json:"targetBranch"`
	CommitMessages      string    `gorm:"column:commit_messages;type:text" json:"commitMessages"`
	Score               int       `gorm:"not null;default:0;index" json:"score"`
	Additions           int       `gorm:"not null;default:0" json:"additions"`
	Deletions           int       `gorm:"not null;default:0" json:"deletions"`
	LastCommitID        string    `gorm:"column:last_commit_id;size:128;index" json:"lastCommitId"`
	URL                 string    `gorm:"size:1024" json:"url"`
	ReviewResult        string    `gorm:"column:review_result;type:text" json:"reviewResult"`
	ShareToken          string    `gorm:"column:share_token;size:128;index" json:"shareToken"`
	ShareTokenExpiresAt int64     `gorm:"column:share_token_expires_at" json:"shareTokenExpiresAt"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

func (MergeRequestReviewLog) TableName() string {
	return "merge_request_review_log"
}
