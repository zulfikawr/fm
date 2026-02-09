package app

import (
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TickMsg is sent on each tick for animations
type TickMsg struct {
	Time time.Time
}

// RAMUpdateMsg is sent periodically to update RAM usage
type RAMUpdateMsg struct {
	RAMUsageMB uint64
}

// StartRAMTicker starts a ticker that sends RAM usage updates every second
func StartRAMTicker() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		ramMB := m.Alloc / 1024 / 1024
		return RAMUpdateMsg{RAMUsageMB: ramMB}
	})
}
