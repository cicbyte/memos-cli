package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/cicbyte/memos-cli/internal/client"
	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MemoDetailModel Memo 详情视图模型
type MemoDetailModel struct {
	// 数据
	memo   *models.Memo
	client *client.Client

	// 视口
	viewport viewport.Model

	// 编辑器
	editor textarea.Model

	// 状态
	editing     bool
	loading     bool
	err         error
	deleteMode  bool
	copied      bool // 复制成功标志
	prevView    ViewType

	// 尺寸
	width  int
	height int
}

// NewMemoDetailModel 创建 Memo 详情模型
func NewMemoDetailModel(memo *models.Memo, client *client.Client, prevView ViewType) *MemoDetailModel {
	// 初始化 viewport，设置默认尺寸
	vp := viewport.New(80, 24)
	vp.SetContent("Loading...")

	m := &MemoDetailModel{
		memo:     memo,
		client:   client,
		viewport: vp,
			prevView: prevView,
	}

	// 初始化编辑器
	m.editor = textarea.New()
	m.editor.SetValue(memo.Content)
	m.editor.SetWidth(60)
	m.editor.SetHeight(15)

	// 渲染内容
	content := m.renderContent()
	m.viewport.SetContent(content)
	return m
}

// renderContent 渲染内容
func (m *MemoDetailModel) renderContent() string {
	var b strings.Builder

	// 标题
	id := m.memo.Uid
	if id == "" && m.memo.Name != "" {
		fmt.Sscanf(m.memo.Name, "memos/%s", &id)
	}

	b.WriteString(titleStyle.Render(fmt.Sprintf("Memo #%s", id)))
	b.WriteString("\n\n")

	// 元信息
	if m.memo.CreateTime != nil {
		b.WriteString(metaStyle.Render("Created: " + m.memo.CreateTime.Format("2006-01-02 15:04:05")))
		b.WriteString("\n")
	}

	if m.memo.UpdateTime != nil && !m.memo.UpdateTime.Equal(*m.memo.CreateTime) {
		b.WriteString(metaStyle.Render("Updated: " + m.memo.UpdateTime.Format("2006-01-02 15:04:05")))
		b.WriteString("\n")
	}

	// 可见性
	visStyle := publicStyle
	if m.memo.Visibility == models.VisibilityPrivate {
		visStyle = privateStyle
	}
	b.WriteString(metaStyle.Render("Visibility: ") + visStyle.Render(string(m.memo.Visibility)))
	b.WriteString("\n")

	// 置顶状态
	if m.memo.Pinned {
		b.WriteString(metaStyle.Render("Pinned: ") + pinStyle.Render("Yes"))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// 内容
	b.WriteString(contentStyle.Render(m.memo.Content))
	b.WriteString("\n\n")

	// 标签
	if m.memo.Property != nil && len(m.memo.Property.Tags) > 0 {
		b.WriteString(detailLabelStyle.Render("Tags:"))
		b.WriteString("\n")
		tags := ""
		for _, tag := range m.memo.Property.Tags {
			tags += "#" + tag + " "
		}
		b.WriteString(tagStyle.Render(tags))
		b.WriteString("\n\n")
	}

	// 资源
	if len(m.memo.Resources) > 0 {
		b.WriteString(detailLabelStyle.Render("Attachments:"))
		b.WriteString("\n")
		for _, res := range m.memo.Resources {
			b.WriteString(fmt.Sprintf("  • %s (%s)\n", res.Filename, res.Type))
		}
		b.WriteString("\n")
	}

	// 反应
	if len(m.memo.Reactions) > 0 {
		b.WriteString(detailLabelStyle.Render("Reactions:"))
		b.WriteString("\n")
		reactions := ""
		for _, r := range m.memo.Reactions {
			reactions += r.ReactionType + " "
		}
		b.WriteString(fmt.Sprintf("  %s\n", reactions))
		b.WriteString("\n")
	}

	// 帮助
	b.WriteString("\n")
	b.WriteString(detailHelpStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	b.WriteString("\n")
	if m.copied {
		b.WriteString(copySuccessStyle.Render("  ✓ Copied to clipboard!"))
		b.WriteString("\n")
	}
	b.WriteString(detailHelpStyle.Render("  c: Copy  •  e: Edit  •  d: Delete  •  Esc/q/b: Back"))

	return b.String()
}

// renderEditMode 渲染编辑模式
func (m *MemoDetailModel) renderEditMode() string {
	var b strings.Builder

	// 标题
	id := m.memo.Uid
	if id == "" && m.memo.Name != "" {
		fmt.Sscanf(m.memo.Name, "memos/%s", &id)
	}

	b.WriteString(editTitleStyle.Render(fmt.Sprintf("✏️  Edit Memo #%s", id)))
	b.WriteString("\n\n")

	// 可见性
	b.WriteString(editLabelStyle.Render("Visibility: "))
	visOptions := []string{"PRIVATE", "PUBLIC"}
	for _, v := range visOptions {
		if models.Visibility(v) == m.memo.Visibility {
			b.WriteString(editSelectedStyle.Render("[" + v + "]"))
		} else {
			b.WriteString(editOptionStyle.Render(" " + v + " "))
		}
		b.WriteString(" ")
	}
	b.WriteString("  ")
	b.WriteString(editHintStyle.Render("(Ctrl+V to toggle)"))
	b.WriteString("\n\n")

	// 内容编辑器
	b.WriteString(editLabelStyle.Render("Content:"))
	b.WriteString("\n")
	b.WriteString(m.editor.View())
	b.WriteString("\n\n")

	// 快捷键
	b.WriteString(editHelpStyle.Render("Ctrl+S: save • Ctrl+V: toggle visibility • Esc: cancel"))

	return b.String()
}

// renderDeleteMode 渲染删除确认
func (m *MemoDetailModel) renderDeleteMode() string {
	id := m.memo.Uid
	if id == "" && m.memo.Name != "" {
		fmt.Sscanf(m.memo.Name, "memos/%s", &id)
	}

	return deleteConfirmStyle.Render(fmt.Sprintf(
		"\n  ⚠️  Delete Memo #%s?\n\n  This action cannot be undone!\n\n  y: confirm • n/Esc: cancel\n",
		id,
	))
}

// Init 初始化
func (m *MemoDetailModel) Init() tea.Cmd {
	return nil
}

// Update 更新
func (m *MemoDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// 编辑模式下的处理
	if m.editing {
		return m.updateEditMode(msg)
	}

	// 删除确认模式
	if m.deleteMode {
		return m.updateDeleteMode(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "b":
			return m, func() tea.Msg {
				prev := m.prevView
					if prev == 0 {
						prev = ViewMemoList
					}
					return ViewChangeMsg{View: prev}
				}

		case "c":
			// 复制 memo 内容到剪贴板
			return m, m.copyMemoCmd

		case "e":
			m.editing = true
			m.editor.SetValue(m.memo.Content)
			m.editor.Focus()
			return m, textarea.Blink

		case "d":
			m.deleteMode = true
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 4
		m.editor.SetWidth(msg.Width - 10)
		m.editor.SetHeight(msg.Height - 15)

	case MemoUpdatedMsg:
		m.loading = false
		if msg.Err == nil {
			m.memo = msg.Memo
			m.editing = false
			m.viewport.SetContent(m.renderContent())
		}

	case MemoDeletedMsg:
		if msg.Err == nil {
			return m, func() tea.Msg {
				prev := m.prevView
					if prev == 0 {
						prev = ViewMemoList
					}
					return ViewChangeMsg{View: prev}
				}
		}

	case CopySuccessMsg:
		m.copied = true
		m.viewport.SetContent(m.renderContent())
		// 2秒后重置 copied 状态
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return ResetCopyMsg{}
		})

	case ResetCopyMsg:
		m.copied = false
		m.viewport.SetContent(m.renderContent())
	}

	// 更新视口
	vp, cmd := m.viewport.Update(msg)
	m.viewport = vp
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// updateEditMode 编辑模式更新
func (m *MemoDetailModel) updateEditMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.editing = false
			return m, nil

		case "ctrl+s":
			return m, m.saveMemoCmd

		case "ctrl+v":
			// 切换可见性
			if m.memo.Visibility == models.VisibilityPrivate {
				m.memo.Visibility = models.VisibilityPublic
			} else {
				m.memo.Visibility = models.VisibilityPrivate
			}
			return m, nil
		}
	}

	// 更新编辑器
	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	return m, cmd
}

// updateDeleteMode 删除模式更新
func (m *MemoDetailModel) updateDeleteMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.deleteMode = false
			return m, m.deleteMemoCmd

		case "n", "N", "esc":
			m.deleteMode = false
			return m, nil
		}
	}

	return m, nil
}

// saveMemoCmd 保存 Memo 命令
func (m *MemoDetailModel) saveMemoCmd() tea.Msg {
	content := m.editor.Value()

	// 提取 memo ID
	id := m.memo.Uid
	if id == "" && m.memo.Name != "" {
		fmt.Sscanf(m.memo.Name, "memos/%s", &id)
	}

	req := &models.UpdateMemoRequest{
		Content:    &content,
		Visibility: &m.memo.Visibility,
	}
	req.UpdateMask = "content,visibility"

	updated, err := client.NewMemoService(m.client).Update(context.Background(), id, req)
	if err != nil {
		return MemoUpdatedMsg{Err: err}
	}

	return MemoUpdatedMsg{Memo: updated}
}

// deleteMemoCmd 删除 Memo 命令
func (m *MemoDetailModel) deleteMemoCmd() tea.Msg {
	// 提取 memo ID
	id := m.memo.Uid
	if id == "" && m.memo.Name != "" {
		fmt.Sscanf(m.memo.Name, "memos/%s", &id)
	}

	err := client.NewMemoService(m.client).Delete(context.Background(), id)
	return MemoDeletedMsg{Err: err}
}

// copyMemoCmd 复制 Memo 内容到剪贴板
func (m *MemoDetailModel) copyMemoCmd() tea.Msg {
	// 只复制内容，不复制元信息
	err := clipboard.WriteAll(m.memo.Content)
	if err != nil {
		return CopyErrorMsg{Err: err}
	}

	return CopySuccessMsg{}
}

// View 渲染
func (m *MemoDetailModel) View() string {
	if m.loading {
		return "\n  Saving..."
	}

	if m.deleteMode {
		return m.renderDeleteMode()
	}

	if m.editing {
		return m.renderEditMode()
	}

	// 直接渲染内容，不使用 viewport
	return m.renderContent()
}

// 消息类型
type MemoUpdatedMsg struct {
	Memo *models.Memo
	Err  error
}

type MemoDeletedMsg struct {
	Err error
}

type CopySuccessMsg struct{}

type CopyErrorMsg struct {
	Err error
}

type ResetCopyMsg struct{}

// 样式
var (
	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	contentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Padding(1, 0)

	detailLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true)

	publicStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	privateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))

	pinStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))

	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6"))

	detailBoxStyle = lipgloss.NewStyle().
			Padding(1, 2)

	detailHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	copySuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true)

	// 编辑模式样式
	editTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true).
			MarginBottom(1)

	editLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	editOptionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	editSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true)

	editHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	editHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			MarginTop(1)

	// 删除确认样式
	deleteConfirmStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("9")).
				Padding(1, 2)
)
