package server

import "fmt"

func validateAddParams(name, url string) error {
	if name == "" {
		return fmt.Errorf("服务器名称不能为空")
	}
	if url == "" {
		return fmt.Errorf("服务器 URL 不能为空")
	}
	return nil
}
