package ai

import (
	"context"
	"strings"

	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

type AIService struct {
	embedding *EmbeddingService
	llmClient *openai.Client
	model     string
	baseURL   string
	provider  string
	apiKey    string
	db        *gorm.DB
}

func NewAIService(provider, baseURL, model, apiKey string, embedding *EmbeddingService, db *gorm.DB) *AIService {
	if provider == "ollama" {
		if !strings.HasSuffix(baseURL, "/v1") {
			baseURL = strings.TrimSuffix(baseURL, "/") + "/v1"
		}
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	return &AIService{
		embedding: embedding,
		llmClient: openai.NewClientWithConfig(config),
		model:     model,
		baseURL:   baseURL,
		provider:  provider,
		apiKey:    apiKey,
		db:        db,
	}
}

type AskRequest struct {
	Question     string
	Filter       *models.SearchFilter
	ContextLimit int
	SearchMode   SearchMode
}

type AskResponse struct {
	Answer           string
	Sources          []*models.LocalMemo
	Model            string
	PromptTokens     int
	CompletionTokens int
}

type ChatMessage struct {
	Role    string
	Content string
}

func (s *AIService) Ask(ctx context.Context, req *AskRequest) (*AskResponse, error) {
	agent := NewAgent(s.llmClient, s.embedding, s.db, s.model)
	result, err := agent.Ask(ctx, req.Question, req.SearchMode)
	if err != nil {
		return nil, err
	}
	return &AskResponse{
		Answer:           result.Answer,
		Sources:          result.Sources,
		Model:            result.Model,
		PromptTokens:     result.PromptTokens,
		CompletionTokens: result.CompletionTokens,
	}, nil
}

func (s *AIService) AskWithHistory(ctx context.Context, question string, history []ChatMessage, mode SearchMode) (*AskResponse, error) {
	agent := NewAgent(s.llmClient, s.embedding, s.db, s.model)
	result, err := agent.AskWithHistory(ctx, question, history, mode)
	if err != nil {
		return nil, err
	}
	return &AskResponse{
		Answer:           result.Answer,
		Sources:          result.Sources,
		Model:            result.Model,
		PromptTokens:     result.PromptTokens,
		CompletionTokens: result.CompletionTokens,
	}, nil
}

func (s *AIService) AskStream(ctx context.Context, req *AskRequest, cb StreamCallback) error {
	agent := NewAgent(s.llmClient, s.embedding, s.db, s.model)
	return agent.AskStream(ctx, req.Question, req.SearchMode, cb)
}

func (s *AIService) AskWithHistoryStream(ctx context.Context, question string, history []ChatMessage, mode SearchMode, cb StreamCallback) error {
	agent := NewAgent(s.llmClient, s.embedding, s.db, s.model)
	return agent.AskWithHistoryStream(ctx, question, history, mode, cb)
}
