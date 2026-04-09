package model

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lucasldab/tuiclicker/internal/persistence"
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

// ClearFlashMsg returns a clearFlashMsg for test injection.
func ClearFlashMsg() tea.Msg { return clearFlashMsg{} }

// clearMutationFlashMsg clears the mutation purchase flash on the next render.
type clearMutationFlashMsg struct{}

// clearHarvesterFlashMsg clears the harvester purchase flash on the next render.
type clearHarvesterFlashMsg struct{}

// SaveResultMsg is returned by saveCmd after the async save completes.
// Errors are silently ignored — the game must never crash on a save failure.
type SaveResultMsg struct{ Err error }

// saveCmd dispatches a non-blocking save as a tea.Cmd goroutine.
// Called from the TickMsg handler. Never call persistence.Save synchronously
// inside Update() — file I/O would block the Bubbletea event loop (Pitfall: ARCHITECTURE.md Pattern 5).
func saveCmd(m GameModel) tea.Cmd {
	return func() tea.Msg {
		sd := ToSaveData(m) // ToSaveData is defined in save.go (same package)
		path, err := persistence.SavePath()
		if err != nil {
			return SaveResultMsg{Err: err}
		}
		return SaveResultMsg{Err: persistence.Save(sd, path)}
	}
}
