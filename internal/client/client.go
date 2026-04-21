// Package client provides HTTP client for Memos API
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cicbyte/memos-cli/internal/log"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// Client Memos API 客户端
type Client struct {
	client  *resty.Client
	baseURL string
	token   string
}

// Config 客户端配置
type Config struct {
	BaseURL string
	Token   string
	Timeout time.Duration
}

// NewClient 创建新的 API 客户端
func NewClient(cfg *Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	client := resty.New().
		SetBaseURL(cfg.BaseURL+"/api/v1").
		SetTimeout(cfg.Timeout).
		SetRetryCount(3).
		SetRetryWaitTime(500 * time.Millisecond).
		SetRetryMaxWaitTime(5 * time.Second).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	// 设置认证
	if cfg.Token != "" {
		client.SetAuthToken(cfg.Token)
	}

	// 添加请求日志
	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		log.Debug("API request",
			zap.String("method", req.Method),
			zap.String("url", req.URL),
		)
		return nil
	})

	// 添加响应日志
	client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		log.Debug("API response",
			zap.Int("status", resp.StatusCode()),
			zap.String("url", resp.Request.URL),
			zap.Duration("duration", resp.Time()),
		)
		return nil
	})

	return &Client{
		client:  client,
		baseURL: cfg.BaseURL,
		token:   cfg.Token,
	}
}

// SetToken 设置认证令牌
func (c *Client) SetToken(token string) {
	c.token = token
	c.client.SetAuthToken(token)
}

// GetToken 获取当前令牌
func (c *Client) GetToken() string {
	return c.token
}

// Get 执行 GET 请求
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(result).
		Get(path)

	return c.handleError(resp, err)
}

// GetWithQuery 执行带查询参数的 GET 请求
func (c *Client) GetWithQuery(ctx context.Context, path string, params map[string]string, result interface{}) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParams(params).
		SetResult(result).
		Get(path)

	return c.handleError(resp, err)
}

// Post 执行 POST 请求
func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(result).
		Post(path)

	return c.handleError(resp, err)
}

// Patch 执行 PATCH 请求
func (c *Client) Patch(ctx context.Context, path string, body interface{}, result interface{}) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(result).
		Patch(path)

	return c.handleError(resp, err)
}

// Delete 执行 DELETE 请求
func (c *Client) Delete(ctx context.Context, path string) error {
	resp, err := c.client.R().
		SetContext(ctx).
		Delete(path)

	return c.handleError(resp, err)
}

// handleError 统一处理 API 错误
func (c *Client) handleError(resp *resty.Response, err error) error {
	if err != nil {
		log.Error("API request failed", zap.Error(err))
		return fmt.Errorf("API request failed: %w", err)
	}

	if resp.IsError() {
		var apiErr APIError
		if jsonErr := json.Unmarshal(resp.Body(), &apiErr); jsonErr == nil && apiErr.Message != "" {
			log.Error("API error response",
				zap.Int("status", resp.StatusCode()),
				zap.String("code", apiErr.Code),
				zap.String("message", apiErr.Message),
			)
			return &apiErr
		}

		log.Error("HTTP error response",
			zap.Int("status", resp.StatusCode()),
			zap.String("body", string(resp.Body())),
		)
		return fmt.Errorf("HTTP error: %d - %s", resp.StatusCode(), resp.Status())
	}

	return nil
}

// APIError Memos API 错误响应
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details []struct {
		Type     string `json:"type"`
		Field    string `json:"field"`
		Description string `json:"description"`
	} `json:"details"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error [%s]: %s", e.Code, e.Message)
}

// IsNotFoundError 检查是否为未找到错误
func IsNotFoundError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == "NOT_FOUND" || apiErr.Code == "404"
	}
	return false
}

// IsUnauthorizedError 检查是否为未授权错误
func IsUnauthorizedError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == "UNAUTHORIZED" || apiErr.Code == "401"
	}
	return false
}

// IsNetworkError 检查是否为网络错误
func IsNetworkError(err error) bool {
	return err != nil && (err == http.ErrHandlerTimeout || err == context.DeadlineExceeded)
}
