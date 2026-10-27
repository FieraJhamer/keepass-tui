package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keyRunes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestMasterAndEntryPasswordVisibilityAreIndependent(t *testing.T) {
	m := initialModel()
	m.screen = screenUnlock

	// Ctrl+P on the unlock screen reveals only the master password.
	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	unlocked := after.(model)
	if !unlocked.showPassword {
		t.Fatal("ctrl+p should reveal the master password")
	}
	if unlocked.showEntryPass {
		t.Fatal("ctrl+p must not reveal entry passwords")
	}

	// p on the detail screen reveals only the entry password.
	unlocked.screen = screenDetail
	after, _ = unlocked.Update(keyRunes("p"))
	detail := after.(model)
	if !detail.showEntryPass {
		t.Fatal("p should reveal the entry password")
	}
	if !detail.showPassword {
		t.Fatal("p must not change master password visibility")
	}
}

func TestEntryPasswordStartsHiddenOnEveryEntry(t *testing.T) {
	m := initialModel()
	m.screen = screenBrowse
	m.focus = focusEntries
	m.entries = testDB().Content.Root.Groups[0].Entries
	m.groups = []groupInfo{{name: "General", path: "General"}}

	// Enter opens the detail with the password hidden.
	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	detail := after.(model)
	if detail.screen != screenDetail {
		t.Fatalf("expected the detail screen, got %v", detail.screen)
	}
	if detail.showEntryPass {
		t.Fatal("entry password must start hidden")
	}

	// p reveals it.
	after, _ = detail.Update(keyRunes("p"))
	revealed := after.(model)
	if !revealed.showEntryPass {
		t.Fatal("p should reveal the entry password")
	}

	// Leaving hides it again.
	after, _ = revealed.Update(tea.KeyMsg{Type: tea.KeyEscape})
	back := after.(model)
	if back.showEntryPass {
		t.Fatal("leaving the detail view should hide the entry password")
	}
}
