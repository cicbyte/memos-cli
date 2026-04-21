package models

import "time"

// SignInRequest 登录请求
type SignInRequest struct {
	// 用户名
	Username string `json:"username"`
	// 密码
	Password string `json:"password"`
	// 记住登录
	Remember bool `json:"remember,omitempty"`
}

// SignInResponse 登录响应
type SignInResponse struct {
	// 用户信息
	User *User `json:"user,omitempty"`
	// 访问令牌
	AccessToken string `json:"accessToken,omitempty"`
	// 过期时间
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// GetCurrentUserResponse 获取当前用户响应
type GetCurrentUserResponse struct {
	// 用户信息
	User *User `json:"user,omitempty"`
}

// RefreshTokenResponse 刷新令牌响应
type RefreshTokenResponse struct {
	// 新访问令牌
	AccessToken string `json:"accessToken,omitempty"`
	// 过期时间
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// PersonalAccessToken 个人访问令牌
type PersonalAccessToken struct {
	// 资源名称
	Name string `json:"name,omitempty"`
	// 唯一标识符
	Uid string `json:"uid,omitempty"`
	// 创建时间
	CreateTime *time.Time `json:"createTime,omitempty"`
	// 描述
	Description string `json:"description,omitempty"`
	// 过期时间
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	// 是否已使用
	Used bool `json:"used,omitempty"`
	// 访问令牌 (仅在创建时返回)
	AccessToken string `json:"accessToken,omitempty"`
}

// CreatePersonalAccessTokenRequest 创建个人访问令牌请求
type CreatePersonalAccessTokenRequest struct {
	// 描述
	Description string `json:"description,omitempty"`
	// 过期时间
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}
