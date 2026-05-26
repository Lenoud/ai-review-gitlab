package model

import "time"

type MemberIMMapping struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	GitUsername string    `gorm:"column:git_username;size:128;not null;index:uk_member_im_mapping_git_platform,unique;index" json:"gitUsername"`
	Platform    string    `gorm:"size:32;not null;index:uk_member_im_mapping_git_platform,unique;index" json:"platform"`
	IMUserID    string    `gorm:"column:im_user_id;size:256;not null" json:"imUserId"`
	DisplayName string    `gorm:"size:128;not null;default:'';index" json:"displayName"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (MemberIMMapping) TableName() string {
	return "member_im_mapping"
}
