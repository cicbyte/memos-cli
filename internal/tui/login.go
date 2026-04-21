package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/cicbyte/memos-cli/internal/client"
	"github.com/cicbyte/memos-cli/internal/common"
	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/cicbyte/memos-cli/internal/utils"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LoginModel 登录视图模型
type LoginModel struct {
	// 输入框
	urlInput      textinput.Model
	usernameInput textinput.Model
	passwordInput textinput.Model

	// 状态
	focused    int // 0: URL, 1: Username, 2: Password
	loading    bool
	err        error
	errMessage string

	// 尺寸
	width  int
	height int
}

// NewLoginModel 创建登录模型
func NewLoginModel() *LoginModel {
	// URL 输入框
	urlInput := textinput.New()
	urlInput.Placeholder = "https://memos.example.com"
	urlInput.Focus()
	urlInput.CharLimit = 256

	// 用户名输入框
	usernameInput := textinput.New()
	usernameInput.Placeholder = "username"
	usernameInput.CharLimit = 64

	// 密码输入框
	passwordInput := textinput.New()
	passwordInput.Placeholder = "password"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.CharLimit = 64

	return &LoginModel{
		urlInput:      urlInput,
		usernameInput: usernameInput,
		passwordInput: passwordInput,
		focused:       0,
	}
}

// Init 初始化
func (m *LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update 更新
func (m *LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.nextField()

		case "shift+tab":
			m.prevField()

		case "up":
			m.prevField()

		case "down":
			m.nextField()

		case "enter":
			return m, m.loginCmd

		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case LoginSuccessMsg:
		m.loading = false
		return m, func() tea.Msg {
			return AuthSuccessMsg{
				User:   msg.User,
				Server: msg.Server,
				Client: msg.Client,
			}
		}

	case LoginFailMsg:
		m.loading = false
		m.err = msg.Err
		m.errMessage = msg.Message
	}

	// 更新输入框
	var cmd tea.Cmd
	switch m.focused {
	case 0:
		m.urlInput, cmd = m.urlInput.Update(msg)
	case 1:
		m.usernameInput, cmd = m.usernameInput.Update(msg)
	case 2:
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View 渲染
func (m *LoginModel) View() string {
	// 标题
	title := titleStyle.Render("Welcome to Memos CLI")
	subtitle := subtitleStyle.Render("Login to your Memos server")

	// 输入框
	urlLabel := labelStyle.Render("Server URL:")
	urlInput := m.styleInput(m.urlInput, m.focused == 0)

	usernameLabel := labelStyle.Render("Username:")
	usernameInput := m.styleInput(m.usernameInput, m.focused == 1)

	passwordLabel := labelStyle.Render("Password:")
	passwordInput := m.styleInput(m.passwordInput, m.focused == 2)

	// 错误信息
	var errorMsg string
	if m.errMessage != "" {
		errorMsg = errorStyle.Render("✗ " + m.errMessage)
	}

	// 提示
	help := helpStyle.Render("Tab to switch fields • Enter to login • Ctrl+C to quit")

	// 组合界面
	content := fmt.Sprintf(`
%s
%s

%s
%s

%s
%s

%s
%s

%s
`,
		title, subtitle,
		urlLabel, urlInput,
		usernameLabel, usernameInput,
		passwordLabel, passwordInput,
		errorMsg,
	)

	// 加载状态
	if m.loading {
		content += "\n  Logging in...\n"
	}

	content += "\n" + help

	// 居中
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		boxStyle.Render(content),
	)
}

// nextField 切换到下一个输入框
func (m *LoginModel) nextField() {
	m.focused = (m.focused + 1) % 3
	m.updateFocus()
}

// prevField 切换到上一个输入框
func (m *LoginModel) prevField() {
	m.focused = (m.focused - 1 + 3) % 3
	m.updateFocus()
}

// updateFocus 更新焦点状态
func (m *LoginModel) updateFocus() {
	m.urlInput.Blur()
	m.usernameInput.Blur()
	m.passwordInput.Blur()

	switch m.focused {
	case 0:
		m.urlInput.Focus()
	case 1:
		m.usernameInput.Focus()
	case 2:
		m.passwordInput.Focus()
	}
}

// styleInput 样式化输入框
func (m *LoginModel) styleInput(input textinput.Model, focused bool) string {
	if focused {
		return focusedInputStyle.Render(input.View())
	}
	return inputStyle.Render(input.View())
}

// loginCmd 登录命令
func (m *LoginModel) loginCmd() tea.Msg {
	url := m.urlInput.Value()
	username := m.usernameInput.Value()
	password := m.passwordInput.Value()

	if url == "" {
		return LoginFailMsg{Message: "Server URL is required"}
	}
	if username == "" {
		return LoginFailMsg{Message: "Username is required"}
	}
	if password == "" {
		return LoginFailMsg{Message: "Password is required"}
	}

	// 创建客户端
	c := client.NewClient(&client.Config{
		BaseURL: url,
		Timeout: 30 * time.Second,
	})

	// 登录
	resp, err := client.NewAuthService(c).SignIn(context.Background(), &models.SignInRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return LoginFailMsg{
			Err:     err,
			Message: "Login failed: " + err.Error(),
		}
	}

	// 保存配置
	serverName := username + "@" + extractHost(url)
	server := models.ServerConfig{
		Name:      serverName,
		URL:       url,
		Token:     resp.AccessToken,
		IsDefault: len(common.GetAppConfig().Servers) == 0,
		Username:  username,
	}

	// 检查是否已存在
	existing := common.GetAppConfig().GetServerByName(serverName)
	if existing != nil {
		existing.Token = resp.AccessToken
	} else {
		common.GetAppConfig().AddServer(server)
	}
	common.GetAppConfig().LastServer = serverName
	utils.ConfigInstance.SaveConfig(common.GetAppConfig())

	return LoginSuccessMsg{
		User:   resp.User,
		Server: &server,
		Client: c,
	}
}

// extractHost 从 URL 提取主机名
func extractHost(url string) string {
	// 简单实现
	for i, c := range url {
		if c == '/' && i > 6 {
			return url[8:i] // https:// 之后的第一个 /
		}
	}
	return url
}

// 消息类型
type LoginSuccessMsg struct {
	User   *models.User
	Server *models.ServerConfig
	Client *client.Client
}

type LoginFailMsg struct {
	Err     error
	Message string
}

// 样式
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			MarginBottom(2)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			MarginTop(1)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("0")).
			Padding(0, 1).
			Width(50)

	focusedInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("4")).
				Padding(0, 1).
				Width(50)

	boxStyle = lipgloss.NewStyle().
			Padding(2, 4).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("6"))
)
