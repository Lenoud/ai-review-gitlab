package model

import (
	"time"

	"gorm.io/datatypes"
)

type ProjectTemplate struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	Name                 string         `gorm:"size:128;not null;index" json:"name"`
	Description          string         `gorm:"size:1024;not null;default:''" json:"description"`
	Extensions           datatypes.JSON `gorm:"type:json" json:"extensions"`
	ReviewPromptTemplate string         `gorm:"column:review_prompt_template;type:text" json:"reviewPromptTemplate"`
	CreatedAt            time.Time      `json:"createdAt"`
	UpdatedAt            time.Time      `json:"updatedAt"`
}

func (ProjectTemplate) TableName() string {
	return "project_template"
}

type ProjectTemplateReviewRule struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	TemplateID   uint           `gorm:"column:template_id;not null;index" json:"templateId"`
	Name         string         `gorm:"size:128;not null" json:"name"`
	Description  string         `gorm:"size:512;not null;default:''" json:"description"`
	GlobPatterns datatypes.JSON `gorm:"column:glob_patterns;type:json" json:"globPatterns"`
	Content      string         `gorm:"type:text;not null" json:"content"`
	Priority     int            `gorm:"not null;default:0;index" json:"priority"`
	Enabled      bool           `gorm:"not null;default:true;index" json:"enabled"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

func (ProjectTemplateReviewRule) TableName() string {
	return "project_template_review_rule"
}
