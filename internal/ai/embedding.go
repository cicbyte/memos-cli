package ai

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
)

const (
	// defaultMaxTextLen 单条文本最大字符数（中文约 2-3 token/字符，保守限制）
	defaultMaxTextLen = 2000
	// defaultMaxBatchLen 批量嵌入时单条文本最大字符数
	defaultMaxBatchLen = 8000
)

// EmbeddingService Embedding服务
type EmbeddingService struct {
	client    *resty.Client
	model     string
	baseURL   string
	dimension int
	db        *gorm.DB
}

// NewEmbeddingService 创建Embedding服务
func NewEmbeddingService(baseURL, model string, dimension int, db *gorm.DB) *EmbeddingService {
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL = strings.TrimSuffix(baseURL, "/") + "/v1"
	}

	client := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(0)

	return &EmbeddingService{
		client:    client,
		model:     model,
		baseURL:   baseURL,
		dimension: dimension,
		db:        db,
	}
}

func (e *EmbeddingService) IsAvailable() bool {
	checkClient := resty.New().SetBaseURL(e.baseURL).SetTimeout(3 * time.Second)
	resp, err := checkClient.R().Get("/models")
	return err == nil && resp.StatusCode() == 200
}

// OllamaEmbedRequest Ollama Embed请求
type OllamaEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"` // 使用 input 而不是 prompt，支持批量
}

// OllamaEmbedResponse Ollama Embed响应
type OllamaEmbedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

// Embed 生成单个文本的向量
func (s *EmbeddingService) Embed(text string) ([]float32, error) {
	// 清理文本并限制长度
	// nomic-embed-text 的上下文窗口是 8192 tokens
	// 对于中文，保守估计 1 字符 ≈ 2-3 tokens
	// 限制在 2000 字符（约 4000-6000 tokens）
	maxLength := defaultMaxTextLen
	cleanText := s.cleanText(text, maxLength)

	// 检查清理后的文本是否为空
	if len(strings.TrimSpace(cleanText)) == 0 {
		return nil, fmt.Errorf("empty text after cleaning")
	}

	req := OllamaEmbedRequest{
		Model: s.model,
		Input: []string{cleanText},
	}

	var resp OllamaEmbedResponse
	r, err := s.client.R().
		SetBody(req).
		SetResult(&resp).
		Post("/embeddings")

	if err != nil {
		return nil, fmt.Errorf("ollama embed request failed: %w", err)
	}

	if r.StatusCode() != 200 {
		return nil, fmt.Errorf("ollama returned status %d (text length: %d)",
			r.StatusCode(), len([]rune(text)))
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("ollama returned empty data array")
	}

	if len(resp.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("ollama returned empty embedding vector")
	}

	return resp.Data[0].Embedding, nil
}

// EmbedBatch 批量生成向量
func (s *EmbeddingService) EmbedBatch(texts []string) ([][]float32, error) {
	// 清理文本
	cleanTexts := make([]string, len(texts))
	for i, text := range texts {
		cleanTexts[i] = s.cleanText(text, defaultMaxBatchLen)
	}

	req := OllamaEmbedRequest{
		Model: s.model,
		Input: cleanTexts,
	}

	var resp OllamaEmbedResponse
	_, err := s.client.R().
		SetBody(req).
		SetResult(&resp).
		Post("/embeddings")

	if err != nil {
		return nil, fmt.Errorf("ollama batch embed request failed: %w", err)
	}

	if len(resp.Data) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(resp.Data))
	}

	vectors := make([][]float32, len(texts))
	for i, data := range resp.Data {
		if len(data.Embedding) == 0 {
			return nil, fmt.Errorf("empty embedding at index %d", i)
		}
		vectors[i] = data.Embedding
	}

	return vectors, nil
}

// CosineSimilarity 计算余弦相似度
func (s *EmbeddingService) CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct float64
	var normA float64
	var normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// IndexMemo 为备忘录建立向量索引
func (s *EmbeddingService) IndexMemo(memo *models.LocalMemo) error {
	// 生成向量
	vector, err := s.Embed(memo.Content)
	if err != nil {
		// 提供更详细的错误信息，包括内容预览
		contentPreview := memo.Content
		if len(contentPreview) > 100 {
			contentPreview = contentPreview[:100] + "..."
		}
		return fmt.Errorf("failed to embed memo %s (content: %q, length: %d): %w",
			memo.UID, contentPreview, len(memo.Content), err)
	}

	// 序列化向量
	vectorJSON, err := json.Marshal(vector)
	if err != nil {
		return fmt.Errorf("failed to marshal vector: %w", err)
	}

	// 查找是否已存在
	var existing models.MemoVector
	err = s.db.Where("memo_uid = ?", memo.UID).First(&existing).Error

	memoVector := &models.MemoVector{
		MemoUID:    memo.UID,
		Vector:     string(vectorJSON),
		Tags:       memo.Property,
		CreatorID:  memo.CreatorID,
		Visibility: memo.Visibility,
		Model:      s.model,
		Dimension:  len(vector),
	}

	if err == gorm.ErrRecordNotFound {
		// 新建
		if err := s.db.Create(memoVector).Error; err != nil {
			return fmt.Errorf("failed to create memo vector: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to query existing vector: %w", err)
	} else {
		// 更新
		if err := s.db.Model(&existing).Updates(memoVector).Error; err != nil {
			return fmt.Errorf("failed to update memo vector: %w", err)
		}
	}

	return nil
}

// IndexMemos 批量为备忘录建立向量索引
func (s *EmbeddingService) IndexMemos(memos []*models.LocalMemo) error {
	successCount := 0
	skipCount := 0
	failCount := 0

	for i, memo := range memos {
		// 跳过空内容
		if len(strings.TrimSpace(memo.Content)) == 0 {
			skipCount++
			continue
		}

		if err := s.IndexMemo(memo); err != nil {
			failCount++
			// 记录错误但继续处理其他备忘录
			contentPreview := strings.ReplaceAll(memo.Content, "\n", "\\n")
			if len(contentPreview) > 200 {
				contentPreview = contentPreview[:200] + "..."
			}
			fmt.Printf("⚠️  向量化失败 [%d/%d, UID=%s]: %v\n", i+1, len(memos), memo.UID, err)
			fmt.Printf("   内容预览: %s\n", contentPreview)
			fmt.Printf("   内容长度: %d 字符\n", len(memo.Content))
		} else {
			successCount++

			// 每10条显示一次进度
			if (i+1)%10 == 0 {
				fmt.Printf("  向量化进度: %d/%d\r", i+1, len(memos))
			}
		}
	}

	if skipCount > 0 {
		fmt.Printf("\nℹ️  跳过 %d 条空内容的备忘录\n", skipCount)
	}

	if failCount > 0 {
		fmt.Printf("\n⚠️  向量化失败: %d 条\n", failCount)
	}

	fmt.Printf("\n✅ 向量化完成: 成功 %d/%d\n", successCount, len(memos))

	return nil
}

// SearchResult 搜索结果
type SearchResult struct {
	Memo  *models.LocalMemo
	Score float64
}

// SemanticSearch 语义搜索
func (s *EmbeddingService) SemanticSearch(query string, filter *models.SearchFilter) ([]*SearchResult, error) {
	// 1. 向量化查询
	queryVector, err := s.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// 2. 从数据库获取所有向量
	var vectors []models.MemoVector
	dbQuery := s.db.Model(&models.MemoVector{})

	// 应用过滤器
	if filter != nil {
		if filter.Visibility != "" {
			dbQuery = dbQuery.Where("visibility = ?", filter.Visibility)
		}
		if len(filter.Tags) > 0 {
			for _, tag := range filter.Tags {
				dbQuery = dbQuery.Where("tags LIKE ?", "%"+tag+"%")
			}
		}
	}

	if err := dbQuery.Find(&vectors).Error; err != nil {
		return nil, fmt.Errorf("failed to query vectors: %w", err)
	}

	// 3. 计算相似度
	type resultWithScore struct {
		result SearchResult
		vector  []float32
	}
	results := make([]resultWithScore, 0, len(vectors))

	for _, v := range vectors {
		// 反序列化向量
		var vector []float32
		if err := json.Unmarshal([]byte(v.Vector), &vector); err != nil {
			continue // 跳过无法解析的向量
		}

		// 计算相似度
		score := s.CosineSimilarity(queryVector, vector)

		// 过滤低分结果
		if filter != nil && score < filter.MinScore {
			continue
		}

		// 获取备忘录
		var memo models.LocalMemo
		memoQuery := s.db.Where("uid = ?", v.MemoUID)
		if filter != nil {
			if filter.StartDate != nil {
				memoQuery = memoQuery.Where("created_time >= ?", filter.StartDate.Unix())
			}
			if filter.EndDate != nil {
				memoQuery = memoQuery.Where("created_time <= ?", filter.EndDate.Unix()+86400-1)
			}
		}
		if err := memoQuery.First(&memo).Error; err != nil {
			continue
		}

		results = append(results, resultWithScore{
			result: SearchResult{
				Memo:  &memo,
				Score: score,
			},
			vector: vector,
		})
	}

	// 4. 排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].result.Score > results[j].result.Score
	})

	// 5. 限制结果数量
	limit := len(results)
	if filter != nil && filter.Limit > 0 && filter.Limit < limit {
		limit = filter.Limit
	}

	// 6. 返回结果
	finalResults := make([]*SearchResult, limit)
	for i := 0; i < limit; i++ {
		finalResults[i] = &results[i].result
	}

	return finalResults, nil
}

// cleanText 清理文本，限制长度
func (s *EmbeddingService) cleanText(text string, maxChars int) string {
	runes := []rune(text)
	if len(runes) <= maxChars {
		return text
	}
	// 文本被截断
	return string(runes[:maxChars]) + "..."
}

// DeleteMemoVector 删除备忘录向量
func (s *EmbeddingService) DeleteMemoVector(memoUID string) error {
	return s.db.Where("memo_uid = ?", memoUID).Delete(&models.MemoVector{}).Error
}

// RebuildAll 重建所有向量索引
func (s *EmbeddingService) RebuildAll(progress func(current, total int)) error {
	// 1. 获取所有备忘录
	var memos []models.LocalMemo
	if err := s.db.Where("is_deleted = ?", false).Find(&memos).Error; err != nil {
		return fmt.Errorf("failed to query memos: %w", err)
	}

	// 2. 删除现有向量
	if err := s.db.Exec("DELETE FROM memo_vectors").Error; err != nil {
		return fmt.Errorf("failed to clear vectors: %w", err)
	}

	// 3. 重新建立索引
	total := len(memos)
	for i, memo := range memos {
		if err := s.IndexMemo(&memo); err != nil {
			return fmt.Errorf("failed to index memo %d: %w", memo.MemoID, err)
		}

		if progress != nil {
			progress(i+1, total)
		}
	}

	return nil
}

// GetVectorStats 获取向量统计信息
func (s *EmbeddingService) GetVectorStats() (map[string]interface{}, error) {
	var count int64
	if err := s.db.Model(&models.MemoVector{}).Count(&count).Error; err != nil {
		return nil, err
	}

	// 获取不同模型的统计
	var modelStats []struct {
		Model   string
		Count   int64
		AvgDim  float64
	}

	if err := s.db.Model(&models.MemoVector{}).
		Select("model, COUNT(*) as count, AVG(dimension) as avg_dim").
		Group("model").
		Scan(&modelStats).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_count":  count,
		"model_stats":  modelStats,
		"current_model": s.model,
	}, nil
}
