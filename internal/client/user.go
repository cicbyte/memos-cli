package client

import (
	"context"
	"fmt"

	"github.com/cicbyte/memos-cli/internal/models"
)

// UserService 用户服务
type UserService struct {
	client *Client
}

// NewUserService 创建用户服务
func NewUserService(client *Client) *UserService {
	return &UserService{client: client}
}

// List 列出所有用户
func (s *UserService) List(ctx context.Context) ([]*models.User, error) {
	var result struct {
		Users []*models.User `json:"users"`
	}

	err := s.client.Get(ctx, "/users", &result)
	if err != nil {
		return nil, err
	}

	return result.Users, nil
}

// Get 获取单个用户
func (s *UserService) Get(ctx context.Context, userID string) (*models.User, error) {
	var result models.User
	err := s.client.Get(ctx, fmt.Sprintf("/users/%s", userID), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetStats 获取用户统计
func (s *UserService) GetStats(ctx context.Context, userID string) (*models.UserStats, error) {
	var result models.GetUserStatsResponse
	err := s.client.Get(ctx, fmt.Sprintf("/users/%s:getStats", userID), &result)
	if err != nil {
		return nil, err
	}
	return result.Stats, nil
}

// GetAllStats 获取所有用户统计
func (s *UserService) GetAllStats(ctx context.Context) (map[string]*models.UserStats, error) {
	var result struct {
		UserStats map[string]*models.UserStats `json:"userStats"`
	}

	err := s.client.Get(ctx, "/users:stats", &result)
	if err != nil {
		return nil, err
	}

	return result.UserStats, nil
}
