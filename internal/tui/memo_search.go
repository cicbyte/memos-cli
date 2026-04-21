package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/cicbyte/memos-cli/internal/client"
	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/cicbyte/memos-cli/internal/utils"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MemoSearchModel Memo 搜索视图模型
type MemoSearchModel struct {
	// 搜索输入
	searchInput textinput.Model

	// 结果列表
	list    list.Model
	results []*models.Memo
	loading bool
	err     error

	// API 客户端
	client *client.Client

	// 尺寸
	width  int
	height int
}

// searchResultItem 搜索结果项
type searchResultItem struct {
	memo   *models.Memo
	query  string
}

func (i searchResultItem) FilterValue() string {
	return i.memo.Content
}

func (i searchResultItem) Title() string {
	id := i.memo.Uid
	if id == "" && i.memo.Name != "" {
		fmt.Sscanf(i.memo.Name, "memos/%s", &id)
	}

	vis := "🔓"
	if i.memo.Visibility == models.VisibilityPrivate {
		vis = "🔒"
	}

	return fmt.Sprintf("#%s %s", id, vis)
}

func (i searchResultItem) Description() string {
	content := i.memo.Content
	// 高亮搜索词
	if i.query != "" {
		// 简单地截取包含搜索词的部分
		lowerContent := strings.ToLower(content)
		lowerQuery := strings.ToLower(i.query)
		if idx := strings.Index(lowerContent, lowerQuery); idx != -1 {
			start := idx - 20
			if start < 0 {
				start = 0
			}
			end := idx + len(i.query) + 30
			if end > len(content) {
				end = len(content)
			}
			content = "..." + content[start:end] + "..."
		}
	}

	if len(content) > 80 {
		content = content[:77] + "..."
	}
	content = strings.ReplaceAll(content, "\n", " ")

	return content
}

// NewMemoSearchModel 创建 Memo 搜索模型
func NewMemoSearchModel(c *client.Client) *MemoSearchModel {
	// 搜索输入框
	searchInput := textinput.New()
	searchInput.Placeholder = "Search memos..."
	searchInput.CharLimit = 100
	searchInput.Focus()

	// 创建列表
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("5")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("5"))

	l := list.New([]list.Item{}, delegate, 80, 20)
	l.Title = "Search Results"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("5")).
		Padding(0, 1)

	return &MemoSearchModel{
		searchInput: searchInput,
		list:        l,
		client:      c,
	}
}

// Init 初始化
func (m *MemoSearchModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update 更新
func (m *MemoSearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			// 直接返回列表，不区分输入框状态
			return m, func() tea.Msg {
				return ViewChangeMsg{View: ViewMemoList}
			}

		case "enter":
			if m.searchInput.Focused() {
				query := m.searchInput.Value()
				if query != "" {
					m.loading = true
					return m, m.searchCmd(query)
				}
			} else {
				// 打开选中的 Memo
				if item, ok := m.list.SelectedItem().(searchResultItem); ok {
					return m, func() tea.Msg {
						return ViewChangeMsg{
							View:     ViewMemoDetail,
									PrevView: ViewMemoSearch,
							Memo: item.memo,
						}
					}
				}
			}

		case "/":
			if !m.searchInput.Focused() {
				m.searchInput.Focus()
				return m, textinput.Blink
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-6)

	case SearchResultsMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.results = msg.Results
			items := make([]list.Item, len(msg.Results))
			for i, memo := range msg.Results {
				items[i] = searchResultItem{
					memo:  memo,
					query: m.searchInput.Value(),
				}
			}
			m.list.SetItems(items)
			// 搜索完成后取消输入框焦点，让用户可以操作结果列表
			m.searchInput.Blur()
		}
	}

	// 更新搜索输入框
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	cmds = append(cmds, cmd)

	// 如果不在输入模式，更新列表
	if !m.searchInput.Focused() {
		newList, cmd := m.list.Update(msg)
		m.list = newList
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View 渲染
func (m *MemoSearchModel) View() string {
	var b strings.Builder

	// 搜索框
	searchBox := searchBoxStyle.Render(m.searchInput.View())
	b.WriteString(searchBox)
	b.WriteString("\n\n")

	// 加载状态
	if m.loading {
		b.WriteString(searchLoadingStyle.Render("Searching..."))
		return b.String()
	}

	// 错误
	if m.err != nil {
		b.WriteString(searchErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		return b.String()
	}

	// 结果列表
	if len(m.results) > 0 {
		b.WriteString(m.list.View())
	} else if m.searchInput.Value() != "" {
		b.WriteString(searchEmptyStyle.Render("No results found."))
	}

	return b.String()
}

// searchCmd 搜索命令
func (m *MemoSearchModel) searchCmd(query string) tea.Cmd {
	return func() tea.Msg {
		// 从本地数据库搜索
		db, err := utils.GetGormDB()
		if err != nil {
			return SearchResultsMsg{Err: err}
		}

		// 使用 LIKE 进行文本搜索
		var localMemos []models.LocalMemo
		searchPattern := "%" + query + "%"
		if err := db.Where("is_deleted = ? AND content LIKE ?", false, searchPattern).
			Order("created_time DESC").
			Limit(50).
			Find(&localMemos).Error; err != nil {
			return SearchResultsMsg{Err: err}
		}

		// 转换为远程模型格式
		results := make([]*models.Memo, 0, len(localMemos))
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

			results = append(results, memo)
		}

		return SearchResultsMsg{Results: results}
	}
}

// SearchResultsMsg 搜索结果消息
type SearchResultsMsg struct {
	Results []*models.Memo
	Err     error
}

// 样式
var (
	searchBoxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("5")).
			Padding(0, 1).
			Width(60)

	searchLoadingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("11")).
				Padding(1, 2)

	searchErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")).
				Padding(1, 2)

	searchEmptyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Padding(2, 4)
)
