package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"gorm.io/gorm"
)

// search_memos 参数
type SearchMemosParams struct {
	Keywords string   `json:"keywords"`
	Tags     []string `json:"tags,omitempty"`
	StartTime string   `json:"start_time,omitempty"` // RFC3339 or relative like "7 days ago"
	EndTime   string   `json:"end_time,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

// semantic_search 参数
type SemanticSearchParams struct {
	Query     string `json:"query"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

// get_memo 参数
type GetMemoParams struct {
	MemoID string `json:"memo_id"`
}

// ToolResult 工具执行结果
type ToolResult struct {
	Content string
	Memos   []*models.LocalMemo
}

func defineSearchMemosTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "search_memos",
			Description: "在本地备忘录中按关键词、标签、时间范围搜索。适合精确查询，如'上周的memo'、'带 #work 标签的笔记'。",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"keywords": {
						Type:        jsonschema.String,
						Description: "搜索关键词，在 memo 内容中模糊匹配",
					},
					"tags": {
						Type:        jsonschema.Array,
						Description: "按标签过滤，如 [\"work\", \"idea\"]",
						Items:       &jsonschema.Definition{Type: jsonschema.String},
					},
					"start_time": {
						Type:        jsonschema.String,
						Description: "起始时间，支持格式：RFC3339 (2026-04-13T00:00:00Z)、相对时间 (7 days ago, last monday)、自然语言 (上周, 本月, 3天前)",
					},
					"end_time": {
						Type:        jsonschema.String,
						Description: "结束时间，格式同 start_time。不填则默认为当前时间",
					},
					"limit": {
						Type:        jsonschema.Number,
						Description: "返回数量上限，默认 20",
					},
				},
			},
		},
	}
}

func defineSemanticSearchTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "semantic_search",
			Description: "语义搜索备忘录，基于向量相似度查找与查询最相关的 memo。适合模糊/概念性查询，如'关于微服务架构的思考'、'项目风险评估'。",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"query": {
						Type:        jsonschema.String,
						Description: "语义查询文本",
					},
					"start_time": {
						Type:        jsonschema.String,
						Description: "起始时间，格式 YYYY-MM-DD",
					},
					"end_time": {
						Type:        jsonschema.String,
						Description: "结束时间，格式 YYYY-MM-DD",
					},
					"limit": {
						Type:        jsonschema.Number,
						Description: "返回数量上限，默认 10",
					},
				},
				Required: []string{"query"},
			},
		},
	}
}

func defineGetMemoTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_memo",
			Description: "根据 memo ID 获取单条备忘录的完整内容",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"memo_id": {
						Type:        jsonschema.String,
						Description: "备忘录的 UID（如 20260413120000-abc123）或短 ID",
					},
				},
				Required: []string{"memo_id"},
			},
		},
	}
}

func defineMemoStatsTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "memo_stats",
			Description: "获取本地备忘录的统计概览：总数、按可见性分布、按标签分布、最近5条备忘录。适合回答'有多少条备忘录'、'备忘录概况'等总览类问题。",
			Parameters: jsonschema.Definition{
				Type:       jsonschema.Object,
				Properties: map[string]jsonschema.Definition{},
			},
		},
	}
}

// SearchMode 检索模式
type SearchMode string

const (
	SearchAuto   SearchMode = "auto"
	SearchDB     SearchMode = "db"
	SearchVector SearchMode = "vector"
)

func DefineTools(mode SearchMode) []openai.Tool {
	statsTool := defineMemoStatsTool()
	switch mode {
	case SearchDB:
		return []openai.Tool{statsTool, defineSearchMemosTool(), defineGetMemoTool()}
	case SearchVector:
		return []openai.Tool{statsTool, defineSemanticSearchTool(), defineGetMemoTool()}
	default:
		return []openai.Tool{statsTool, defineSearchMemosTool(), defineSemanticSearchTool(), defineGetMemoTool()}
	}
}

// ExecuteTool 执行 tool 调用
func ExecuteTool(ctx context.Context, db *gorm.DB, embedding *EmbeddingService, name string, arguments string) (*ToolResult, error) {
	switch name {
	case "memo_stats":
		return ExecuteMemoStats(db)
	case "search_memos":
		return ExecuteSearchMemos(db, arguments)
	case "semantic_search":
		return ExecuteSemanticSearch(ctx, embedding, arguments)
	case "get_memo":
		return ExecuteGetMemo(db, arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func ExecuteSearchMemos(db *gorm.DB, arguments string) (*ToolResult, error) {
	var params SearchMemosParams
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		return nil, fmt.Errorf("parse search_memos params: %w", err)
	}

	query := db.Model(&models.LocalMemo{}).Where("is_deleted = ?", false)

	if params.Keywords != "" {
		query = query.Where("content LIKE ?", "%"+params.Keywords+"%")
	}

	if len(params.Tags) > 0 {
		for _, tag := range params.Tags {
			query = query.Where("property LIKE ?", "%\""+tag+"\"%")
		}
	}

	// 解析时间
	if params.StartTime != "" {
		startTS, err := ParseTimeExpression(params.StartTime)
		if err == nil && startTS > 0 {
			query = query.Where("created_time >= ?", startTS)
		}
	}

	if params.EndTime != "" {
		endTS, err := ParseTimeExpression(params.EndTime)
		if err == nil && endTS > 0 {
			query = query.Where("created_time <= ?", endTS)
		}
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	var memos []models.LocalMemo
	if err := query.Order("created_time DESC").Limit(limit).Find(&memos).Error; err != nil {
		return nil, fmt.Errorf("search memos failed: %w", err)
	}

	if len(memos) == 0 {
		return &ToolResult{Content: "未找到匹配的备忘录。"}, nil
	}

	pointers := make([]*models.LocalMemo, len(memos))
	for i := range memos {
		pointers[i] = &memos[i]
	}

	return &ToolResult{Content: formatMemosForLLM(memos), Memos: pointers}, nil
}

func ExecuteSemanticSearch(ctx context.Context, embedding *EmbeddingService, arguments string) (*ToolResult, error) {
	var params SemanticSearchParams
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		return nil, fmt.Errorf("parse semantic_search params: %w", err)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}

	filter := &models.SearchFilter{
		MinScore: 0.5,
		Limit:    limit,
	}

	if params.StartTime != "" {
		if ts, err := ParseTimeExpression(params.StartTime); err == nil && ts > 0 {
			t := time.Unix(ts, 0)
			filter.StartDate = &t
		}
	}
	if params.EndTime != "" {
		if ts, err := ParseTimeExpression(params.EndTime); err == nil && ts > 0 {
			t := time.Unix(ts, 0)
			filter.EndDate = &t
		}
	}

	results, err := embedding.SemanticSearch(params.Query, filter)
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	if len(results) == 0 {
		return &ToolResult{Content: "未找到语义相关的备忘录。"}, nil
	}

	localMemos := make([]models.LocalMemo, len(results))
	pointers := make([]*models.LocalMemo, 0, len(results))
	for i, r := range results {
		if r.Memo != nil {
			localMemos[i] = *r.Memo
			pointers = append(pointers, r.Memo)
		}
	}

	return &ToolResult{Content: formatMemosForLLM(localMemos), Memos: pointers}, nil
}

func ExecuteGetMemo(db *gorm.DB, arguments string) (*ToolResult, error) {
	var params GetMemoParams
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		return nil, fmt.Errorf("parse get_memo params: %w", err)
	}

	var memo models.LocalMemo
	if err := db.Where("uid = ? OR uid = ?", params.MemoID, "memos/"+params.MemoID).
		Where("is_deleted = ?", false).
		First(&memo).Error; err != nil {
		return nil, fmt.Errorf("memo not found: %s", params.MemoID)
	}

	return &ToolResult{Content: formatMemosForLLM([]models.LocalMemo{memo}), Memos: []*models.LocalMemo{&memo}}, nil
}

func ExecuteMemoStats(db *gorm.DB) (*ToolResult, error) {
	var totalCount int64
	db.Model(&models.LocalMemo{}).Where("is_deleted = ?", false).Count(&totalCount)

	type VisibilityCount struct {
		Visibility string
		Count      int64
	}
	var visCounts []VisibilityCount
	db.Model(&models.LocalMemo{}).
		Select("visibility, count(*) as count").
		Where("is_deleted = ?", false).
		Group("visibility").
		Find(&visCounts)

	var properties []string
	db.Model(&models.LocalMemo{}).
		Select("DISTINCT property").
		Where("is_deleted = ? AND property != ?", false, "").
		Find(&properties)

	tagFreq := make(map[string]int64)
	for _, p := range properties {
		var prop struct {
			Tags []string `json:"tags"`
		}
		if json.Unmarshal([]byte(p), &prop) == nil {
			for _, t := range prop.Tags {
				tagFreq[t]++
			}
		}
	}
	type kv struct { k string; v int64 }
	var sortedTags []kv
	for k, v := range tagFreq {
		sortedTags = append(sortedTags, kv{k, v})
	}
	for i := 0; i < len(sortedTags); i++ {
		for j := i + 1; j < len(sortedTags); j++ {
			if sortedTags[j].v > sortedTags[i].v {
				sortedTags[i], sortedTags[j] = sortedTags[j], sortedTags[i]
			}
		}
	}
	if len(sortedTags) > 10 {
		sortedTags = sortedTags[:10]
	}

	var recentMemos []models.LocalMemo
	db.Where("is_deleted = ?", false).
		Order("created_time DESC").
		Limit(5).
		Find(&recentMemos)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 备忘录统计\n\n"))
	sb.WriteString(fmt.Sprintf("- **总计**: %d 条\n", totalCount))

	if len(visCounts) > 0 {
		sb.WriteString("- **可见性分布**:\n")
		for _, v := range visCounts {
			sb.WriteString(fmt.Sprintf("  - %s: %d 条\n", v.Visibility, v.Count))
		}
	}

	if len(sortedTags) > 0 {
		sb.WriteString("- **热门标签** (Top 10):\n")
		for _, t := range sortedTags {
			sb.WriteString(fmt.Sprintf("  - %s: %d 条\n", t.k, t.v))
		}
	}

	if len(recentMemos) > 0 {
		sb.WriteString(fmt.Sprintf("\n### 最近 %d 条备忘录\n\n", len(recentMemos)))
		for i, m := range recentMemos {
			uid := m.UID
			if strings.HasPrefix(uid, "memos/") {
				uid = strings.TrimPrefix(uid, "memos/")
			}
			preview := m.Content
			if len(preview) > 80 {
				preview = preview[:80] + "..."
			}
			preview = strings.ReplaceAll(preview, "\n", " ")
			sb.WriteString(fmt.Sprintf("%d. **%s** (%s) - %s\n",
				i+1, uid,
				time.Unix(m.CreatedTime, 0).Format("2006-01-02"),
				preview))
		}
	}

	pointers := make([]*models.LocalMemo, len(recentMemos))
	for i := range recentMemos {
		pointers[i] = &recentMemos[i]
	}

	return &ToolResult{Content: sb.String(), Memos: pointers}, nil
}

// formatMemosForLLM 将 memo 列表格式化为 LLM 可读文本
func formatMemosForLLM(memos []models.LocalMemo) string {
	var sb strings.Builder
	for i, m := range memos {
		uid := m.UID
		if strings.HasPrefix(uid, "memos/") {
			uid = strings.TrimPrefix(uid, "memos/")
		}

		sb.WriteString(fmt.Sprintf("## 备忘录 #%d (UID: %s)\n", i+1, uid))
		sb.WriteString(fmt.Sprintf("- **创建时间**: %s\n", time.Unix(m.CreatedTime, 0).Format("2006-01-02 15:04:05")))
		if m.UpdatedTime > 0 && m.UpdatedTime != m.CreatedTime {
			sb.WriteString(fmt.Sprintf("- **更新时间**: %s\n", time.Unix(m.UpdatedTime, 0).Format("2006-01-02 15:04:05")))
		}
		sb.WriteString(fmt.Sprintf("- **可见性**: %s\n", m.Visibility))
		if m.Property != "" {
			sb.WriteString(fmt.Sprintf("- **标签**: %s\n", m.Property))
		}
		sb.WriteString(fmt.Sprintf("- **内容**:\n```\n%s\n```\n\n", m.Content))
	}
	return sb.String()
}

// ParseTimeExpression 解析时间表达式为 Unix 时间戳
func ParseTimeExpression(expr string) (int64, error) {
	expr = strings.TrimSpace(expr)
	now := time.Now()

	// 尝试解析 RFC3339
	if t, err := time.Parse(time.RFC3339, expr); err == nil {
		return t.Unix(), nil
	}
	if t, err := time.Parse("2006-01-02", expr); err == nil {
		return t.Unix(), nil
	}
	if t, err := time.Parse("2006-01-02 15:04:05", expr); err == nil {
		return t.Unix(), nil
	}

	// 相对时间
	lower := strings.ToLower(expr)
	days := 0

	if strings.Contains(lower, "今天") || strings.Contains(lower, "today") {
		days = 0
	} else if strings.Contains(lower, "昨天") || strings.Contains(lower, "yesterday") {
		days = 1
	} else if strings.Contains(lower, "前天") {
		days = 2
	} else if strings.Contains(lower, "本周") || strings.Contains(lower, "this week") {
		weekday := now.Weekday()
		if weekday == 0 {
			weekday = 7
		}
		days = int(weekday) - 1
	} else if strings.Contains(lower, "上周") || strings.Contains(lower, "last week") {
		weekday := now.Weekday()
		if weekday == 0 {
			weekday = 7
		}
		days = int(weekday) - 1 + 7
	} else if strings.Contains(lower, "本月") || strings.Contains(lower, "this month") {
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix(), nil
	} else if strings.Contains(lower, "上月") || strings.Contains(lower, "last month") {
		return time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location()).Unix(), nil
	} else if strings.Contains(lower, "天前") || strings.Contains(lower, "days ago") {
		var n int
		fmt.Sscanf(lower, "%d", &n)
		days = n
	} else if strings.Contains(lower, "小时前") || strings.Contains(lower, "hours ago") {
		var n int
		fmt.Sscanf(lower, "%d", &n)
		return now.Add(-time.Duration(n) * time.Hour).Unix(), nil
	}

	if days > 0 {
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return start.AddDate(0, 0, -days).Unix(), nil
	}

	return 0, fmt.Errorf("无法解析时间表达式: %s", expr)
}
