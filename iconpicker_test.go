package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// The grid must fill the frame and never overflow the terminal.
func TestIconPickerFitsTheTerminalWidth(t *testing.T) {
	for _, width := range []int{60, 70, 100, 140} {
		m := searchModel(t)
		m.width = width
		m.height = 30
		m.iconPickMode = true
		m.iconPickTarget = "entry"

		view := m.iconPickView()
		for _, line := range strings.Split(view, "\n") {
			if w := lipgloss.Width(line); w > width {
				t.Fatalf("at width %d a line is %d columns wide", width, w)
			}
		}
	}
}

// The number of columns follows the terminal width.
func TestIconPickerColumnsFollowTheTerminal(t *testing.T) {
	m := searchModel(t)

	m.width = 100
	if got := m.iconPickerCols(); got != 11 {
		t.Fatalf("at width 100 the grid has %d columns, want 11", got)
	}

	m.width = 70
	if got := m.iconPickerCols(); got != 7 {
		t.Fatalf("at width 70 the grid has %d columns, want 7", got)
	}
}

// The cursor must move by the same number of columns the grid draws.
func TestIconPickerCursorFollowsTheGrid(t *testing.T) {
	m := searchModel(t)
	m.width = 100
	m.height = 30
	m.focus = focusEntries
	m.iconPickMode = true
	m.iconPickTarget = "entry"
	m.iconPickCursor = 0

	// Down moves one row, which is one full row of columns.
	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = after.(model)
	if m.iconPickCursor != m.iconPickerCols() {
		t.Fatalf("down moved to %d, want %d", m.iconPickCursor, m.iconPickerCols())
	}

	// Right moves one icon.
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = after.(model)
	if m.iconPickCursor != m.iconPickerCols()+1 {
		t.Fatalf("right moved to %d, want %d", m.iconPickCursor, m.iconPickerCols()+1)
	}
}

// The picker must name the icon under the cursor so it is clear what is
// selected.
func TestIconPickerShowsTheSelectedName(t *testing.T) {
	m := searchModel(t)
	m.width = 100
	m.height = 30
	m.iconPickMode = true
	m.iconPickTarget = "entry"
	m.iconPickCursor = 3

	view := m.iconPickView()
	if !strings.Contains(view, "Selected:") {
		t.Fatal("the picker should show which icon is selected")
	}
	if !strings.Contains(view, icons[3].Name) {
		t.Fatalf("the picker should show %q, got %q", icons[3].Name, view)
	}
}

// The picker must fit in the terminal, even a short one.
func TestIconPickerFitsTheTerminal(t *testing.T) {
	for _, height := range []int{14, 20, 30, 60} {
		m := searchModel(t)
		m.width = 100
		m.height = height
		m.iconPickMode = true
		m.iconPickTarget = "entry"
		m.iconPickCursor = len(icons) - 1

		view := m.iconPickView()
		if lines := len(strings.Split(view, "\n")); lines > height {
			t.Fatalf("at height %d the picker is %d lines tall", height, lines)
		}
		if !strings.Contains(view, "Selected:") {
			t.Fatalf("at height %d the picker lost the selected name", height)
		}
	}
}
