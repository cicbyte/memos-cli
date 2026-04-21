package memo

import (
	"fmt"

	"github.com/cicbyte/memos-cli/internal/common"
)

func checkServerConfig() error {
	server := common.GetAppConfig().GetDefaultServer()
	if server == nil {
		return fmt.Errorf("❌ 未配置服务器。\n\n请先运行 'memos-cli auth login' 登录服务器。")
	}
	if server.Token == "" {
		return fmt.Errorf("❌ 未认证。\n\n请先运行 'memos-cli auth login' 进行认证。")
	}
	return nil
}
