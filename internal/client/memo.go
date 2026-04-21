package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cicbyte/memos-cli/internal/models"
)

// MemoService 备忘录服务
type MemoService struct {
	client *Client
}

// NewMemoService 创建备忘录服务
func NewMemoService(client *Client) *MemoService {
	return &MemoService{client: client}
}

// ListOptions 列表查询选项
type ListOptions struct {
	// 页面大小
	PageSize int32
	// 页面令牌
	PageToken string
	// 过滤条件 (Google AIP-160 格式)
	Filter string
	// 父资源 (通常是 users/{id})
	Parent string
	// 标签过滤
	Tag string
	// 可见性过滤
	Visibility models.Visibility
	// 状态过滤
	RowStatus models.RowStatus
}

// List 列出备忘录
func (s *MemoService) List(ctx context.Context, opts *ListOptions) (*models.ListMemosResponse, error) {
	if opts == nil {
		opts = &ListOptions{}
	}

	// 构建过滤条件 - 注意：Memos API 的 filter 语法可能有版本差异
	// 暂时只使用基本的分页参数
	params := make(map[string]string)
	if opts.PageSize > 0 {
		params["pageSize"] = strconv.FormatInt(int64(opts.PageSize), 10)
	}
	if opts.PageToken != "" {
		params["pageToken"] = opts.PageToken
	}
	// 暂时禁用 filter，因为不同版本的 Memos API filter 语法可能不同
	// 如果需要过滤，可以在获取数据后在客户端进行过滤

	var result models.ListMemosResponse
	err := s.client.GetWithQuery(ctx, "/memos", params, &result)
	if err != nil {
		return nil, err
	}

	// 客户端过滤
	if opts.RowStatus != "" || opts.Visibility != "" || opts.Tag != "" {
		var filtered []*models.Memo
		for _, m := range result.Memos {
			// 检查状态
			if opts.RowStatus != "" {
				expectedState := "NORMAL"
				if opts.RowStatus == models.RowStatusArchived {
					expectedState = "ARCHIVED"
				}
				if string(m.RowStatus) != expectedState && m.RowStatus != models.RowStatus(opts.RowStatus) {
					continue
				}
			}
			// 检查可见性
			if opts.Visibility != "" && m.Visibility != opts.Visibility {
				continue
			}
			// 检查标签
			if opts.Tag != "" {
				hasTag := false
				if m.Property != nil {
					for _, t := range m.Property.Tags {
						if t == opts.Tag {
							hasTag = true
							break
						}
					}
				}
				if !hasTag {
					continue
				}
			}
			filtered = append(filtered, m)
		}
		result.Memos = filtered
	}

	return &result, nil
}

// Get 获取单个备忘录
func (s *MemoService) Get(ctx context.Context, memoID string) (*models.Memo, error) {
	var result models.Memo
	err := s.client.Get(ctx, fmt.Sprintf("/memos/%s", memoID), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create 创建备忘录
func (s *MemoService) Create(ctx context.Context, req *models.CreateMemoRequest) (*models.Memo, error) {
	var result models.Memo
	err := s.client.Post(ctx, "/memos", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update 更新备忘录
func (s *MemoService) Update(ctx context.Context, memoID string, req *models.UpdateMemoRequest) (*models.Memo, error) {
	var result models.Memo
	err := s.client.Patch(ctx, fmt.Sprintf("/memos/%s", memoID), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete 删除备忘录
func (s *MemoService) Delete(ctx context.Context, memoID string) error {
	return s.client.Delete(ctx, fmt.Sprintf("/memos/%s", memoID))
}

// GetComments 获取备忘录评论
func (s *MemoService) GetComments(ctx context.Context, memoID string) ([]*models.Memo, error) {
	var result struct {
		Comments []*models.Memo `json:"comments"`
	}
	err := s.client.Get(ctx, fmt.Sprintf("/memos/%s/comments", memoID), &result)
	if err != nil {
		return nil, err
	}
	return result.Comments, nil
}

// CreateComment 创建评论
func (s *MemoService) CreateComment(ctx context.Context, memoID string, req *models.CreateMemoRequest) (*models.Memo, error) {
	// 设置父备忘录
	parent := fmt.Sprintf("memos/%s", memoID)
	req.Parent = &parent

	var result models.Memo
	err := s.client.Post(ctx, fmt.Sprintf("/memos/%s/comments", memoID), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetReactions 获取备忘录反应
func (s *MemoService) GetReactions(ctx context.Context, memoID string) ([]*models.Reaction, error) {
	var result struct {
		Reactions []*models.Reaction `json:"reactions"`
	}
	err := s.client.Get(ctx, fmt.Sprintf("/memos/%s/reactions", memoID), &result)
	if err != nil {
		return nil, err
	}
	return result.Reactions, nil
}

// CreateReaction 创建反应
func (s *MemoService) CreateReaction(ctx context.Context, memoID string, reactionType string) (*models.Reaction, error) {
	req := struct {
		ReactionType string `json:"reactionType"`
	}{
		ReactionType: reactionType,
	}

	var result models.Reaction
	err := s.client.Post(ctx, fmt.Sprintf("/memos/%s/reactions", memoID), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteReaction 删除反应
func (s *MemoService) DeleteReaction(ctx context.Context, memoID string, reactionID string) error {
	return s.client.Delete(ctx, fmt.Sprintf("/memos/%s/reactions/%s", memoID, reactionID))
}

// Search 搜索备忘录
func (s *MemoService) Search(ctx context.Context, query string, opts *ListOptions) (*models.ListMemosResponse, error) {
	if opts == nil {
		opts = &ListOptions{}
	}

	// 添加搜索过滤条件
	searchFilter := fmt.Sprintf(`content.search("%s")`, query)
	if opts.Filter != "" {
		opts.Filter = fmt.Sprintf("%s && %s", opts.Filter, searchFilter)
	} else {
		opts.Filter = searchFilter
	}

	return s.List(ctx, opts)
}
