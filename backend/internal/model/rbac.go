package model

import "time"

type SysRole struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Code        string    `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name        string    `gorm:"size:64;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (SysRole) TableName() string {
	return "sys_role"
}

type SysPermission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"size:128;uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Category  string    `gorm:"size:64" json:"category"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (SysPermission) TableName() string {
	return "sys_permission"
}

type SysUserRole struct {
	UserID    uint      `gorm:"primaryKey;autoIncrement:false" json:"userId"`
	RoleID    uint      `gorm:"primaryKey;autoIncrement:false" json:"roleId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (SysUserRole) TableName() string {
	return "sys_user_role"
}

type SysRolePermission struct {
	RoleID       uint      `gorm:"primaryKey;autoIncrement:false" json:"roleId"`
	PermissionID uint      `gorm:"primaryKey;autoIncrement:false" json:"permissionId"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (SysRolePermission) TableName() string {
	return "sys_role_permission"
}
