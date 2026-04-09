package ui

import "github.com/charmbracelet/lipgloss"

// All Lipgloss style objects live here. Import this package from panels.go only.
// Color numbers are 16-color ANSI per UI-SPEC — no 256-color, no 24-bit RGB.

var (
	// Text roles
	StylePrimary = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))  // white
	StyleDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // dark gray

	// Resource accent colors
	StyleBlood = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))  // ANSI red
	StyleFlesh = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))  // bright red
	StyleBones = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3"))  // yellow

	// Harvest flash
	StyleFlash = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")) // bright white

	// Tab bar
	StyleTabActive   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	StyleTabInactive = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// Headings: bold + resource color — compose with resource styles above
	StyleHeading = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7"))
)
