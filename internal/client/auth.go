package client

import (
	"context"

	"github.com/cicbyte/memos-cli/internal/models"
)

// AuthService 认证服务
type AuthService struct {
	client *Client
}

// NewAuthService 创建认证服务
func NewAuthService(client *Client) *AuthService {
	return &AuthService{client: client}
}

// SignIn 登录
func (s *AuthService) SignIn(ctx context.Context, req *models.SignInRequest) (*models.SignInResponse, error) {
	var result struct {
		User        *models.User `json:"user"`
		AccessToken string       `json:"accessToken"`
		ExpiresAt   string       `json:"expiresAt"`
	}

	// 使用 form 数据登录
	resp, err := s.client.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&result).
		Post("/auth/signin")

	if err := s.client.handleError(resp, err); err != nil {
		return nil, err
	}

	// 保存 token
	if result.AccessToken != "" {
		s.client.SetToken(result.AccessToken)
	}

	return &models.SignInResponse{
		User:        result.User,
		AccessToken: result.AccessToken,
	}, nil
}

// SignOut 登出
func (s *AuthService) SignOut(ctx context.Context) error {
	resp, err := s.client.client.R().
		SetContext(ctx).
		Post("/auth/signout")

	if err := s.client.handleError(resp, err); err != nil {
		return err
	}

	// 清除 token
	s.client.SetToken("")

	return nil
}

// GetCurrentUser 获取当前用户信息
// 注意: Memos API 没有 /auth/me 端点
// 我们通过获取 memos 列表来验证 token 是否有效
func (s *AuthService) GetCurrentUser(ctx context.Context) (*models.User, error) {
	// 尝试获取 memos 列表来验证 token
	var result struct {
		Memos []*models.Memo `json:"memos"`
	}

	err := s.client.Get(ctx, "/memos?pageSize=1", &result)
	if err != nil {
		return nil, err
	}

	// 如果能获取到 memos，说明 token 有效
	// 从第一个 memo 中获取 creator 信息
	if len(result.Memos) > 0 && result.Memos[0].Creator != "" {
		// 获取用户信息
		var user models.User
		err = s.client.Get(ctx, "/"+result.Memos[0].Creator, &user)
		if err == nil {
			return &user, nil
		}
	}

	// 返回一个基本用户信息
	return &models.User{
		Name:     "users/1",
		Username: "user",
		Role:     models.UserRoleUser,
	}, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(ctx context.Context) (*models.RefreshTokenResponse, error) {
	var result models.RefreshTokenResponse

	err := s.client.Post(ctx, "/auth/refresh", nil, &result)
	if err != nil {
		return nil, err
	}

	// 更新 token
	if result.AccessToken != "" {
		s.client.SetToken(result.AccessToken)
	}

	return &result, nil
}

// IsAuthenticated 检查是否已认证
func (s *AuthService) IsAuthenticated() bool {
	return s.client.GetToken() != ""
}
