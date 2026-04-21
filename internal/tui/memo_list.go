package tui

import (
	"fmt"
	"strings"

	"github.com/cicbyte/memos-cli/internal/client"
	"github.com/cicbyte/memos-cli/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// MemoListModel Memo 列表视图模型
type MemoListModel struct {
	memos         []*models.Memo
	filteredMemos []*models.Memo
	loading       bool
	err           error

	// 分页
	currentPage int
	pageSize    int
	totalPages  int

	// 光标
	cursor int

	// 搜索/过滤
	filter string

	// API 客户端
	client *client.Client

	// 尺寸
	width  int
	height int

	// 预计算的样式（避免重复创建）
	headerStyle    lipgloss.Style
	rowStyle       lipgloss.Style
	selectedStyle  lipgloss.Style
	altRowStyle    lipgloss.Style
}

// NewMemoListModel 创建 Memo 列表模型
func NewMemoListModel() *MemoListModel {
	return &MemoListModel{
		pageSize:    10,
		currentPage: 1,
		cursor:      0,
		headerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("6")),
		rowStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")),
		selectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("4")).
			Bold(true),
		altRowStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")),
	}
}

// SetClient 设置客户端
func (m *MemoListModel) SetClient(c *client.Client) {
	m.client = c
}

// SetMemos 设置备忘录列表
func (m *MemoListModel) SetMemos(memos []*models.Memo) {
	m.memos = memos
	m.applyFilter()
}

// applyFilter 应用过滤并更新表格
func (m *MemoListModel) applyFilter() {
	// 过滤
	if m.filter == "" {
		m.filteredMemos = m.memos
	} else {
		m.filteredMemos = make([]*models.Memo, 0)
		filterLower := strings.ToLower(m.filter)
		for _, memo := range m.memos {
			if strings.Contains(strings.ToLower(memo.Content), filterLower) {
				m.filteredMemos = append(m.filteredMemos, memo)
			}
		}
	}

	// 计算总页数
	m.totalPages = (len(m.filteredMemos) + m.pageSize - 1) / m.pageSize
	if m.totalPages < 1 {
		m.totalPages = 1
	}

	// 确保当前页在有效范围内
	if m.currentPage > m.totalPages {
		m.currentPage = m.totalPages
	}
	if m.currentPage < 1 {
		m.currentPage = 1
	}

	// 重置光标
	m.cursor = 0
}

// getPageItems 获取当前页的数据
func (m *MemoListModel) getPageItems() []*models.Memo {
	startIdx := (m.currentPage - 1) * m.pageSize
	endIdx := startIdx + m.pageSize
	if endIdx > len(m.filteredMemos) {
		endIdx = len(m.filteredMemos)
	}
	if startIdx >= len(m.filteredMemos) {
		return []*models.Memo{}
	}
	return m.filteredMemos[startIdx:endIdx]
}

// getSelectedMemo 获取当前选中的 Memo
func (m *MemoListModel) getSelectedMemo() *models.Memo {
	items := m.getPageItems()
	if m.cursor >= 0 && m.cursor < len(items) {
		return items[m.cursor]
	}
	return nil
}

// moveCursor 移动光标
func (m *MemoListModel) moveCursor(delta int) {
	items := m.getPageItems()
	newCursor := m.cursor + delta
	if newCursor < 0 {
		newCursor = 0
	}
	if newCursor >= len(items) {
		newCursor = len(items) - 1
	}
	if newCursor < 0 {
		newCursor = 0
	}
	m.cursor = newCursor
}

// nextPage 下一页
func (m *MemoListModel) nextPage() {
	if m.currentPage < m.totalPages {
		m.currentPage++
		m.cursor = 0
	}
}

// prevPage 上一页
func (m *MemoListModel) prevPage() {
	if m.currentPage > 1 {
		m.currentPage--
		m.cursor = 0
	}
}

// Init 初始化
func (m *MemoListModel) Init() tea.Cmd {
	return nil
}

// Update 更新
func (m *MemoListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if memo := m.getSelectedMemo(); memo != nil {
				return m, func() tea.Msg {
					return ViewChangeMsg{
						View: ViewMemoDetail,
						Memo: memo,
					}
				}
			}

		case "n":
			return m, func() tea.Msg {
				return ViewChangeMsg{View: ViewMemoCreate}
			}

		case "r":
			if m.client != nil {
				m.loading = true
				return m, loadMemosCmd(m.client)
			}

		case "s":
			return m, func() tea.Msg {
				return ViewChangeMsg{View: ViewSettings}
			}

		case "/":
			return m, func() tea.Msg {
				return ViewChangeMsg{View: ViewMemoSearch}
			}

		case "up", "k":
			m.moveCursor(-1)

		case "down", "j":
			m.moveCursor(1)

		case "h", "left":
			m.prevPage()

		case "l", "right":
			m.nextPage()

		case "pgup":
			m.moveCursor(-5)

		case "pgdown":
			m.moveCursor(5)

		case "g", "home":
			m.currentPage = 1
			m.cursor = 0

		case "G", "end":
			m.currentPage = m.totalPages
			m.cursor = 0

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			page := int(msg.Runes[0] - '0')
			if page <= m.totalPages {
				m.currentPage = page
				m.cursor = 0
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// 计算每页显示行数（留出标题(1)、表头(1)、分隔线(1)、分页信息(1)、帮助(1)、状态栏(1)的空间）
		// 总固定开销约 6 行
		m.pageSize = msg.Height - 6
		if m.pageSize < 3 {
			m.pageSize = 3
		}
		if m.pageSize > 20 {
			m.pageSize = 20 // 限制最大行数，避免闪烁
		}
		// 重新计算分页
		m.applyFilter()

	case MemosLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.currentPage = 1
			m.SetMemos(msg.Memos)
		}
	}

	return m, nil
}

// View 渲染
func (m *MemoListModel) View() string {
	if m.loading {
		return m.renderFixedHeight("  加载中...")
	}

	if m.err != nil {
		return m.renderFixedHeight(errorStyle.Render(fmt.Sprintf("  加载失败: %v", m.err)))
	}

	if len(m.memos) == 0 {
		return m.renderFixedHeight(emptyStyle.Render(`  暂无备忘录

  按 'n' 创建新备忘录
  按 'r' 刷新列表`))
	}

	if len(m.filteredMemos) == 0 {
		return m.renderFixedHeight(emptyStyle.Render(`  没有找到匹配的备忘录

  按 '/' 搜索
  按 'r' 刷新列表`))
	}

	var b strings.Builder

	// 标题
	title := listTitleStyle.Render(" 📋 Memos 列表 ")
	b.WriteString(title)
	b.WriteString("\n")

	// 渲染表格
	b.WriteString(m.renderTable())

	// 分页信息
	pagination := fmt.Sprintf("  第 %d/%d 页 | 共 %d 条 ", m.currentPage, m.totalPages, len(m.filteredMemos))
	if m.filter != "" {
		pagination += fmt.Sprintf("(过滤: %s)", m.filter)
	}
	b.WriteString(paginationStyle.Render(pagination))
	b.WriteString("\n")

	// 帮助提示
	b.WriteString("  ↑/↓导航 ←/→翻页 Enter查看 n新建 /搜索 r刷新")

	return b.String()
}

// getFixedOutputHeight 返回固定的输出行数
func (m *MemoListModel) getFixedOutputHeight() int {
	// 标题(1) + 表头(1) + 分隔线(1) + 数据行(pageSize) + 分页(1) + 帮助(1) = pageSize + 5
	return m.pageSize + 5
}

// renderFixedHeight 确保输出固定高度，避免闪烁
func (m *MemoListModel) renderFixedHeight(content string) string {
	lines := strings.Split(content, "\n")
	var b strings.Builder
	
	// 标题行
	title := listTitleStyle.Render(" 📋 Memos 列表 ")
	b.WriteString(title)
	b.WriteString("\n")
	
	// 内容行（填充到固定高度）
	fixedHeight := m.getFixedOutputHeight() - 2 // 减去标题和底部信息
	for i := 0; i < fixedHeight; i++ {
		if i < len(lines) {
			b.WriteString(lines[i])
		} else {
			// 填充空行
			b.WriteString(strings.Repeat(" ", m.width-1))
		}
		b.WriteString("\n")
	}
	
	// 底部信息
	b.WriteString(paginationStyle.Render("  加载中..."))
	b.WriteString("\n")
	b.WriteString("  ↑/↓导航 ←/→翻页 Enter查看 n新建 /搜索 r刷新")
	
	return b.String()
}

// renderTable 渲染表格
func (m *MemoListModel) renderTable() string {
	items := m.getPageItems()
	
	// 计算列宽 - 限制内容展示，给时间留足空间
	idWidth := 12
	timeWidth := 16 // 足够显示完整日期时间: 2006-01-02 15:04
	minContentWidth := 25
	maxContentWidth := m.width
	
	// 内容宽度 = 总宽度 - ID - 时间 - 边距(4)
	contentWidth := m.width - idWidth - timeWidth - 4
	if contentWidth < minContentWidth {
		contentWidth = minContentWidth
	}
	if contentWidth > maxContentWidth {
		contentWidth = maxContentWidth
	}

	var b strings.Builder

	// 渲染表头
	header := fmt.Sprintf(" %-*s %-*s %-*s",
		idWidth, "ID",
		contentWidth, "内容预览",
		timeWidth, "时间")
	b.WriteString(m.headerStyle.Render(truncateString(header, m.width-2)))
	b.WriteString("\n")

	// 渲染分隔线
	sepLine := strings.Repeat("─", m.width-2)
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(sepLine))
	b.WriteString("\n")

	// 渲染行（固定行数，不足用空行填充）
	for i := 0; i < m.pageSize; i++ {
		if i < len(items) {
			row := m.formatRow(items[i], idWidth, contentWidth)
			// 选中行使用特殊样式
			if i == m.cursor {
				b.WriteString(m.selectedStyle.Render(truncateString(row, m.width-2)))
			} else {
				b.WriteString(m.rowStyle.Render(truncateString(row, m.width-2)))
			}
		} else {
			// 空行占位
			b.WriteString(strings.Repeat(" ", m.width-2))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// formatRow 格式化行数据
func (m *MemoListModel) formatRow(memo *models.Memo, idWidth, contentWidth int) string {
	// ID
	id := memo.Uid
	if id == "" && memo.Name != "" {
		fmt.Sscanf(memo.Name, "memos/%s", &id)
	}
	if len(id) > idWidth-2 {
		id = id[:idWidth-2]
	}

	// 置顶标记
	pin := "  "
	if memo.Pinned {
		pin = "📌"
	}
	idStr := pin + id

	// 内容预览 - 严格限制宽度
	content := memo.Content
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\t", " ")
	content = strings.TrimSpace(content)
	
	// 添加可见性标记
	visMark := ""
	if memo.Visibility == models.VisibilityPrivate {
		visMark = "🔒"
	} else if memo.Visibility == models.VisibilityProtected {
		visMark = "🔐"
	} else {
		visMark = "🔓"
	}
	
	content = visMark + " " + content
	// 严格限制内容长度，确保不占用时间列空间
	if runewidth.StringWidth(content) > contentWidth-1 {
		content = truncateString(content, contentWidth-1)
	}
	// 用空格填充到固定宽度，保持列对齐
	contentDisplayWidth := runewidth.StringWidth(content)
	if contentDisplayWidth < contentWidth {
		content += strings.Repeat(" ", contentWidth-contentDisplayWidth)
	}

	// 时间 - 显示完整日期时间
	timeWidth := 16
	var timeStr string
	if memo.DisplayTime != nil {
		timeStr = memo.DisplayTime.Format("2006-01-02 15:04")
	} else if memo.CreateTime != nil {
		timeStr = memo.CreateTime.Format("2006-01-02 15:04")
	}
	// 截断或填充时间列
	if runewidth.StringWidth(timeStr) > timeWidth {
		timeStr = truncateString(timeStr, timeWidth)
	} else {
		timeDisplayWidth := runewidth.StringWidth(timeStr)
		if timeDisplayWidth < timeWidth {
			timeStr += strings.Repeat(" ", timeWidth-timeDisplayWidth)
		}
	}

	return fmt.Sprintf(" %-*s %s %s",
		idWidth, idStr,
		content,
		timeStr)
}

// truncateString 截断字符串到指定显示宽度
func truncateString(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := runewidth.StringWidth(s)
	if w <= width {
		return s
	}
	
	// 从后往前找到截断位置
	runes := []rune(s)
	currentWidth := 0
	for i, r := range runes {
		rw := runewidth.RuneWidth(r)
		if currentWidth+rw > width-3 { // 预留 "..." 的位置
			return string(runes[:i]) + "..."
		}
		currentWidth += rw
	}
	return s
}

// 样式
var (
	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Padding(2, 4)

	listTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("6")).
			Bold(true).
			Padding(0, 1)

	paginationStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			Padding(0, 1)
)
