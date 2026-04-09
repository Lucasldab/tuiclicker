package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BuildLayout assembles the full terminal UI string from a GameView.
// Returns the too-small error string immediately when v.Width < 60 or v.Height < 20.
//
// Layout (per UI-SPEC):
//
//	row 0:    tab bar
//	row 1:    separator line
//	rows 2-N: three-column content area
//	row N+1:  separator line
//	row N+2:  status bar
//
// Column split: left 30% | center fills remainder | right 30%
// Column widths computed from v.Width — never hardcoded (Anti-Pattern 1).
func BuildLayout(v GameView) string {
	if v.Width < 60 || v.Height < 20 {
		return "terminal too small (min 60x20)"
	}

	leftW := (v.Width * 30) / 100
	rightW := (v.Width * 30) / 100
	centerW := v.Width - leftW - rightW

	// contentH: total rows minus tab bar (2 rows: bar + separator) and
	// status area (2 rows: separator + bar) = height - 4.
	contentH := v.Height - 4
	if contentH < 1 {
		contentH = 1
	}

	leftPanel := renderResourcePanel(v, leftW, contentH)
	centerPanel := renderCreaturePanel(v, centerW, contentH)
	var rightPanel string
	switch v.ActiveTab {
	case TabMutations:
		rightPanel = renderMutationPanel(v, rightW, contentH)
	case TabHarvesters:
		rightPanel = renderHarvesterPanel(v, rightW, contentH)
	default: // TabZones and any future tabs fall through to zones
		rightPanel = renderZonePanel(v, rightW, contentH)
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, centerPanel, rightPanel)
	tabBar := renderTabBar(v, v.Width)
	separator := strings.Repeat("-", v.Width)
	statusBar := renderStatusBar(v.Width)

	rows := []string{tabBar, separator, body, separator}
	if v.OfflineCreditMsg != "" {
		rows = append(rows, renderOfflineBanner(v.OfflineCreditMsg, v.Width))
	}
	rows = append(rows, statusBar)
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
