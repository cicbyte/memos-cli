package memologic

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cicbyte/memos-cli/internal/client"
	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/cicbyte/memos-cli/internal/utils"
	"gorm.io/gorm"
)

// List

type ListConfig struct {
	Limit      int32
	Visibility string
	Tag        string
	Archived   bool
	Page       string
	Search     string
}

type ListResult struct {
	Memos         []models.LocalMemo
	TotalCount    int64
	FilteredCount int64
	PageSize      int32
}

type ListProcessor struct {
	config    *ListConfig
	appConfig *models.AppConfig
}

func NewListProcessor(config *ListConfig, appConfig *models.AppConfig) *ListProcessor {
	return &ListProcessor{config: config, appConfig: appConfig}
}

func (p *ListProcessor) Execute() (*ListResult, error) {
	db, err := utils.GetGormDB()
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	var totalCount int64
	db.Model(&models.LocalMemo{}).Where("is_deleted = ?", false).Count(&totalCount)

	query := db.Model(&models.LocalMemo{}).Where("is_deleted = ?", false)

	if p.config.Visibility != "" {
		query = query.Where("visibility = ?", p.config.Visibility)
	}
	if p.config.Tag != "" {
		query = query.Where("property LIKE ?", fmt.Sprintf("%%\"%s\"%%", p.config.Tag))
	}
	if p.config.Archived {
		query = query.Where("row_status = ?", "ARCHIVED")
	}
	if p.config.Search != "" {
		query = query.Where("content LIKE ?", fmt.Sprintf("%%%s%%", p.config.Search))
	}

	var filteredCount int64
	query.Count(&filteredCount)

	var memos []models.LocalMemo
	if p.config.Page == "all" {
		query.Order("created_time DESC").Find(&memos)
	} else if p.config.Page != "" {
		pages := parsePageNumbers(p.config.Page)
		memos = fetchSpecificPages(query, pages, int(p.config.Limit))
	} else {
		query.Order("created_time DESC").Limit(int(p.config.Limit)).Find(&memos)
	}

	return &ListResult{
		Memos:         memos,
		TotalCount:    totalCount,
		FilteredCount: filteredCount,
		PageSize:      p.config.Limit,
	}, nil
}

func parsePageNumbers(pageStr string) []int {
	var pages []int
	parts := strings.Split(pageStr, ",")
	for _, part := range parts {
		var page int
		if _, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &page); err == nil && page > 0 {
			pages = append(pages, page)
		}
	}
	return pages
}

func fetchSpecificPages(query *gorm.DB, pages []int, pageSize int) []models.LocalMemo {
	var allMemos []models.LocalMemo

	maxPage := 1
	for _, p := range pages {
		if p > maxPage {
			maxPage = p
		}
	}

	offset := 0
	for currentPage := 1; currentPage <= maxPage; currentPage++ {
		var pageMemos []models.LocalMemo
		query.Order("created_time DESC").Offset(offset).Limit(pageSize).Find(&pageMemos)

		for _, p := range pages {
			if p == currentPage {
				allMemos = append(allMemos, pageMemos...)
				break
			}
		}

		offset += pageSize
		if len(pageMemos) < pageSize {
			break
		}
	}

	return allMemos
}

// Get

type GetConfig struct {
	Raw bool
}

type GetProcessor struct {
	config    *GetConfig
	appConfig *models.AppConfig
}

func NewGetProcessor(config *GetConfig, appConfig *models.AppConfig) *GetProcessor {
	return &GetProcessor{config: config, appConfig: appConfig}
}

func (p *GetProcessor) Execute(ctx context.Context, memoID string) (*models.Memo, error) {
	c := getClient(p.appConfig)
	return client.NewMemoService(c).Get(ctx, memoID)
}

// Create

type CreateConfig struct {
	Content    string
	FilePath   string
	Visibility string
}

type CreateResult struct {
	MemoID     string
	Visibility string
}

type CreateProcessor struct {
	config    *CreateConfig
	appConfig *models.AppConfig
}

func NewCreateProcessor(config *CreateConfig, appConfig *models.AppConfig) *CreateProcessor {
	return &CreateProcessor{config: config, appConfig: appConfig}
}

func (p *CreateProcessor) Execute(ctx context.Context) (*CreateResult, error) {
	c := getClient(p.appConfig)

	visibility := models.VisibilityPrivate
	if p.config.Visibility != "" {
		visibility = models.Visibility(strings.ToUpper(p.config.Visibility))
	}

	req := &models.CreateMemoRequest{
		Content:    p.config.Content,
		Visibility: visibility,
	}

	memo, err := client.NewMemoService(c).Create(ctx, req)
	if err != nil {
		return nil, err
	}

	memoID := memo.Uid
	if memoID == "" && memo.Name != "" {
		fmt.Sscanf(memo.Name, "memos/%s", &memoID)
	}

	return &CreateResult{MemoID: memoID, Visibility: string(memo.Visibility)}, nil
}

// Update

type UpdateConfig struct {
	MemoID     string
	Content    string
	Visibility string
	Archive    bool
	Restore    bool
	Pin        bool
	Unpin      bool
}

type UpdateProcessor struct {
	config    *UpdateConfig
	appConfig *models.AppConfig
}

func NewUpdateProcessor(config *UpdateConfig, appConfig *models.AppConfig) *UpdateProcessor {
	return &UpdateProcessor{config: config, appConfig: appConfig}
}

func (p *UpdateProcessor) Execute(ctx context.Context) (*models.Memo, error) {
	c := getClient(p.appConfig)

	req := &models.UpdateMemoRequest{}
	var updateMask []string

	if p.config.Content != "" {
		req.Content = &p.config.Content
		updateMask = append(updateMask, "content")
	}

	if p.config.Visibility != "" {
		v := models.Visibility(strings.ToUpper(p.config.Visibility))
		req.Visibility = &v
		updateMask = append(updateMask, "visibility")
	}

	if p.config.Archive {
		status := models.RowStatusArchived
		req.RowStatus = &status
		updateMask = append(updateMask, "row_status")
	}

	if p.config.Restore {
		status := models.RowStatusNormal
		req.RowStatus = &status
		updateMask = append(updateMask, "row_status")
	}

	if p.config.Pin {
		pinned := true
		req.Pinned = &pinned
		updateMask = append(updateMask, "pinned")
	}

	if p.config.Unpin {
		pinned := false
		req.Pinned = &pinned
		updateMask = append(updateMask, "pinned")
	}

	req.UpdateMask = strings.Join(updateMask, ",")

	return client.NewMemoService(c).Update(ctx, p.config.MemoID, req)
}

// Delete

type DeleteConfig struct {
	MemoID string
}

type DeleteProcessor struct {
	config    *DeleteConfig
	appConfig *models.AppConfig
}

func NewDeleteProcessor(config *DeleteConfig, appConfig *models.AppConfig) *DeleteProcessor {
	return &DeleteProcessor{config: config, appConfig: appConfig}
}

func (p *DeleteProcessor) Execute(ctx context.Context) error {
	c := getClient(p.appConfig)
	return client.NewMemoService(c).Delete(ctx, p.config.MemoID)
}

// Stats

type TagCount struct {
	Tag   string
	Count int64
}

type StatsResult struct {
	TotalCount      int64
	VisibilityCount map[string]int64
	TopTags         []TagCount
	RecentMemos     []models.LocalMemo
}

type StatsProcessor struct {
	appConfig *models.AppConfig
}

func NewStatsProcessor(appConfig *models.AppConfig) *StatsProcessor {
	return &StatsProcessor{appConfig: appConfig}
}

func (p *StatsProcessor) Execute() (*StatsResult, error) {
	db, err := utils.GetGormDB()
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	var totalCount int64
	db.Model(&models.LocalMemo{}).Where("is_deleted = ?", false).Count(&totalCount)

	type VisCount struct {
		Visibility string
		Count      int64
	}
	var visCounts []VisCount
	db.Model(&models.LocalMemo{}).
		Select("visibility, count(*) as count").
		Where("is_deleted = ?", false).
		Group("visibility").
		Find(&visCounts)
	visMap := make(map[string]int64)
	for _, v := range visCounts {
		visMap[v.Visibility] = v.Count
	}

	var properties []string
	db.Model(&models.LocalMemo{}).
		Select("DISTINCT property").
		Where("is_deleted = ? AND property != '' AND property IS NOT NULL", false).
		Find(&properties)

	tagFreq := make(map[string]int64)
	for _, prop := range properties {
		var tags []string
		if json.Unmarshal([]byte(prop), &tags) == nil {
			for _, t := range tags {
				tagFreq[t]++
			}
		}
	}
	type kv struct {
		k string
		v int64
	}
	var sorted []kv
	for k, v := range tagFreq {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].v > sorted[j].v })
	if len(sorted) > 10 {
		sorted = sorted[:10]
	}
	topTags := make([]TagCount, len(sorted))
	for i, s := range sorted {
		topTags[i] = TagCount{Tag: s.k, Count: s.v}
	}

	var recentMemos []models.LocalMemo
	db.Where("is_deleted = ?", false).Order("created_time DESC").Limit(5).Find(&recentMemos)

	return &StatsResult{
		TotalCount:      totalCount,
		VisibilityCount: visMap,
		TopTags:         topTags,
		RecentMemos:     recentMemos,
	}, nil
}

// shared

func getClient(appConfig *models.AppConfig) *client.Client {
	server := appConfig.GetDefaultServer()
	return client.NewClient(&client.Config{
		BaseURL: server.URL,
		Token:   server.Token,
		Timeout: 30 * time.Second,
	})
}
