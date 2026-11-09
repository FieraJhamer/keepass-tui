package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

// Ctrl+S saves the database and clears the dirty flag.
func TestCtrlSSavesTheDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.kdbx")
	m := initialModel()
	m.db = testDB()
	m.filePath = path
	m.screen = screenBrowse
	m.dirty = true

	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = after.(model)

	if m.dirty {
		t.Fatal("ctrl+s should clear the dirty flag")
	}
	if !strings.Contains(m.statusMsg, "saved") {
		t.Fatalf("statusMsg = %q, want a save confirmation", m.statusMsg)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the database file should exist: %v", err)
	}

	// It also works from the detail screen.
	m.dirty = true
	m.screen = screenDetail
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = after.(model)
	if m.dirty {
		t.Fatal("ctrl+s should clear the dirty flag on the detail screen too")
	}
}

// The save hint shows up only when there is something to save.
func TestSaveHintOnlyShowsWhenDirty(t *testing.T) {
	m := searchModel(t)

	if strings.Contains(m.renderFooter(), "Ctrl+S") {
		t.Fatal("the save hint should not show when there is nothing to save")
	}

	m.dirty = true
	if !strings.Contains(m.renderFooter(), "[Ctrl+S] Save") {
		t.Fatalf("the footer should offer ctrl+s, got %q", m.renderFooter())
	}
}

// The c key copies the password from the detail screen.
func TestCopyKeyInTheDetailCopiesThePassword(t *testing.T) {
	m := searchModel(t)
	m.screen = screenDetail
	m.detailEntry = &m.db.Content.Root.Groups[0].Entries[0]

	after, _ := m.Update(keyRunes("c"))
	m = after.(model)

	if strings.HasPrefix(m.statusMsg, "Clipboard error") {
		t.Skipf("no clipboard in this environment: %s", m.statusMsg)
	}
	got, err := clipboard.ReadAll()
	if err != nil {
		t.Skipf("no clipboard in this environment: %v", err)
	}
	if got != "p1" {
		t.Fatalf("clipboard = %q, want %q", got, "p1")
	}
	if !strings.Contains(m.statusMsg, "copied") {
		t.Fatalf("statusMsg = %q, want a copy confirmation", m.statusMsg)
	}
}

// The detail screen offers the copy key.
func TestDetailHintsOfferCopy(t *testing.T) {
	m := searchModel(t)
	m.screen = screenDetail

	if !strings.Contains(m.renderFooter(), "[c] Copy") {
		t.Fatalf("the footer should offer the copy key, got %q", m.renderFooter())
	}
}

// Quitting with unsaved changes asks first. Saving and quitting has to save
// the database and close the program.
func TestSaveAndQuitClosesTheProgram(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.kdbx")
	m := initialModel()
	m.db = testDB()
	m.filePath = path
	m.screen = screenBrowse
	m.dirty = true

	// The q key opens the confirmation.
	after, _ := m.Update(keyRunes("q"))
	m = after.(model)
	if m.confirmMode != "quit" {
		t.Fatalf("confirmMode = %q, want %q", m.confirmMode, "quit")
	}

	// S saves the database and closes the program.
	after, cmd := m.Update(keyRunes("s"))
	m = after.(model)
	if !isQuit(cmd) {
		t.Fatal("save and quit should close the program")
	}
	if m.dirty {
		t.Fatal("save and quit should clear the dirty flag")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the database file should exist: %v", err)
	}
}

// Discarding the changes closes the program too.
func TestDiscardAndQuitClosesTheProgram(t *testing.T) {
	m := searchModel(t)
	m.dirty = true
	m.confirmMode = "quit"

	_, cmd := m.Update(keyRunes("d"))
	if !isQuit(cmd) {
		t.Fatal("discard and quit should close the program")
	}
}

// The quit command closes the program when there is nothing to save.
func TestQuitCommandClosesTheProgram(t *testing.T) {
	m := searchModel(t)

	_, cmd := m.executeCommand("/quit")
	if !isQuit(cmd) {
		t.Fatal("/quit should close the program")
	}
}

// isQuit tells if a command closes the program.
func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}
