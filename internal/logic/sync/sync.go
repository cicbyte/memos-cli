package synclogic

import (
	"fmt"
	"time"

	"github.com/cicbyte/memos-cli/internal/ai"
	"github.com/cicbyte/memos-cli/internal/models"
	syncpkg "github.com/cicbyte/memos-cli/internal/sync"
	"github.com/cicbyte/memos-cli/internal/utils"
)

// Sync

type SyncConfig struct {
	FullSync      bool
	Force         bool
	NoVectorize   bool
	Verbose       bool
}

type SyncResult struct {
	ServerName   string
	LastSyncTime int64
	Added        int
	Updated      int
	Deleted      int
	Skipped      int
	Vectorized   int
	Duration     time.Duration
}

type SyncProcessor struct {
	config    *SyncConfig
	appConfig *models.AppConfig
}

func NewSyncProcessor(config *SyncConfig, appConfig *models.AppConfig) *SyncProcessor {
	return &SyncProcessor{config: config, appConfig: appConfig}
}

func (p *SyncProcessor) Execute() (*SyncResult, error) {
	db, err := utils.GetGormDB()
	if err != nil {
		return nil, fmt.Errorf("数据库初始化失败: %w", err)
	}
	if err := utils.InitSchema(db); err != nil {
		return nil, fmt.Errorf("初始化数据库表失败: %w", err)
	}

	serverConfig := p.appConfig.GetDefaultServer()
	if serverConfig == nil {
		return nil, fmt.Errorf("未配置服务器")
	}

	embeddingService := p.newEmbeddingService()

	syncService := syncpkg.NewSyncService(db, serverConfig.URL, serverConfig.Token, embeddingService, serverConfig.Name)

	if embeddingService == nil && !p.config.NoVectorize {
		fmt.Println("  Embedding 服务不可用，跳过向量化步骤")
	}

	opts := &syncpkg.SyncOptions{
		FullSync:       p.config.FullSync,
		AutoVectorize: !p.config.NoVectorize,
		ForceVectorize: p.config.FullSync && p.config.Force,
	}

	result := syncService.Sync(opts)
	if result.Error != nil {
		return nil, result.Error
	}

	return &SyncResult{
		ServerName: serverConfig.Name,
		Added:      result.Added,
		Updated:    result.Updated,
		Deleted:    result.Deleted,
		Skipped:    result.Skipped,
		Vectorized: result.Vectorized,
		Duration:   result.Duration,
	}, nil
}

func (p *SyncProcessor) GetLastSyncTime() (int64, error) {
	db, _ := utils.GetGormDB()
	serverConfig := p.appConfig.GetDefaultServer()
	if serverConfig == nil {
		return 0, nil
	}
	state, err := utils.GetSyncState(db, serverConfig.Name)
	if err != nil {
		return 0, err
	}
	if state == nil {
		return 0, nil
	}
	return state.LastSyncTime, nil
}

func (p *SyncProcessor) newEmbeddingService() *ai.EmbeddingService {
	url := p.appConfig.Embedding.BaseURL
	if url == "" {
		url = "http://localhost:11434/v1"
	}
	model := p.appConfig.Embedding.Model
	if model == "" {
		model = "nomic-embed-text"
	}
	db, _ := utils.GetGormDB()
	svc := ai.NewEmbeddingService(url, model, 768, db)
	if !svc.IsAvailable() {
		return nil
	}
	return svc
}

// Status

type StatusResult struct {
	ServerName  string
	LastSyncTime int64
	MemoCount   int
	SyncStatus  string
	ErrorMsg    string
}

type StatusProcessor struct {
	appConfig *models.AppConfig
}

func NewStatusProcessor(appConfig *models.AppConfig) *StatusProcessor {
	return &StatusProcessor{appConfig: appConfig}
}

func (p *StatusProcessor) Execute() (*StatusResult, error) {
	db, err := utils.GetGormDB()
	if err != nil {
		return nil, fmt.Errorf("数据库初始化失败: %w", err)
	}

	serverConfig := p.appConfig.GetDefaultServer()
	if serverConfig == nil {
		return nil, fmt.Errorf("未配置服务器")
	}

	state, err := utils.GetSyncState(db, serverConfig.Name)
	if err != nil {
		return nil, fmt.Errorf("获取同步状态失败: %w", err)
	}

	return &StatusResult{
		ServerName:   serverConfig.Name,
		LastSyncTime: state.LastSyncTime,
		MemoCount:    state.MemoCount,
		SyncStatus:   state.SyncStatus,
		ErrorMsg:     state.ErrorMsg,
	}, nil
}
