package main

import "github.com/charmbracelet/lipgloss"

var (
	// Paleta inspirada en macOS: azul de sistema para lo activo, grises para
	// la estructura y la segunda linea de texto, rojo y naranja para avisos.
	borderColor = lipgloss.Color("#2D2D30") // gris de panel, tambien el fondo del footer
	accentColor = lipgloss.Color("#0091FF") // azul de sistema, cursor y foco
	textColor   = lipgloss.Color("#FFFFFF") // texto de alta prioridad
	grayColor   = lipgloss.Color("#A1A1AA") // etiquetas y hints
	redColor    = lipgloss.Color("#FF453A") // rojo de sistema, errores
	yellowColor = lipgloss.Color("#FF9F0A") // naranja de sistema, advertencias

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
