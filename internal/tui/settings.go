package tui

import (
	"fmt"

	"github.com/cicbyte/memos-cli/internal/common"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SettingsModel 设置视图模型
type SettingsModel struct {
	// 列表
	list list.Model

	// 尺寸
	width  int
	height int
}

// settingsItem 设置项
type settingsItem struct {
	title       string
	description string
	action      string
}

func (i settingsItem) FilterValue() string {
	return i.title
}

func (i settingsItem) Title() string {
	return i.title
}

func (i settingsItem) Description() string {
	return i.description
}

// NewSettingsModel 创建设置模型
func NewSettingsModel() *SettingsModel {
	// 创建列表项
	items := []list.Item{
		settingsItem{
			title:       "Servers",
			description: "Manage server configurations",
			action:      "servers",
		},
		settingsItem{
			title:       "Account",
			description: "View account information",
			action:      "account",
		},
		settingsItem{
			title:       "Logout",
			description: "Logout from current server",
			action:      "logout",
		},
	}

	// 创建列表
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("4")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4"))

	l := list.New(items, delegate, 0, 0)
	l.Title = "Settings"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("5")).
		Padding(0, 1)

	return &SettingsModel{
		list: l,
	}
}

// Init 初始化
func (m *SettingsModel) Init() tea.Cmd {
	return nil
}

// Update 更新
func (m *SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(settingsItem); ok {
				return m, m.handleAction(item.action)
			}

		case "esc":
			return m, func() tea.Msg {
				return ViewChangeMsg{View: ViewMemoList}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-2)
	}

	// 更新列表
	newList, cmd := m.list.Update(msg)
	m.list = newList
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// handleAction 处理操作
func (m *SettingsModel) handleAction(action string) tea.Cmd {
	switch action {
	case "servers":
		// TODO: 实现服务器管理
		return nil

	case "account":
		// 显示账户信息
		return func() tea.Msg {
			return ShowAccountMsg{}
		}

	case "logout":
		// 登出
		server := common.GetAppConfig().GetDefaultServer()
		if server != nil {
			server.Token = ""
			common.GetAppConfig().LastServer = ""
		}
		return func() tea.Msg {
			return ViewChangeMsg{View: ViewLogin}
		}
	}
	return nil
}

// View 渲染
func (m *SettingsModel) View() string {
	// 添加服务器信息
	server := common.GetAppConfig().GetDefaultServer()
	var serverInfo string
	if server != nil {
		serverInfo = fmt.Sprintf("\n  Current Server: %s\n  URL: %s\n",
			server.Name, server.URL)
	}

	return serverInfo + m.list.View()
}

// 消息类型
type ShowAccountMsg struct{}

type ShowServersMsg struct{}
