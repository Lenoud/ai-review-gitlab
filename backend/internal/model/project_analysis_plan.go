package model

import "time"

type ProjectAnalysisPlan struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	ProjectID         uint      `gorm:"column:project_id;not null;index" json:"projectId"`
	Name              string    `gorm:"size:128;not null" json:"name"`
	Prompt            string    `gorm:"type:text" json:"prompt"`
	CronExpression    string    `gorm:"column:cron_expression;size:128" json:"cronExpression"`
	Enabled           bool      `gorm:"not null;default:true" json:"enabled"`
	IMEnabled         bool      `gorm:"column:im_enabled;not null;default:false" json:"imEnabled"`
	IMRobotID         uint      `gorm:"column:im_robot_id" json:"imRobotId"`
	HTMLReportEnabled bool      `gorm:"column:html_report_enabled;not null;default:true" json:"htmlReportEnabled"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func (ProjectAnalysisPlan) TableName() string {
	return "project_analysis_plan"
}

type ProjectAnalysisPlanExecutionLog struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	PlanID              uint      `gorm:"column:plan_id;index" json:"planId"`
	ProjectID           uint      `gorm:"column:project_id;index" json:"projectId"`
	Status              string    `gorm:"size:32;not null;index" json:"status"`
	StartedAt           time.Time `gorm:"column:started_at" json:"startedAt"`
	CompletedAt         time.Time `gorm:"column:completed_at" json:"completedAt"`
	DurationMs          int64     `gorm:"column:duration_ms" json:"durationMs"`
	ResultContent       string    `gorm:"column:result_content;type:text" json:"resultContent"`
	ResultActions       string    `gorm:"column:result_actions;type:text" json:"resultActions"`
	ShareToken          string    `gorm:"column:share_token;size:128;index" json:"shareToken"`
	ShareTokenExpiresAt int64     `gorm:"column:share_token_expires_at" json:"shareTokenExpiresAt"`
	ErrorMessage        string    `gorm:"column:error_message;type:text" json:"errorMessage"`
	ErrorStack          string    `gorm:"column:error_stack;type:text" json:"errorStack"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

func (ProjectAnalysisPlanExecutionLog) TableName() string {
	return "project_analysis_plan_execution_log"
}
