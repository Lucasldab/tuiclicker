package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lucasldab/tuiclicker/internal/model"
	"github.com/lucasldab/tuiclicker/internal/persistence"
)

func main() {
	m := loadOrNew()
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // CellMotion, not AllMotion — prevents CPU burn (Pitfall 1)
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// loadOrNew attempts to load a save file and apply offline progress.
// Falls back to model.New() on first launch or if the save is corrupt.
// ApplyOfflineProgress is called on SaveData (before ToGameModel) so that
// offline credit uses production rates at time-of-save, not recalculated rates.
func loadOrNew() model.GameModel {
	path, err := persistence.SavePath()
	if err != nil {
		return model.New()
	}
	sd, found, err := persistence.Load(path)
	if !found || err != nil {
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not load save: %v\n", err)
		}
		return model.New()
	}
	persistence.ApplyOfflineProgress(&sd, time.Now())
	return model.FromSaveData(sd)
}
