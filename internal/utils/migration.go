package utils

import (
	"fmt"
	"time"

	"github.com/cicbyte/memos-cli/internal/models"
	"gorm.io/gorm"
)

// InitSchema 初始化数据库表结构
func InitSchema(db *gorm.DB) error {
	// 创建本地备忘录表
	if err := db.AutoMigrate(&models.LocalMemo{}); err != nil {
		return fmt.Errorf("failed to migrate LocalMemo: %w", err)
	}

	// 创建备忘录向量表
	if err := db.AutoMigrate(&models.MemoVector{}); err != nil {
		return fmt.Errorf("failed to migrate MemoVector: %w", err)
	}

	// 创建同步状态表
	if err := db.AutoMigrate(&models.SyncState{}); err != nil {
		return fmt.Errorf("failed to migrate SyncState: %w", err)
	}

	// 创建AI对话历史表
	if err := db.AutoMigrate(&models.AIConversation{}); err != nil {
		return fmt.Errorf("failed to migrate AIConversation: %w", err)
	}

	// 创建索引
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// createIndexes 创建必要的索引
func createIndexes(db *gorm.DB) error {
	// LocalMemo 索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_local_memos_created_time ON local_memos(created_time)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_local_memos_updated_time ON local_memos(updated_time)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_local_memos_is_deleted ON local_memos(is_deleted)").Error; err != nil {
		return err
	}

	// MemoVector 索引 (使用 memo_uid 而不是 memo_id)
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_memo_vectors_memo_uid ON memo_vectors(memo_uid)").Error; err != nil {
		return err
	}

	// SyncState 索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_sync_states_server_name ON sync_states(server_name)").Error; err != nil {
		return err
	}

	// AIConversation 索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_conversations_session_id ON ai_conversations(session_id)").Error; err != nil {
		return err
	}

	return nil
}

// GetSyncState 获取同步状态
func GetSyncState(db *gorm.DB, serverName string) (*models.SyncState, error) {
	var state models.SyncState
	err := db.Where("server_name = ?", serverName).First(&state).Error
	if err == gorm.ErrRecordNotFound {
		// 创建新的同步状态
		now := time.Now()
		state = models.SyncState{
			ServerName:   serverName,
			LastSyncTime: 0,
			LastMemoID:   0,
			MemoCount:    0,
			SyncStatus:   "idle",
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := db.Create(&state).Error; err != nil {
			return nil, err
		}
		return &state, nil
	}
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// UpdateSyncState 更新同步状态
func UpdateSyncState(db *gorm.DB, serverName string, updates map[string]interface{}) error {
	return db.Model(&models.SyncState{}).
		Where("server_name = ?", serverName).
		Updates(updates).Error
}
