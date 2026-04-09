package ui

import "github.com/charmbracelet/lipgloss"

// 16-color ANSI palette per UI-SPEC. No 256-color, no 24-bit RGB.

var (
	StylePrimary = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	StyleDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	StyleBlood = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	StyleFlesh = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	StyleBones = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3"))
	StyleFlash = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))

	StyleTabActive   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	StyleTabInactive = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	StyleHeading     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7"))
)
