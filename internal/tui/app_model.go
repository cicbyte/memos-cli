package tui

// App TUI 应用包装
type App struct {
	Model
}

// NewApp 创建新应用
func NewApp() *App {
	return &App{
		Model: NewModel(),
	}
}
