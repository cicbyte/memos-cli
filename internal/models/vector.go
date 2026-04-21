package models

import "time"

// LocalMemo 本地存储的备忘录（用于同步和搜索）
type LocalMemo struct {
	// 远程ID（数字ID，可能为0）
	MemoID int64 `gorm:"index"`
	// UID（字符串唯一标识符）
	UID string `gorm:"index;not null;unique"`
	// 内容
	Content string `gorm:"type:text;not null"`
	// 内容 MD5 指纹（用于快速检测变更）
	ContentHash string `gorm:"index;size:32"`
	// 创建时间（Unix时间戳）
	CreatedTime int64 `gorm:"index"`
	// 更新时间（Unix时间戳）
	UpdatedTime int64 `gorm:"index"`
	// 创建者ID
	CreatorID int32
	// 可见性
	Visibility string
	// 是否置顶
	Pinned bool
	// 状态
	RowStatus string
	// 属性（JSON）
	Property string `gorm:"type:text"`
	// 父备忘录
	Parent *string

	// 本地管理字段
	RowID     uint       `gorm:"primaryKey;autoIncrement"`
	SyncedAt  time.Time  `gorm:"index"` // 同步时间
	IsDeleted bool       `gorm:"index;default:false"` // 软删除标记
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName 指定表名
func (LocalMemo) TableName() string {
	return "local_memos"
}

// MemoVector 备忘录向量
type MemoVector struct {
	// 备忘录 UID（唯一标识符）
	MemoUID    string     `gorm:"index;not null;unique"`
	Vector    string    `gorm:"type:blob;not null"` // JSON编码的[]float32
	Tags      string    `gorm:"type:text"`          // JSON数组
	CreatorID int32
	Visibility string

	// 模型信息
	Model     string    // embedding模型名称
	Dimension int       // 向量维度

	RowID     uint      `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName 指定表名
func (MemoVector) TableName() string {
	return "memo_vectors"
}

// SyncState 同步状态
type SyncState struct {
	ServerName   string    `gorm:"index;not null;unique"`
	LastSyncTime int64     // 最后同步时间戳（Unix）
	LastMemoID   int64     // 最后同步的备忘录ID
	MemoCount    int       // 已同步的备忘录数量
	SyncStatus   string    // idle, syncing, error
	ErrorMsg     string    `gorm:"type:text"`
	UpdatedAt    time.Time

	RowID     uint      `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time
}

// TableName 指定表名
func (SyncState) TableName() string {
	return "sync_states"
}

// AIConversation AI对话历史
type AIConversation struct {
	SessionID string    `gorm:"index;not null"`
	Role      string    `gorm:"not null"` // user, assistant, system
	Content   string    `gorm:"type:text;not null"`
	MemoIDs   string    `gorm:"type:text"` // 引用的备忘录ID（JSON数组）
	CreatedAt time.Time

	RowID     uint      `gorm:"primaryKey;autoIncrement"`
}

// TableName 指定表名
func (AIConversation) TableName() string {
	return "ai_conversations"
}

// SemanticSearchResult 语义搜索结果
type SemanticSearchResult struct {
	Memo      *LocalMemo
	Score     float64
	Highlight string
}

// AIAnswer AI回答
type AIAnswer struct {
	Answer   string
	Sources  []*LocalMemo
	UsedTokens int
	Model    string
}

// SearchFilter 搜索过滤器
type SearchFilter struct {
	Tags       []string
	Visibility string
	CreatorID  int32
	StartDate  *time.Time
	EndDate    *time.Time
	MinScore   float64
	Limit      int
}
