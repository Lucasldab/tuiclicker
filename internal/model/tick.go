package model

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TickMsg is sent once per second by doTick. Update() re-issues doTick on
// every TickMsg to keep the loop alive. Never use time.Sleep.
type TickMsg time.Time

// doTick returns a Cmd that fires a TickMsg after one second.
// Re-issue this command from Update() on every TickMsg received.
func doTick() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// clearFlashMsg is sent immediately after a harvest click to clear the
// flash highlight on the next render, achieving a 1-render visual burst.
type clearFlashMsg struct{}
