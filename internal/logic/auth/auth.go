package authlogic

import (
	"context"
	"fmt"
	"strings"

	"github.com/cicbyte/memos-cli/internal/client"
	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/cicbyte/memos-cli/internal/utils"
)

// Login

type LoginConfig struct {
	Name     string
	URL      string
	Username string
	Password string
	Token    string
}

type LoginResult struct {
	User       *models.User
	ServerName string
}

type LoginProcessor struct {
	config    *LoginConfig
	appConfig *models.AppConfig
}

func NewLoginProcessor(config *LoginConfig, appConfig *models.AppConfig) *LoginProcessor {
	return &LoginProcessor{config: config, appConfig: appConfig}
}

func (p *LoginProcessor) Execute(ctx context.Context) (*LoginResult, error) {
	url := strings.TrimSuffix(p.config.URL, "/")

	// Token 登录
	if p.config.Token != "" {
		c := client.NewClient(&client.Config{BaseURL: url, Token: p.config.Token})
		user, err := client.NewAuthService(c).GetCurrentUser(ctx)
		if err != nil {
			return nil, fmt.Errorf("token 验证失败: %w", err)
		}
		serverName := p.saveServerConfig(url, p.config.Token, user)
		return &LoginResult{User: user, ServerName: serverName}, nil
	}

	// 用户名密码登录
	c := client.NewClient(&client.Config{BaseURL: url, Timeout: 0})
	resp, err := client.NewAuthService(c).SignIn(ctx, &models.SignInRequest{
		Username: p.config.Username,
		Password: p.config.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}

	serverName := p.saveServerConfig(url, resp.AccessToken, resp.User)
	return &LoginResult{User: resp.User, ServerName: serverName}, nil
}

func (p *LoginProcessor) saveServerConfig(url, token string, user *models.User) string {
	name := p.config.Name
	if name == "" {
		if user != nil && user.Username != "" {
			name = user.Username + "@" + extractHost(url)
		} else {
			name = extractHost(url)
		}
	}

	existing := p.appConfig.GetServerByName(name)
	if existing != nil {
		existing.URL = url
		existing.Token = token
		if user != nil {
			existing.Username = user.Username
		}
	} else {
		server := models.ServerConfig{
			Name:      name,
			URL:       url,
			Token:     token,
			IsDefault: len(p.appConfig.Servers) == 0,
		}
		if user != nil {
			server.Username = user.Username
		}
		p.appConfig.AddServer(server)
	}

	p.appConfig.LastServer = name
	utils.ConfigInstance.SaveConfig(p.appConfig)
	return name
}

// Logout

type LogoutConfig struct {
	ServerName string
}

type LogoutProcessor struct {
	config    *LogoutConfig
	appConfig *models.AppConfig
}

func NewLogoutProcessor(config *LogoutConfig, appConfig *models.AppConfig) *LogoutProcessor {
	return &LogoutProcessor{config: config, appConfig: appConfig}
}

func (p *LogoutProcessor) Execute(ctx context.Context) (string, error) {
	var server *models.ServerConfig
	if p.config.ServerName != "" {
		server = p.appConfig.GetServerByName(p.config.ServerName)
	} else {
		server = p.appConfig.GetDefaultServer()
	}

	if server == nil {
		return "", fmt.Errorf("未找到服务器配置")
	}
	if server.Token == "" {
		return "", fmt.Errorf("服务器 '%s' 未认证", server.Name)
	}

	c := client.NewClient(&client.Config{BaseURL: server.URL, Token: server.Token})
	_ = client.NewAuthService(c).SignOut(ctx)

	server.Token = ""
	if p.appConfig.LastServer == server.Name {
		p.appConfig.LastServer = ""
	}
	utils.ConfigInstance.SaveConfig(p.appConfig)

	return server.Name, nil
}

// Status

type StatusResult struct {
	ServerName string
	ServerURL  string
	Authenticated bool
	User       *models.User
	AuthError  error
}

type StatusProcessor struct {
	appConfig *models.AppConfig
}

func NewStatusProcessor(appConfig *models.AppConfig) *StatusProcessor {
	return &StatusProcessor{appConfig: appConfig}
}

func (p *StatusProcessor) Execute(ctx context.Context) (*StatusResult, error) {
	server := p.appConfig.GetDefaultServer()
	if server == nil {
		return &StatusResult{Authenticated: false}, nil
	}

	result := &StatusResult{
		ServerName: server.Name,
		ServerURL:  server.URL,
	}

	if server.Token == "" {
		return result, nil
	}

	c := client.NewClient(&client.Config{BaseURL: server.URL, Token: server.Token})
	user, err := client.NewAuthService(c).GetCurrentUser(ctx)
	if err != nil {
		result.AuthError = err
		return result, nil
	}

	result.Authenticated = true
	result.User = user
	return result, nil
}

func extractHost(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	parts := strings.Split(url, "/")
	return parts[0]
}
