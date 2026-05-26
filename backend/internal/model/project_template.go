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
