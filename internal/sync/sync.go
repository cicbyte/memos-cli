package sync

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cicbyte/memos-cli/internal/ai"
	"github.com/cicbyte/memos-cli/internal/client"
	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/cicbyte/memos-cli/internal/utils"
	"gorm.io/gorm"
)

type SyncService struct {
	db         *gorm.DB
	memoClient *client.MemoService
	embedding  *ai.EmbeddingService
	serverName string
}

func (s *SyncService) canVectorize() bool {
	return s.embedding != nil
}

func NewSyncService(db *gorm.DB, baseURL, token string, embedding *ai.EmbeddingService, serverName string) *SyncService {
	cfg := &client.Config{
		BaseURL: baseURL,
		Token:   token,
	}
	apiClient := client.NewClient(cfg)
	memoClient := client.NewMemoService(apiClient)

	return &SyncService{
		db:         db,
		memoClient: memoClient,
		embedding:  embedding,
		serverName: serverName,
	}
}

type SyncResult struct {
	Added      int
	Updated    int
	Deleted    int
	Skipped    int
	Vectorized int
	Duration   time.Duration
	Error      error
}

type SyncOptions struct {
	FullSync       bool
	AutoVectorize  bool
	ForceVectorize bool
}

func (s *SyncService) Sync(opts *SyncOptions) *SyncResult {
	startTime := time.Now()
	result := &SyncResult{}

	state, err := utils.GetSyncState(s.db, s.serverName)
	if err != nil {
		result.Error = fmt.Errorf("failed to get sync state: %w", err)
		return result
	}

	if err := utils.UpdateSyncState(s.db, s.serverName, map[string]interface{}{
		"sync_status": "syncing",
		"updated_at":  time.Now(),
	}); err != nil {
		result.Error = fmt.Errorf("failed to update sync status: %w", err)
		return result
	}

	if opts.FullSync {
		if err := s.fullSyncCleanup(); err != nil {
			result.Error = fmt.Errorf("failed to cleanup for full sync: %w", err)
			return result
		}
		state.LastSyncTime = 0
		state.LastMemoID = 0
	}

	remoteMemos, err := s.fetchRemoteMemos(state.LastSyncTime)
	if err != nil {
		result.Error = fmt.Errorf("failed to fetch remote memos: %w", err)
		s.updateSyncStatus("error", err.Error())
		return result
	}

	for _, memo := range remoteMemos {
		added, updated, skipped, err := s.syncMemo(memo)
		if err != nil {
			fmt.Printf("  同步备忘录 %d 失败: %v\n", getMemoID(memo.Name), err)
			continue
		}
		result.Added += added
		result.Updated += updated
		result.Skipped += skipped
	}

	if err := s.syncDeleted(remoteMemos, result); err != nil {
		fmt.Printf("  同步删除失败: %v\n", err)
	}

	if (opts.AutoVectorize || opts.ForceVectorize) && s.canVectorize() {
		vectorized, err := s.vectorizeMemos(opts.ForceVectorize)
		if err != nil {
			fmt.Printf("  向量化失败: %v\n", err)
		} else {
			result.Vectorized = vectorized
		}
	}

	result.Duration = time.Since(startTime)
	now := time.Now()
	s.updateSyncStatus("idle", "")
	utils.UpdateSyncState(s.db, s.serverName, map[string]interface{}{
		"last_sync_time": now.Unix(),
		"memo_count":     s.getLocalMemoCount(),
		"updated_at":     now,
	})

	return result
}

func (s *SyncService) fetchRemoteMemos(lastSyncTime int64) ([]*models.Memo, error) {
	var allMemos []*models.Memo
	pageSize := int32(100)
	pageToken := ""

	for {
		opts := &client.ListOptions{
			PageSize:  pageSize,
			PageToken: pageToken,
		}

		resp, err := s.memoClient.List(context.Background(), opts)
		if err != nil {
			return nil, err
		}

		allMemos = append(allMemos, resp.Memos...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return allMemos, nil
}

func (s *SyncService) syncMemo(memo *models.Memo) (added, updated, skipped int, err error) {
	uid := extractMemoUID(memo.Name)

	if uid == "" {
		return 0, 0, 0, fmt.Errorf("invalid memo name: %s", memo.Name)
	}

	localMemo := s.convertToLocalMemo(memo)

	var existing models.LocalMemo
	err = s.db.Where("uid = ?", uid).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		if err := s.db.Create(localMemo).Error; err != nil {
			return 0, 0, 0, err
		}
		return 1, 0, 0, nil
	} else if err != nil {
		return 0, 0, 0, err
	}

	if existing.ContentHash == localMemo.ContentHash {
		return 0, 0, 1, nil
	}

	if s.canVectorize() {
		if err := s.embedding.DeleteMemoVector(existing.UID); err != nil {
			fmt.Printf("  删除旧向量失败 [%s]: %v\n", safeUID(existing.UID), err)
		}
	}

	localMemo.RowID = existing.RowID
	if err := s.db.Save(localMemo).Error; err != nil {
		return 0, 0, 0, err
	}

	return 0, 1, 0, nil
}

func (s *SyncService) convertToLocalMemo(memo *models.Memo) *models.LocalMemo {
	memoID := getMemoID(memo.Name)
	uid := extractMemoUID(memo.Name)
	creatorID := getCreatorID(memo.Creator)

	propertyJSON := ""
	if memo.Property != nil && len(memo.Property.Tags) > 0 {
		tagsJSON, _ := json.Marshal(memo.Property.Tags)
		propertyJSON = string(tagsJSON)
	}

	var parent *string
	if memo.Parent != nil && *memo.Parent != "" {
		parent = memo.Parent
	}

	contentHash := calculateMD5(memo.Content)

	return &models.LocalMemo{
		MemoID:      memoID,
		UID:         uid,
		Content:     memo.Content,
		ContentHash: contentHash,
		CreatedTime: timeToUnix(memo.CreateTime),
		UpdatedTime: timeToUnix(memo.UpdateTime),
		CreatorID:   creatorID,
		Visibility:  string(memo.Visibility),
		Pinned:      memo.Pinned,
		RowStatus:   string(memo.RowStatus),
		Property:    propertyJSON,
		Parent:      parent,
		IsDeleted:   memo.RowStatus == models.RowStatusArchived,
		SyncedAt:    time.Now(),
	}
}

func (s *SyncService) syncDeleted(remoteMemos []*models.Memo, result *SyncResult) error {
	remoteUIDs := make(map[string]bool)
	for _, memo := range remoteMemos {
		uid := extractMemoUID(memo.Name)
		if uid != "" {
			remoteUIDs[uid] = true
		}
	}

	var localDeleted []models.LocalMemo
	if err := s.db.Where("is_deleted = ?", true).Find(&localDeleted).Error; err != nil {
		return err
	}

	for _, local := range localDeleted {
		if !remoteUIDs[local.UID] {
			if err := s.db.Delete(&local).Error; err == nil {
				result.Deleted++
			} else {
				fmt.Printf("删除本地 memo %s 失败: %v\n", local.UID, err)
			}
		}
	}

	return nil
}

func (s *SyncService) vectorizeMemos(force bool) (int, error) {
	var memos []*models.LocalMemo

	if force {
		if err := s.db.Where("is_deleted = ?", false).Find(&memos).Error; err != nil {
			return 0, err
		}
		s.db.Exec("DELETE FROM memo_vectors")
	} else {
		if err := s.db.Raw(`
			SELECT lm.* FROM local_memos lm
			LEFT JOIN memo_vectors mv ON lm.uid = mv.memo_uid
			WHERE lm.is_deleted = ? AND mv.row_id IS NULL
		`, false).Scan(&memos).Error; err != nil {
			return 0, err
		}
	}

	if len(memos) == 0 {
		return 0, nil
	}

	if err := s.embedding.IndexMemos(memos); err != nil {
		return 0, err
	}

	return len(memos), nil
}

func (s *SyncService) fullSyncCleanup() error {
	if err := s.db.Exec("DELETE FROM local_memos").Error; err != nil {
		return err
	}
	if err := s.db.Exec("DELETE FROM memo_vectors").Error; err != nil {
		return err
	}
	return nil
}

func (s *SyncService) getLocalMemoCount() int {
	var count int64
	s.db.Model(&models.LocalMemo{}).Where("is_deleted = ?", false).Count(&count)
	return int(count)
}

func (s *SyncService) updateSyncStatus(status, errorMsg string) {
	updates := map[string]interface{}{
		"sync_status": status,
		"updated_at":  time.Now(),
	}
	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}
	utils.UpdateSyncState(s.db, s.serverName, updates)
}

func getMemoID(name string) int64 {
	var id int64
	fmt.Sscanf(name, "memos/%d", &id)
	return id
}

func extractMemoUID(name string) string {
	if len(name) > 7 && name[:7] == "memos/" {
		return name[7:]
	}
	return ""
}

func getCreatorID(creator string) int32 {
	var id int32
	if creator != "" {
		fmt.Sscanf(creator, "users/%d", &id)
	}
	return id
}

func timeToUnix(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.Unix()
}

func calculateMD5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func safeUID(uid string) string {
	if len(uid) > 8 {
		return uid[:8]
	}
	return uid
}
