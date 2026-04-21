package models

// ServerConfig Memos 服务器配置
type ServerConfig struct {
	Name     string `yaml:"name"`      // 服务器名称
	URL      string `yaml:"url"`       // 服务器地址 (e.g., https://memos.example.com)
	Token    string `yaml:"token"`     // Bearer Token
	IsDefault bool   `yaml:"is_default"` // 是否为默认服务器
	Username string `yaml:"username"`  // 用户名 (可选，用于显示)
}

type AppConfig struct {
	Version string `yaml:"version"` // 版本号，用于升级时判断

	// Memos 服务器配置
	Servers    []ServerConfig `yaml:"servers"`     // 服务器列表
	LastServer string         `yaml:"last_server"` // 上次使用的服务器名称

	AI struct {
		Provider    string  `yaml:"provider"` // openai/ollama/zhipu
		BaseURL     string  `yaml:"base_url"`
		ApiKey      string  `yaml:"api_key"`
		Model       string  `yaml:"model"`
		MaxTokens   int     `yaml:"max_tokens"`
		Temperature float64 `yaml:"temperature"`
		Timeout     int     `yaml:"timeout"`
	} `yaml:"ai"`

	Embedding struct {
		Provider string `yaml:"provider"` // ollama/openai/zhipu
		BaseURL  string `yaml:"base_url"`
		ApiKey   string `yaml:"api_key"`
		Model    string `yaml:"model"`
		Timeout  int    `yaml:"timeout"`
	} `yaml:"embedding"`

	Log struct {
		Level      string `yaml:"level"`
		MaxSize    int    `yaml:"maxSize"`
		MaxBackups int    `yaml:"maxBackups"`
		MaxAge     int    `yaml:"maxAge"`
		Compress   bool   `yaml:"compress"`
	} `yaml:"log"`
}

// GetDefaultServer 获取默认服务器配置
func (c *AppConfig) GetDefaultServer() *ServerConfig {
	// 优先返回上次使用的服务器
	if c.LastServer != "" {
		for i := range c.Servers {
			if c.Servers[i].Name == c.LastServer {
				return &c.Servers[i]
			}
		}
	}
	// 其次返回标记为默认的服务器
	for i := range c.Servers {
		if c.Servers[i].IsDefault {
			return &c.Servers[i]
		}
	}
	// 最后返回第一个服务器
	if len(c.Servers) > 0 {
		return &c.Servers[0]
	}
	return nil
}

// GetServerByName 根据名称获取服务器配置
func (c *AppConfig) GetServerByName(name string) *ServerConfig {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			return &c.Servers[i]
		}
	}
	return nil
}

// AddServer 添加服务器配置
func (c *AppConfig) AddServer(server ServerConfig) {
	// 如果是第一个服务器，设为默认
	if len(c.Servers) == 0 {
		server.IsDefault = true
	}
	// 如果新服务器设为默认，取消其他服务器的默认标记
	if server.IsDefault {
		for i := range c.Servers {
			c.Servers[i].IsDefault = false
		}
	}
	c.Servers = append(c.Servers, server)
}

// RemoveServer 删除服务器配置
func (c *AppConfig) RemoveServer(name string) bool {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			wasDefault := c.Servers[i].IsDefault
			c.Servers = append(c.Servers[:i], c.Servers[i+1:]...)
			// 如果删除的是默认服务器，将第一个设为默认
			if wasDefault && len(c.Servers) > 0 {
				c.Servers[0].IsDefault = true
			}
			return true
		}
	}
	return false
}

// SetDefaultServer 设置默认服务器
func (c *AppConfig) SetDefaultServer(name string) bool {
	found := false
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			c.Servers[i].IsDefault = true
			found = true
		} else {
			c.Servers[i].IsDefault = false
		}
	}
	return found
}
