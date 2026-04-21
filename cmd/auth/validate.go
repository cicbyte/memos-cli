package auth

import "fmt"

func validateLoginParams(url, token, username, password string) error {
	if url == "" {
		return fmt.Errorf("请指定服务器地址")
	}
	if token == "" && (username == "" || password == "") {
		return fmt.Errorf("请指定用户名和密码（或使用 --token）")
	}
	return nil
}
