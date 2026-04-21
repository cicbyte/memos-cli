// Package tui provides terminal user interface for memos-cli
package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// StartTUI 启动 TUI 应用
func StartTUI() error {
	// 检查是否在终端中运行
	if !isTerminal() {
		return fmt.Errorf("TUI requires a terminal. Please run in a terminal emulator")
	}

	app := NewApp()
	// 不使用 WithAltScreen() 以允许终端原生文本选择和复制
	p := tea.NewProgram(app)

	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}
	return nil
}

// isTerminal 检查是否在终端中运行
func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
