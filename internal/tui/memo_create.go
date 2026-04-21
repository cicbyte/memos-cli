package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/cicbyte/memos-cli/internal/client"
	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MemoCreateModel Memo 创建视图模型
type MemoCreateModel struct {
	// 输入组件
	contentArea  textarea.Model
	tagInput     textinput.Model

	// 状态
	focused      int // 0: content, 1: tags
	visibility   models.Visibility
	loading      bool
	err          error
	errMessage   string
	success      bool

	// API 客户端
	client *client.Client

	// 尺寸
	width  int
	height int
}

// NewMemoCreateModel 创建 Memo 创建模型
func NewMemoCreateModel(c *client.Client) *MemoCreateModel {
	// 内容输入区
	contentArea := textarea.New()
	contentArea.Placeholder = "Enter your memo content here..."
	contentArea.SetWidth(60)
	contentArea.SetHeight(10)
	contentArea.Focus()

	// 标签输入
	tagInput := textinput.New()
	tagInput.Placeholder = "tag1, tag2, tag3"

	return &MemoCreateModel{
		contentArea: contentArea,
		tagInput:    tagInput,
		focused:     0,
		visibility:  models.VisibilityPrivate,
		client:      c,
	}
}

// Init 初始化
func (m *MemoCreateModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update 更新
func (m *MemoCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, func() tea.Msg {
				return ViewChangeMsg{View: ViewMemoList}
			}

		case "tab":
			m.nextField()

		case "shift+tab":
			m.prevField()

		case "ctrl+v":
			// 切换可见性
			m.toggleVisibility()

		case "ctrl+s", "ctrl+enter":
			return m, m.createMemoCmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.contentArea.SetWidth(msg.Width - 10)
		m.contentArea.SetHeight(msg.Height - 15)

	case MemoCreatedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			m.errMessage = "Failed to create memo: " + msg.Err.Error()
		} else {
			m.success = true
			// 返回列表视图
			return m, func() tea.Msg {
				return ViewChangeMsg{View: ViewMemoList}
			}
		}
	}

	// 更新输入组件
	var cmd tea.Cmd
	if m.focused == 0 {
		m.contentArea, cmd = m.contentArea.Update(msg)
	} else {
		m.tagInput, cmd = m.tagInput.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View 渲染
func (m *MemoCreateModel) View() string {
	var b strings.Builder

	// 标题
	title := createTitleStyle.Render("✏️  Create New Memo")
	b.WriteString(title)
	b.WriteString("\n\n")

	// 内容输入区
	contentLabel := createLabelStyle.Render("Content:")
	b.WriteString(contentLabel)
	b.WriteString("\n")
	b.WriteString(m.contentArea.View())
	b.WriteString("\n\n")

	// 标签输入
	tagLabel := createLabelStyle.Render("Tags (comma separated):")
	b.WriteString(tagLabel)
	b.WriteString("\n")
	tagInputStyle := createInputStyle
	if m.focused == 1 {
		tagInputStyle = createFocusedInputStyle
	}
	b.WriteString(tagInputStyle.Render(m.tagInput.View()))
	b.WriteString("\n\n")

	// 可见性选择
	visLabel := createLabelStyle.Render("Visibility:")
	b.WriteString(visLabel)
	b.WriteString(" ")

	visOptions := []string{"PRIVATE", "PUBLIC", "PROTECTED"}
	for _, v := range visOptions {
		if models.Visibility(v) == m.visibility {
			b.WriteString(createSelectedStyle.Render("[" + v + "]"))
		} else {
			b.WriteString(createOptionStyle.Render(" " + v + " "))
		}
		b.WriteString(" ")
	}
	b.WriteString("\n\n")

	// 快捷键提示
	help := createHelpStyle.Render(
		"Tab: switch field • Ctrl+V: toggle visibility • Ctrl+S: save • Esc: cancel",
	)
	b.WriteString(help)
	b.WriteString("\n")

	// 错误信息
	if m.errMessage != "" {
		b.WriteString("\n")
		b.WriteString(createErrorStyle.Render("✗ " + m.errMessage))
	}

	// 加载状态
	if m.loading {
		b.WriteString("\n")
		b.WriteString(createLoadingStyle.Render("Creating memo..."))
	}

	return b.String()
}

// nextField 切换到下一个字段
func (m *MemoCreateModel) nextField() {
	m.focused = (m.focused + 1) % 2
	m.updateFocus()
}

// prevField 切换到上一个字段
func (m *MemoCreateModel) prevField() {
	m.focused = (m.focused - 1 + 2) % 2
	m.updateFocus()
}

// updateFocus 更新焦点
func (m *MemoCreateModel) updateFocus() {
	if m.focused == 0 {
		m.contentArea.Focus()
		m.tagInput.Blur()
	} else {
		m.contentArea.Blur()
		m.tagInput.Focus()
	}
}

// toggleVisibility 切换可见性
func (m *MemoCreateModel) toggleVisibility() {
	switch m.visibility {
	case models.VisibilityPrivate:
		m.visibility = models.VisibilityPublic
	case models.VisibilityPublic:
		m.visibility = models.VisibilityProtected
	default:
		m.visibility = models.VisibilityPrivate
	}
}

// createMemoCmd 创建 Memo 命令
func (m *MemoCreateModel) createMemoCmd() tea.Msg {
	content := m.contentArea.Value()
	if strings.TrimSpace(content) == "" {
		return MemoCreatedMsg{Err: fmt.Errorf("content is required")}
	}

	req := &models.CreateMemoRequest{
		Content:    content,
		Visibility: m.visibility,
	}

	memo, err := client.NewMemoService(m.client).Create(context.Background(), req)
	if err != nil {
		return MemoCreatedMsg{Err: err}
	}

	return MemoCreatedMsg{Memo: memo}
}

// MemoCreatedMsg Memo 创建消息
type MemoCreatedMsg struct {
	Memo *models.Memo
	Err  error
}

// 样式
var (
	createTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true).
			MarginBottom(1)

	createLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("7")).
				MarginTop(1)

	createInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("0")).
				Padding(0, 1).
				Width(50)

	createFocusedInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("4")).
				Padding(0, 1).
				Width(50)

	createOptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))

	createSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true)

	createHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				MarginTop(1)

	createErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9"))

	createLoadingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("11"))
)
