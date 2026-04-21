package utils

import (
	"fmt"
	"os"
)

// 确保目录存在，如果不存在则创建
func EnsureDir(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	}
	return nil
}

// 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// 初始化应用目录结构
func InitAppDirs() error {
	config := ConfigInstance

	// 检查并创建各级目录
	dirs := []string{
		config.GetAppSeriesDir(),
		config.GetAppDir(),
		config.GetConfigDir(),
		config.GetDbDir(),
		config.GetLogDir(),
	}

	for _, dir := range dirs {
		if err := EnsureDir(dir); err != nil {
			return fmt.Errorf("directory init failed: %v", err)
		}
	}

	return nil
}
