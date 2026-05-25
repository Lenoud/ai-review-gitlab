package model

import "time"

type LLMModel struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Provider   string    `gorm:"size:64;not null;index" json:"provider"`
	ModelCode  string    `gorm:"column:model_code;size:128;not null;index" json:"modelCode"`
	APIBaseURL string    `gorm:"column:api_base_url;size:512;not null" json:"apiBaseUrl"`
	APIKey     string    `gorm:"column:api_key;size:512;not null" json:"apiKey"`
	MaxTokens  int       `gorm:"column:max_tokens;not null;default:4096" json:"maxTokens"`
	IsDefault  bool      `gorm:"column:is_default;not null;default:false;index" json:"isDefault"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (LLMModel) TableName() string {
	return "llm_model"
}
