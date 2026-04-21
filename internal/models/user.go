package models

import "time"

// User 用户模型
type User struct {
	// 资源名称，格式: users/{id}
	Name string `json:"name,omitempty"`
	// 唯一标识符
	Uid string `json:"uid,omitempty"`
	// 创建时间
	CreateTime *time.Time `json:"createTime,omitempty"`
	// 更新时间
	UpdateTime *time.Time `json:"updateTime,omitempty"`
	// 用户名
	Username string `json:"username,omitempty"`
	// 邮箱
	Email string `json:"email,omitempty"`
	// 昵称
	Nickname string `json:"nickname,omitempty"`
	// 头像 URL
	AvatarUrl string `json:"avatarUrl,omitempty"`
	// 角色
	Role UserRole `json:"role,omitempty"`
	// 用户状态
	RowStatus RowStatus `json:"rowStatus,omitempty"`
}

// UserRole 用户角色
type UserRole string

const (
	UserRoleAdmin   UserRole = "ADMIN"
	UserRoleHost    UserRole = "HOST"
	UserRoleUser    UserRole = "USER"
	UserRoleUnknown UserRole = "ROLE_UNSPECIFIED"
)

// UserStats 用户统计
type UserStats struct {
	// Memo 数量
	MemoCount int32 `json:"memoCount,omitempty"`
	// 资源数量
	ResourceCount int32 `json:"resourceCount,omitempty"`
}

// GetUserStatsResponse 获取用户统计响应
type GetUserStatsResponse struct {
	Stats *UserStats `json:"stats,omitempty"`
}
