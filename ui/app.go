package ui

import tea "github.com/charmbracelet/bubbletea"

// App wraps the Bubble Tea program so the collector goroutine can push snapshots in.
type App struct {
	program *tea.Program
}

// NewApp creates the Bubble Tea program with the alternate screen.
func NewApp() *App {
	p := tea.NewProgram(newModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	return &App{program: p}
}

// Run starts the TUI event loop and blocks until the user quits.
func (a *App) Run() error {
	_, err := a.program.Run()
	return err
}

// Send delivers a fresh data snapshot into the TUI from any goroutine.
func (a *App) Send(snap Snapshot) {
	a.program.Send(snapMsg(snap))
}

// Stop quits the TUI programmatically (e.g. from a signal handler).
func (a *App) Stop() {
	a.program.Quit()
}
