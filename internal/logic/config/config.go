package configlogic

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/cicbyte/memos-cli/internal/utils"
)

// Add

type AddConfig struct {
	Name     string
	URL      string
	Token    string
	Default  bool
}

type AddProcessor struct {
	config    *AddConfig
	appConfig *models.AppConfig
}

func NewAddProcessor(config *AddConfig, appConfig *models.AppConfig) *AddProcessor {
	return &AddProcessor{config: config, appConfig: appConfig}
}

func (p *AddProcessor) Execute() (string, bool, error) {
	name := p.config.Name
	url := strings.TrimSuffix(p.config.URL, "/")

	if name == "" {
		return "", false, fmt.Errorf("服务器名称不能为空")
	}
	if url == "" {
		return "", false, fmt.Errorf("服务器 URL 不能为空")
	}

	if p.appConfig.GetServerByName(name) != nil {
		return "", false, fmt.Errorf("服务器 '%s' 已存在", name)
	}

	setDefault := p.config.Default
	if len(p.appConfig.Servers) == 0 {
		setDefault = true
	}

	server := models.ServerConfig{
		Name:      name,
		URL:       url,
		Token:     p.config.Token,
		IsDefault: setDefault,
	}
	p.appConfig.AddServer(server)
	utils.ConfigInstance.SaveConfig(p.appConfig)

	return name, setDefault, nil
}

// List

type ListResult struct {
	Servers   []ServerInfo
	LastUsed  string
}

type ServerInfo struct {
	Name     string
	URL      string
	IsDefault bool
	TokenPreview string
}

type ListProcessor struct {
	appConfig *models.AppConfig
}

func NewListProcessor(appConfig *models.AppConfig) *ListProcessor {
	return &ListProcessor{appConfig: appConfig}
}

func (p *ListProcessor) Execute() *ListResult {
	servers := make([]ServerInfo, 0, len(p.appConfig.Servers))
	for _, s := range p.appConfig.Servers {
		tokenPreview := "no token"
		if s.Token != "" {
			if len(s.Token) > 8 {
				tokenPreview = s.Token[:4] + "..." + s.Token[len(s.Token)-4:]
			} else {
				tokenPreview = "***"
			}
		}
		servers = append(servers, ServerInfo{
			Name:        s.Name,
			URL:         s.URL,
			IsDefault:   s.IsDefault,
			TokenPreview: tokenPreview,
		})
	}
	return &ListResult{Servers: servers, LastUsed: p.appConfig.LastServer}
}

// Default

type DefaultConfig struct {
	Name string
}

type DefaultProcessor struct {
	config    *DefaultConfig
	appConfig *models.AppConfig
}

func NewDefaultProcessor(config *DefaultConfig, appConfig *models.AppConfig) *DefaultProcessor {
	return &DefaultProcessor{config: config, appConfig: appConfig}
}

func (p *DefaultProcessor) Execute() error {
	if p.appConfig.GetServerByName(p.config.Name) == nil {
		return fmt.Errorf("服务器 '%s' 不存在", p.config.Name)
	}
	if !p.appConfig.SetDefaultServer(p.config.Name) {
		return fmt.Errorf("设置默认服务器失败")
	}
	utils.ConfigInstance.SaveConfig(p.appConfig)
	return nil
}

// Remove

type RemoveConfig struct {
	Name string
}

type RemoveProcessor struct {
	config    *RemoveConfig
	appConfig *models.AppConfig
}

func NewRemoveProcessor(config *RemoveConfig, appConfig *models.AppConfig) *RemoveProcessor {
	return &RemoveProcessor{config: config, appConfig: appConfig}
}

func (p *RemoveProcessor) Execute() (string, error) {
	if p.appConfig.GetServerByName(p.config.Name) == nil {
		return "", fmt.Errorf("服务器 '%s' 不存在", p.config.Name)
	}
	if !p.appConfig.RemoveServer(p.config.Name) {
		return "", fmt.Errorf("删除服务器失败")
	}
	utils.ConfigInstance.SaveConfig(p.appConfig)
	return p.config.Name, nil
}

// ConfigItem 配置项元数据
type ConfigItem struct {
	Key       string
	Section   string
	Type      string
	Sensitive bool
	Desc      string
}

func AllConfigItems() []ConfigItem {
	return []ConfigItem{
		{Key: "ai.provider", Section: "AI", Type: "string", Desc: "LLM 提供商 (ollama/openai/zhipu)"},
		{Key: "ai.base_url", Section: "AI", Type: "string", Desc: "LLM API 地址"},
		{Key: "ai.model", Section: "AI", Type: "string", Desc: "LLM 模型名称"},
		{Key: "ai.api_key", Section: "AI", Type: "string", Sensitive: true, Desc: "LLM API 密钥"},
		{Key: "ai.max_tokens", Section: "AI", Type: "int", Desc: "最大 token 数"},
		{Key: "ai.temperature", Section: "AI", Type: "float", Desc: "温度参数 (0.0-2.0)"},
		{Key: "ai.timeout", Section: "AI", Type: "int", Desc: "请求超时秒数"},
		{Key: "embedding.provider", Section: "Embedding", Type: "string", Desc: "Embedding 提供商"},
		{Key: "embedding.base_url", Section: "Embedding", Type: "string", Desc: "Embedding API 地址"},
		{Key: "embedding.model", Section: "Embedding", Type: "string", Desc: "Embedding 模型名称"},
		{Key: "embedding.api_key", Section: "Embedding", Type: "string", Sensitive: true, Desc: "Embedding API 密钥"},
		{Key: "embedding.timeout", Section: "Embedding", Type: "int", Desc: "请求超时秒数"},
		{Key: "log.level", Section: "Log", Type: "string", Desc: "日志级别 (debug/info/warn/error)"},
		{Key: "log.max_size", Section: "Log", Type: "int", Desc: "单个日志文件最大 MB"},
		{Key: "log.max_backups", Section: "Log", Type: "int", Desc: "日志备份数量"},
		{Key: "log.max_age", Section: "Log", Type: "int", Desc: "日志保留天数"},
		{Key: "log.compress", Section: "Log", Type: "bool", Desc: "是否压缩日志"},
	}
}

func FindConfigItem(key string) *ConfigItem {
	items := AllConfigItems()
	for i := range items {
		if items[i].Key == key {
			return &items[i]
		}
	}
	return nil
}

// GetResult 单个配置项查询结果
type GetResult struct {
	Item  ConfigItem
	Value string
}

type GetProcessor struct {
	appConfig *models.AppConfig
}

func NewGetProcessor(appConfig *models.AppConfig) *GetProcessor {
	return &GetProcessor{appConfig: appConfig}
}

func (p *GetProcessor) Execute(key string) (*GetResult, error) {
	item := FindConfigItem(key)
	if item == nil {
		return nil, fmt.Errorf("未知配置项: %s\n使用 'memos-cli config list' 查看所有配置项", key)
	}

	value := GetConfigValue(p.appConfig, key)
	return &GetResult{Item: *item, Value: value}, nil
}

type SetProcessor struct {
	appConfig *models.AppConfig
}

func NewSetProcessor(appConfig *models.AppConfig) *SetProcessor {
	return &SetProcessor{appConfig: appConfig}
}

func (p *SetProcessor) Execute(key, value string) error {
	item := FindConfigItem(key)
	if item == nil {
		return fmt.Errorf("未知配置项: %s\n使用 'memos-cli config list' 查看所有配置项", key)
	}

	if err := setConfigValue(p.appConfig, key, value); err != nil {
		return fmt.Errorf("设置失败: %w", err)
	}

	utils.ConfigInstance.SaveConfig(p.appConfig)
	return nil
}

func GetConfigValue(c *models.AppConfig, key string) string {
	switch key {
	case "ai.provider":
		return c.AI.Provider
	case "ai.base_url":
		return c.AI.BaseURL
	case "ai.model":
		return c.AI.Model
	case "ai.api_key":
		return c.AI.ApiKey
	case "ai.max_tokens":
		return strconv.Itoa(c.AI.MaxTokens)
	case "ai.temperature":
		return strconv.FormatFloat(c.AI.Temperature, 'f', -1, 64)
	case "ai.timeout":
		return strconv.Itoa(c.AI.Timeout)
	case "embedding.provider":
		return c.Embedding.Provider
	case "embedding.base_url":
		return c.Embedding.BaseURL
	case "embedding.model":
		return c.Embedding.Model
	case "embedding.api_key":
		return c.Embedding.ApiKey
	case "embedding.timeout":
		return strconv.Itoa(c.Embedding.Timeout)
	case "log.level":
		return c.Log.Level
	case "log.max_size":
		return strconv.Itoa(c.Log.MaxSize)
	case "log.max_backups":
		return strconv.Itoa(c.Log.MaxBackups)
	case "log.max_age":
		return strconv.Itoa(c.Log.MaxAge)
	case "log.compress":
		return strconv.FormatBool(c.Log.Compress)
	default:
		return ""
	}
}

func setConfigValue(c *models.AppConfig, key, value string) error {
	switch key {
	case "ai.provider":
		c.AI.Provider = value
	case "ai.base_url":
		c.AI.BaseURL = value
	case "ai.model":
		c.AI.Model = value
	case "ai.api_key":
		c.AI.ApiKey = value
	case "ai.max_tokens":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("无效的整数值: %s", value)
		}
		c.AI.MaxTokens = v
	case "ai.temperature":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("无效的浮点数值: %s", value)
		}
		c.AI.Temperature = v
	case "ai.timeout":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("无效的整数值: %s", value)
		}
		c.AI.Timeout = v
	case "embedding.provider":
		c.Embedding.Provider = value
	case "embedding.base_url":
		c.Embedding.BaseURL = value
	case "embedding.model":
		c.Embedding.Model = value
	case "embedding.api_key":
		c.Embedding.ApiKey = value
	case "embedding.timeout":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("无效的整数值: %s", value)
		}
		c.Embedding.Timeout = v
	case "log.level":
		c.Log.Level = value
	case "log.max_size":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("无效的整数值: %s", value)
		}
		c.Log.MaxSize = v
	case "log.max_backups":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("无效的整数值: %s", value)
		}
		c.Log.MaxBackups = v
	case "log.max_age":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("无效的整数值: %s", value)
		}
		c.Log.MaxAge = v
	case "log.compress":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("无效的布尔值: %s (true/false)", value)
		}
		c.Log.Compress = v
	}
	return nil
}
