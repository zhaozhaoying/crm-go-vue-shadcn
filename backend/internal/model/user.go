package model

import "time"

const (
	UserStatusEnabled  = "enabled"
	UserStatusDisabled = "disabled"
)

type User struct {
	ID        int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Username  string    `json:"username" gorm:"column:username;size:64;not null;uniqueIndex"`
	Password  string    `json:"-" gorm:"column:password;not null"`
	Salt      string    `json:"-" gorm:"column:salt;size:128;not null;default:''"`
	Nickname  string    `json:"nickname" gorm:"column:nickname;size:128;not null;default:''"`
	Email     string    `json:"email" gorm:"column:email;size:128;not null;default:''"`
	Mobile    string    `json:"mobile" gorm:"column:mobile;size:32;not null;default:''"`
	Avatar    string    `json:"avatar" gorm:"column:avatar;size:255;not null;default:''"`
	RoleID    int64     `json:"roleId" gorm:"column:role_id;not null;default:0;index"`
	ParentID  *int64    `json:"parentId" gorm:"column:parent_id"`
	Status    string    `json:"status" gorm:"column:status;size:32;not null;default:'enabled';index"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`

	Role *Role `json:"-" gorm:"foreignKey:RoleID;references:ID"`
}

// UserWithRole 用于列表展示，带角色名
type UserWithRole struct {
	User
	RoleName  string `json:"roleName" gorm:"column:role_name;->"`
	RoleLabel string `json:"roleLabel" gorm:"column:role_label;->"`
}

func (User) TableName() string { return "users" }
