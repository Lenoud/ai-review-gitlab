package model

import (
	"time"

	"gorm.io/datatypes"
)

type Project struct {
	ID                       uint           `gorm:"primaryKey" json:"id"`
	Name                     string         `gorm:"size:128;not null;index" json:"name"`
	Description              string         `gorm:"size:1024" json:"description"`
	WebURL                   string         `gorm:"column:web_url;size:512;uniqueIndex;not null" json:"webUrl"`
	Platform                 string         `gorm:"size:32;not null;default:gitlab;index" json:"platform"`
	AccessToken              string         `gorm:"size:512" json:"accessToken"`
	IMEnabled                bool           `gorm:"column:im_enabled;not null;default:false" json:"imEnabled"`
	IMRobotID                uint           `gorm:"column:im_robot_id" json:"imRobotId"`
	IMAtMemberEnabled        bool           `gorm:"column:im_at_member_enabled;not null;default:false" json:"imAtMemberEnabled"`
	IMAtMemberScoreThreshold int            `gorm:"column:im_at_member_score_threshold;not null;default:0" json:"imAtMemberScoreThreshold"`
	AIReviewEnabled          bool           `gorm:"column:ai_review_enabled;not null;default:true" json:"aiReviewEnabled"`
	TemplateID               uint           `gorm:"column:template_id" json:"templateId"`
	Extensions               datatypes.JSON `gorm:"type:json" json:"extensions"`
	ReviewEventTypes         datatypes.JSON `gorm:"column:review_event_types;type:json" json:"reviewEventTypes"`
	ReviewPromptTemplate     string         `gorm:"column:review_prompt_template;type:text" json:"reviewPromptTemplate"`
	HTMLReportEnabled        bool           `gorm:"column:html_report_enabled;not null;default:false" json:"htmlReportEnabled"`
	DeepReviewEnabled        bool           `gorm:"column:deep_review_enabled;not null;default:false" json:"deepReviewEnabled"`
	CreatedAt                time.Time      `json:"createdAt"`
	UpdatedAt                time.Time      `json:"updatedAt"`
}

func (Project) TableName() string {
	return "project"
}
