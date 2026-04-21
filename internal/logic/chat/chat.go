package chatlogic

import (
	"context"
	"fmt"
	"sync"

	"github.com/cicbyte/memos-cli/internal/ai"
	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/cicbyte/memos-cli/internal/utils"
	"gorm.io/gorm"
)

type Config struct {
	Tags       []string
	Visibility string
	Limit      int
	SearchMode ai.SearchMode
}

type Processor struct {
	config    *Config
	appConfig *models.AppConfig
	mu        sync.Mutex
	sessions  map[string][]ai.ChatMessage
}

type ChatMessage = ai.ChatMessage

func NewProcessor(config *Config, appConfig *models.AppConfig) *Processor {
	return &Processor{
		config:    config,
		appConfig: appConfig,
		sessions:  make(map[string][]ai.ChatMessage),
	}
}

func (p *Processor) Execute(ctx context.Context, question string) (*ai.AskResponse, error) {
	db, err := utils.GetGormDB()
	if err != nil {
		return nil, fmt.Errorf("数据库初始化失败: %w", err)
	}

	aiService := p.newAIService(db)
	req := &ai.AskRequest{
		Question: question,
		Filter: &models.SearchFilter{
			Tags:       p.config.Tags,
			Visibility: p.config.Visibility,
			MinScore:   0.6,
			Limit:      p.limit(),
		},
		ContextLimit: p.limit(),
		SearchMode:   p.config.SearchMode,
	}

	return aiService.Ask(ctx, req)
}

func (p *Processor) ExecuteWithSession(ctx context.Context, sessionID, question string) (*ai.AskResponse, error) {
	db, err := utils.GetGormDB()
	if err != nil {
		return nil, fmt.Errorf("数据库初始化失败: %w", err)
	}

	aiService := p.newAIService(db)

	p.mu.Lock()
	p.sessions[sessionID] = append(p.sessions[sessionID], ai.ChatMessage{Role: "user", Content: question})
	history := p.sessions[sessionID]
	p.mu.Unlock()

	resp, err := aiService.AskWithHistory(ctx, question, history, p.config.SearchMode)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	p.sessions[sessionID] = append(p.sessions[sessionID], ai.ChatMessage{Role: "assistant", Content: resp.Answer})
	p.mu.Unlock()

	return resp, nil
}

func (p *Processor) ExecuteStream(ctx context.Context, question string, cb ai.StreamCallback) error {
	db, err := utils.GetGormDB()
	if err != nil {
		return fmt.Errorf("数据库初始化失败: %w", err)
	}
	aiService := p.newAIService(db)
	req := &ai.AskRequest{
		Question:   question,
		SearchMode: p.config.SearchMode,
	}
	return aiService.AskStream(ctx, req, cb)
}

func (p *Processor) ExecuteWithSessionStream(ctx context.Context, sessionID, question string, cb ai.StreamCallback) error {
	db, err := utils.GetGormDB()
	if err != nil {
		return fmt.Errorf("数据库初始化失败: %w", err)
	}
	aiService := p.newAIService(db)

	p.mu.Lock()
	p.sessions[sessionID] = append(p.sessions[sessionID], ai.ChatMessage{Role: "user", Content: question})
	history := p.sessions[sessionID]
	p.mu.Unlock()

	var answer string
	wrappedCb := func(event ai.StreamEvent) {
		if event.Type == "content" {
			answer += event.Content
		}
		if event.Type == "done" {
			p.mu.Lock()
			p.sessions[sessionID] = append(p.sessions[sessionID], ai.ChatMessage{Role: "assistant", Content: answer})
			p.mu.Unlock()
		}
		cb(event)
	}

	return aiService.AskWithHistoryStream(ctx, question, history, p.config.SearchMode, wrappedCb)
}

func (p *Processor) NewSession() string {
	return fmt.Sprintf("session_%d", len(p.sessions)+1)
}

func (p *Processor) ClearSession(sessionID string) {
	p.mu.Lock()
	delete(p.sessions, sessionID)
	p.mu.Unlock()
}

func (p *Processor) limit() int {
	if p.config.Limit > 0 {
		return p.config.Limit
	}
	return 10
}

func (p *Processor) newAIService(db *gorm.DB) *ai.AIService {
	embedding := p.newEmbeddingService(db)
	return ai.NewAIService(
		p.appConfig.AI.Provider,
		p.appConfig.AI.BaseURL,
		p.appConfig.AI.Model,
		p.appConfig.AI.ApiKey,
		embedding,
		db,
	)
}

func (p *Processor) newEmbeddingService(db *gorm.DB) *ai.EmbeddingService {
	url := p.appConfig.Embedding.BaseURL
	if url == "" {
		url = "http://localhost:11434/v1"
	}
	model := p.appConfig.Embedding.Model
	if model == "" {
		model = "nomic-embed-text"
	}
	return ai.NewEmbeddingService(url, model, 768, db)
}
