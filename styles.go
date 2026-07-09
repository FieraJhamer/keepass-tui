package main

import "github.com/charmbracelet/lipgloss"

var (
	borderColor = lipgloss.Color("#383838")
	accentColor = lipgloss.Color("#5f5fff")
	textColor   = lipgloss.Color("#ffffff")
	grayColor   = lipgloss.Color("#888888")
	greenColor  = lipgloss.Color("#5fff87")
	redColor    = lipgloss.Color("#ff5f5f")
	yellowColor = lipgloss.Color("#ffff5f")

	docStyle = lipgloss.NewStyle().Padding(1, 2)

	sidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

	contentStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Background(borderColor).
			Foreground(grayColor).
			Padding(0, 1)

	unlockStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(accentColor).
			Padding(1, 3)

	formStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(accentColor).
			Padding(1, 2)

	labelStyle = lipgloss.NewStyle().
			Foreground(grayColor).
			Width(10)

	helpStyle = lipgloss.NewStyle().
			Foreground(grayColor).
			Italic(true)

	confirmStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(yellowColor).
			Padding(1, 2).
			Foreground(textColor)
)
