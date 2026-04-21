package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cicbyte/memos-cli/internal/client"
	"github.com/cicbyte/memos-cli/internal/common"
	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/cicbyte/memos-cli/internal/utils"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// 视图类型
type ViewType int

const (
	ViewLogin ViewType = iota
	ViewMemoList
	ViewMemoDetail
	ViewMemoCreate
	ViewMemoSearch
	ViewSettings
	ViewHelp
)

// Model TUI 主模型
type Model struct {
	// 状态
	currentView  ViewType
	ready        bool
	width        int
	height       int
	loading      bool
	loadingMsg   string
	err          error

	// 认证状态
	authenticated bool
	currentUser   *models.User
	server        *models.ServerConfig

	// API 客户端
	client *client.Client

	// 子视图数据
	memos        []*models.Memo
	selectedMemo *models.Memo
	memoCursor   int
	searchQuery  string

	// 组件
	loginModel       *LoginModel
	memoListModel    *MemoListModel
	memoDetailModel  *MemoDetailModel
	memoCreateModel  *MemoCreateModel
	memoSearchModel  *MemoSearchModel
	settingsModel    *SettingsModel
}

// NewModel 创建新模型
func NewModel() Model {
	m := Model{
		currentView: ViewLogin,
	}

	// 检查是否已认证
	server := common.GetAppConfig().GetDefaultServer()
	if server != nil && server.Token != "" {
		m.server = server
		m.authenticated = true
		m.currentView = ViewMemoList

		// 创建客户端
		m.client = client.NewClient(&client.Config{
			BaseURL: server.URL,
			Token:   server.Token,
			Timeout: 30 * time.Second,
		})
	}

	// 初始化子模型
	m.loginModel = NewLoginModel()
	m.memoListModel = NewMemoListModel()
	m.settingsModel = NewSettingsModel()

	// 如果已认证，设置 client 到子模型
	if m.client != nil {
		m.memoListModel.client = m.client
	}

	return m
}

// Init 初始化
func (m Model) Init() tea.Cmd {
	// 如果已经认证，加载 memos
	if m.authenticated && m.client != nil {
		return tea.Batch(
			loadMemosCmd(m.client),
		)
	}
	// 否则返回一个空命令
	return nil
}

// Update 更新
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentView != ViewHelp {
				return m, tea.Quit
			}
		case "?":
			if m.currentView != ViewHelp && m.currentView != ViewLogin {
				m.currentView = ViewHelp
				return m, nil
			}
		case "esc":
			if m.currentView == ViewHelp {
				m.currentView = ViewMemoList
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// 更新子组件尺寸
		m.memoListModel.width = m.width
		m.memoListModel.height = m.height

	case AuthSuccessMsg:
		m.authenticated = true
		m.currentUser = msg.User
		m.server = msg.Server
		m.client = msg.Client
		m.currentView = ViewMemoList
		// 设置 client 到子模型
		m.memoListModel.client = m.client
		return m, loadMemosCmd(m.client)

	case AuthFailMsg:
		m.err = msg.Err
		m.authenticated = false
		m.currentView = ViewLogin

	case MemosLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.memos = msg.Memos
			m.memoListModel.memos = msg.Memos
		}

	case LoadingMsg:
		m.loading = true
		m.loadingMsg = msg.Message

	case ViewChangeMsg:
		m.currentView = msg.View
		if msg.Memo != nil {
			m.selectedMemo = msg.Memo
			m.memoDetailModel = NewMemoDetailModel(msg.Memo, m.client, msg.PrevView)
		}
		// 初始化创建/搜索视图（仅首次进入时创建，返回时复用保留状态）
		if msg.View == ViewMemoCreate && m.client != nil && m.memoCreateModel == nil {
			m.memoCreateModel = NewMemoCreateModel(m.client)
		}
		if msg.View == ViewMemoSearch && m.client != nil && m.memoSearchModel == nil {
			m.memoSearchModel = NewMemoSearchModel(m.client)
		}
		// 返回列表时重新加载
		if msg.View == ViewMemoList && m.client != nil && msg.Reload {
			return m, loadMemosCmd(m.client)
		}
	}

	// 根据当前视图更新子组件（ViewChangeMsg 已在上面处理，不再传递）
	if _, ok := msg.(ViewChangeMsg); !ok {
	switch m.currentView {
	case ViewLogin:
		loginModel, cmd := m.loginModel.Update(msg)
		m.loginModel = loginModel.(*LoginModel)
		cmds = append(cmds, cmd)

	case ViewMemoList:
		listModel, cmd := m.memoListModel.Update(msg)
		m.memoListModel = listModel.(*MemoListModel)
		cmds = append(cmds, cmd)

	case ViewMemoDetail:
		if m.memoDetailModel != nil {
			detailModel, cmd := m.memoDetailModel.Update(msg)
			m.memoDetailModel = detailModel.(*MemoDetailModel)
			cmds = append(cmds, cmd)
		}

	case ViewMemoCreate:
		if m.memoCreateModel != nil {
			createModel, cmd := m.memoCreateModel.Update(msg)
			m.memoCreateModel = createModel.(*MemoCreateModel)
			cmds = append(cmds, cmd)
		}

	case ViewMemoSearch:
		if m.memoSearchModel != nil {
			searchModel, cmd := m.memoSearchModel.Update(msg)
			m.memoSearchModel = searchModel.(*MemoSearchModel)
			cmds = append(cmds, cmd)
		}

	case ViewSettings:
		settingsModel, cmd := m.settingsModel.Update(msg)
		m.settingsModel = settingsModel.(*SettingsModel)
		cmds = append(cmds, cmd)
	}
	}

	return m, tea.Batch(cmds...)
}

// View 渲染
func (m Model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	// 显示加载状态
	if m.loading {
		return fmt.Sprintf("\n  %s...", m.loadingMsg)
	}

	// 显示错误
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("\n  Error: %v", m.err))
	}

	// 根据视图渲染
	var content string
	switch m.currentView {
	case ViewLogin:
		content = m.loginModel.View()

	case ViewMemoList:
		content = m.memoListModel.View()

	case ViewMemoDetail:
		if m.memoDetailModel != nil {
			content = m.memoDetailModel.View()
		}

	case ViewMemoCreate:
		if m.memoCreateModel != nil {
			content = m.memoCreateModel.View()
		}

	case ViewMemoSearch:
		if m.memoSearchModel != nil {
			content = m.memoSearchModel.View()
		}

	case ViewSettings:
		content = m.settingsModel.View()

	case ViewHelp:
		content = m.renderHelp()

	default:
		content = "Unknown view"
	}

	// 添加状态栏
	return m.renderWithStatusBar(content)
}

// renderWithStatusBar 渲染带状态栏的界面
func (m Model) renderWithStatusBar(content string) string {
	// 状态栏
	var status string
	if m.authenticated && m.server != nil {
		user := "user"
		// 优先从 currentUser 获取
		if m.currentUser != nil && m.currentUser.Username != "" {
			user = m.currentUser.Username
		} else {
			// 从服务器名称中提取用户名 (格式: username@server)
			if parts := strings.Split(m.server.Name, "@"); len(parts) > 0 {
				user = parts[0]
			}
		}
		status = fmt.Sprintf(" %s@%s | ?帮助 q退出 ", user, m.server.Name)
	} else {
		status = " 未登录 | q退出 "
	}

	// 填充到固定宽度，避免重绘时宽度变化
	if len(status) < m.width {
		status = status + strings.Repeat(" ", m.width-len(status))
	} else if len(status) > m.width {
		status = status[:m.width]
	}

	statusBar := statusBarStyle.Render(status)

	// 组合界面 - 使用字符串拼接代替 JoinVertical 减少渲染开销
	return content + "\n" + statusBar
}

// renderHelp 渲染帮助信息
func (m Model) renderHelp() string {
	helpText := `
  Keyboard Shortcuts
  ─────────────────────────────

  Navigation:
    j/k, ↑/↓    Move up/down
    h/l, ←/→    Navigate panels
    Enter       Select/Open
    Esc         Back/Cancel

  Actions:
    n           New memo
    e           Edit memo
    d           Delete memo
    r           Refresh
    s           Settings
    /           Search

  Memo Create:
    Tab         Switch fields
    Ctrl+V      Toggle visibility
    Ctrl+S      Save memo

  General:
    ?           Show this help
    q/Ctrl+c    Quit

  Press Esc to close
`
	return helpStyle.Render(helpText)
}

// 消息类型
type AuthSuccessMsg struct {
	User   *models.User
	Server *models.ServerConfig
	Client *client.Client
}

type AuthFailMsg struct {
	Err error
}

type MemosLoadedMsg struct {
	Memos []*models.Memo
	Err   error
}

type LoadingMsg struct {
	Message string
}

type ViewChangeMsg struct {
	View      ViewType
	PrevView  ViewType
	Memo      *models.Memo
	Reload    bool // 返回列表时是否重新加载数据
}

// 命令
func checkAuthCmd() tea.Msg {
	server := common.GetAppConfig().GetDefaultServer()
	if server == nil || server.Token == "" {
		return nil
	}

	c := client.NewClient(&client.Config{
		BaseURL: server.URL,
		Token:   server.Token,
		Timeout: 10 * time.Second,
	})

	user, err := client.NewAuthService(c).GetCurrentUser(context.Background())
	if err != nil {
		return AuthFailMsg{Err: err}
	}

	return AuthSuccessMsg{
		User:   user,
		Server: server,
		Client: c,
	}
}

func loadMemosCmd(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		// 从本地数据库加载备忘录
		db, err := utils.GetGormDB()
		if err != nil {
			return MemosLoadedMsg{Err: err}
		}

		var localMemos []models.LocalMemo
		// 只加载未删除的备忘录，按创建时间降序
		if err := db.Where("is_deleted = ?", false).Order("created_time DESC").Limit(100).Find(&localMemos).Error; err != nil {
			return MemosLoadedMsg{Err: err}
		}

		// 转换为远程模型格式（TUI 使用）
		memos := make([]*models.Memo, 0, len(localMemos))
		for _, lm := range localMemos {
			memo := &models.Memo{
				Name:       lm.UID,
				Uid:        strings.TrimPrefix(lm.UID, "memos/"),
				Content:    lm.Content,
				Visibility: models.Visibility(lm.Visibility),
				Pinned:     lm.Pinned,
				RowStatus:  models.RowStatus(lm.RowStatus),
			}

			// 转换时间
			if lm.CreatedTime > 0 {
				t := time.Unix(lm.CreatedTime, 0)
				memo.CreateTime = &t
				memo.DisplayTime = &t
			}
			if lm.UpdatedTime > 0 {
				t := time.Unix(lm.UpdatedTime, 0)
				memo.UpdateTime = &t
			}

			// 解析 Property JSON
			if lm.Property != "" {
				var prop models.MemoProperty
				if err := json.Unmarshal([]byte(lm.Property), &prop); err == nil {
					memo.Property = &prop
				}
			}

			memos = append(memos, memo)
		}

		return MemosLoadedMsg{Memos: memos}
	}
}

// 样式
var (
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Padding(1, 2)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("8")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Padding(2, 4).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("6"))
)
