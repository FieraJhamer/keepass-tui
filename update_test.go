package main

import (
	"strings"
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
	m.db = testDB()
	m.groups = m.groupTree()
	m.groupCursor = 1 // the General group
	m = m.loadEntries()

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

// The command bar sits on top of the screen, not below the content.
func TestCommandBarSitsOnTopOfTheScreen(t *testing.T) {
	m := searchModel(t)

	after, _ := m.Update(keyRunes("/"))
	m = after.(model)
	m.command = "/search gmail"

	view := m.View()
	lines := strings.Split(view, "\n")
	if !strings.Contains(lines[0], "/search gmail") {
		t.Fatalf("the command bar should be the first line, got %q", lines[0])
	}
	if strings.Contains(view, "[/] Command") {
		t.Fatal("the key hints should be hidden while the command bar is up")
	}
}

// The key hints sit on top of the screen, not below the content.
func TestFooterSitsOnTopOfTheScreen(t *testing.T) {
	m := searchModel(t)

	lines := strings.Split(m.View(), "\n")
	if !strings.Contains(lines[0], "[/] Command") {
		t.Fatalf("the key hints should be the first line, got %q", lines[0])
	}
	if strings.Contains(m.View(), "\n\n[/] Command") {
		t.Fatal("the key hints should not be repeated at the bottom")
	}
}

// Editing from the detail view must change the entry that is open, not
// another entry of the same group.
func TestEditingFromTheDetailViewChangesTheOpenEntry(t *testing.T) {
	m := searchModel(t)
	m.focus = focusEntries

	// Open the detail of the first entry (Gmail).
	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)
	if got := getEntryValue(m.detailEntry, "Title"); got != "Gmail" {
		t.Fatalf("open entry = %q, want %q", got, "Gmail")
	}

	// Edit it.
	after, _ = m.Update(keyRunes("e"))
	m = after.(model)
	if m.editMode != editEntry {
		t.Fatalf("editMode = %v, want editEntry", m.editMode)
	}
	m.formTitle.SetValue("Gmail nuevo")
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	group := groupByName(m.db.Content.Root.Groups, "General")
	if got := getEntryValue(&group.Entries[0], "Title"); got != "Gmail nuevo" {
		t.Fatalf("first entry = %q, want %q", got, "Gmail nuevo")
	}
	if got := getEntryValue(&group.Entries[1], "Title"); got != "Bank" {
		t.Fatalf("the other entry should be untouched, got %q", got)
	}
}
