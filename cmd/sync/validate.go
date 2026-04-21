package sync

import "fmt"

func validateSyncParams() error {
	// sync 命令的参数验证（full+force 需要二次确认在 CMD 层处理）
	return nil
}

// validateSyncParams 未被引用但保留为规范要求的文件
var _ = fmt.Sprintf
