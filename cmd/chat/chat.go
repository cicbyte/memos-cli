package chat

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/cicbyte/memos-cli/internal/ai"
	"github.com/cicbyte/memos-cli/internal/common"
	chatlogic "github.com/cicbyte/memos-cli/internal/logic/chat"
	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	chatTag         []string
	chatVisibility  string
	chatLimit       int
	chatShowMemos   bool
	chatOutput      string
	chatInteractive bool
	chatMode        string
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Padding(0, 1)
	dividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	mutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	tagStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	userStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	promptStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	toolStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)

var mdRenderer, _ = glamour.NewTermRenderer(
	glamour.WithAutoStyle(),
	glamour.WithWordWrap(0),
)

func parseSearchMode(mode string) ai.SearchMode {
	switch mode {
	case "db":
		return ai.SearchDB
	case "vector":
		return ai.SearchVector
	default:
		return ai.SearchAuto
	}
}

func runChat(cmd *cobra.Command, args []string) {
	if !chatInteractive && len(args) == 0 {
		fmt.Println(errorStyle.Render("  请输入问题，或使用 --interactive 进入多轮对话模式"))
		cmd.Help()
		return
	}

	appConfig := common.GetAppConfig()
	config := &chatlogic.Config{
		Tags:       chatTag,
		Visibility: chatVisibility,
		Limit:      chatLimit,
		SearchMode: parseSearchMode(chatMode),
	}

	processor := chatlogic.NewProcessor(config, appConfig)
	ctx := context.Background()

	if chatInteractive {
		runInteractive(ctx, args, processor)
	} else {
		question := strings.Join(args, " ")
		runSingleTurn(ctx, question, processor)
	}
}

func runSingleTurn(ctx context.Context, question string, processor *chatlogic.Processor) {
	start := time.Now()
	var sources []*models.LocalMemo
	var answerBuf strings.Builder
	var rawLines int

	cb := func(event ai.StreamEvent) {
		handleStreamEvent(event, &sources, &answerBuf, &rawLines, start)
	}

	if err := processor.ExecuteStream(ctx, question, cb); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("\n  对话失败: %v", err)))
		os.Exit(1)
	}

	if chatOutput != "" {
		saveToFile(answerBuf.String(), sources, chatOutput)
	}
}

func runInteractive(ctx context.Context, args []string, processor *chatlogic.Processor) {
	sessionID := processor.NewSession()

	fmt.Println(titleStyle.Render(" 多轮对话模式"))
	fmt.Println(dividerStyle.Render("────────────────────────────────────────────────"))
	fmt.Println(mutedStyle.Render("  输入问题开始对话，输入 /quit 退出，/clear 清除上下文"))
	fmt.Println()

	if len(args) > 0 {
		question := strings.Join(args, " ")
		fmt.Println(userStyle.Render("  > " + question))
		streamTurn(ctx, processor, sessionID, question)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println()
		fmt.Print(promptStyle.Render("  user > "))
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}
		if input == "/quit" || input == "/exit" || input == "q" {
			fmt.Println(mutedStyle.Render("  再见!"))
			break
		}
		if input == "/clear" {
			processor.ClearSession(sessionID)
			fmt.Println(successStyle.Render("  对话上下文已清除"))
			continue
		}
		streamTurn(ctx, processor, sessionID, input)
	}
}

func streamTurn(ctx context.Context, processor *chatlogic.Processor, sessionID, question string) {
	start := time.Now()
	var sources []*models.LocalMemo
	var answerBuf strings.Builder
	var rawLines int

	cb := func(event ai.StreamEvent) {
		handleStreamEvent(event, &sources, &answerBuf, &rawLines, start)
	}

	if err := processor.ExecuteWithSessionStream(ctx, sessionID, question, cb); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("\n  对话失败: %v", err)))
	}
}

func handleStreamEvent(event ai.StreamEvent, sources *[]*models.LocalMemo, answerBuf *strings.Builder, rawLines *int, start time.Time) {
	switch event.Type {
	case "tool_call":
		fmt.Println(toolStyle.Render(fmt.Sprintf("  > %s: %s", toolDisplayName(event.Tool), truncateArgs(event.Content))))
	case "tool_result":
		fmt.Println(successStyle.Render(fmt.Sprintf("  %s", event.Content)))
	case "content":
		answerBuf.WriteString(event.Content)
		fmt.Print(event.Content)
		for _, c := range event.Content {
			if c == '\n' {
				*rawLines++
			}
		}
	case "done":
		*sources = event.Sources
		if answerBuf.Len() > 0 {
			moveUpAndClear(*rawLines)
			rendered, err := mdRenderer.Render(answerBuf.String())
			if err != nil {
				fmt.Print(answerBuf.String())
			} else {
				fmt.Print(rendered)
			}
			fmt.Println()
		}
		printSources(event.Sources)
		if event.PromptTokens > 0 || event.CompletionTokens > 0 {
			total := event.PromptTokens + event.CompletionTokens
			elapsed := time.Since(start)
			fmt.Println(mutedStyle.Render(fmt.Sprintf("  Token: %d + %d = %d · %.1fs",
				event.PromptTokens, event.CompletionTokens, total, elapsed.Seconds())))
		}
	case "error":
		fmt.Println(errorStyle.Render(fmt.Sprintf("  %s", event.Content)))
	}
}

func toolDisplayName(name string) string {
	switch name {
	case "search_memos":
		return "search_memos"
	case "semantic_search":
		return "semantic_search"
	case "get_memo":
		return "get_memo"
	default:
		return name
	}
}

func truncateArgs(args string) string {
	if len(args) > 80 {
		return args[:80] + "..."
	}
	return args
}

func printSources(sources []*models.LocalMemo) {
	if len(sources) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(titleStyle.Render(fmt.Sprintf(" 参考来源 (%d条)", len(sources))))
	fmt.Println(dividerStyle.Render("────────────────────────────────────────────────"))

	for i, source := range sources {
		content := strings.TrimSpace(source.Content)
		if idx := strings.IndexByte(content, '\n'); idx > 0 {
			content = content[:idx]
		}
		if len(content) > 60 {
			content = content[:60] + "..."
		}

		created := time.Unix(source.CreatedTime, 0).Format("2006-01-02 15:04")
		fmt.Printf("  %d. %s  %s\n", i+1, labelStyle.Render(source.UID), mutedStyle.Render(created))
		if content != "" {
			fmt.Printf("     %s\n", mutedStyle.Render(content))
		}
	}
}

func moveUpAndClear(lines int) {
	if lines <= 0 {
		return
	}
	fmt.Printf("\033[%dA", lines)   // move cursor up
	fmt.Printf("\033[J")             // clear from cursor to end of screen
}

func saveToFile(answer string, sources []*models.LocalMemo, filename string) {
	content := fmt.Sprintf("# AI 对话记录\n\n%s\n\n", answer)

	if len(sources) > 0 {
		content += "## 来源备忘录\n\n"
		for i, source := range sources {
			content += fmt.Sprintf("### 备忘录 #%d\n", i+1)
			content += fmt.Sprintf("- **UID**: %s\n", source.UID)
			content += fmt.Sprintf("- **时间**: %d\n", source.CreatedTime)
			if source.Property != "" {
				content += fmt.Sprintf("- **标签**: %s\n", source.Property)
			}
			content += fmt.Sprintf("- **内容**:\n\n%s\n\n", source.Content)
		}
	}

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("  保存失败: %v", err)))
	} else {
		fmt.Println(successStyle.Render(fmt.Sprintf("  已保存到: %s", filename)))
	}
}
