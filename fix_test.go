package main

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// The copy commands that were removed must report an unknown command.
func TestRemovedCopyCommandsAreUnknown(t *testing.T) {
	for _, cmd := range []string{"/copy-url", "/copy-all"} {
		m := searchModel(t)
		m.focus = focusEntries

		after, _ := m.Update(keyRunes("/"))
		m = after.(model)
		after, _ = m.Update(keyRunes(cmd[1:]))
		m = after.(model)
		after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = after.(model)

		if !strings.Contains(m.statusMsg, "Unknown command") {
			t.Fatalf("%s should be unknown, statusMsg = %q", cmd, m.statusMsg)
		}
	}
}

// The key hints must stay on one line at any width and always keep the quit
// hint.
func TestHintsFitOnOneLine(t *testing.T) {
	for _, width := range []int{60, 80, 100, 140} {
		m := searchModel(t)
		m.width = width
		m.height = 24

		lines := strings.Split(m.View(), "\n")
		if !strings.Contains(lines[0], "[q] Quit") {
			t.Fatalf("at width %d the quit hint is missing", width)
		}
		if strings.Contains(lines[1], "[q] Quit") {
			t.Fatalf("at width %d the hints wrapped to a second line", width)
		}
	}
}

// The first entry of an empty group must be reachable: the focus has to move
// to the entry list even when the group has nothing in it.
func TestFirstEntryOfAnEmptyGroupIsReachable(t *testing.T) {
	m := groupModel(t)
	// Point the group cursor at an empty group.
	m.db.Content.Root.Groups[0].Groups = append(m.db.Content.Root.Groups[0].Groups, newGroup("Empty", 0))
	m.groups = flattenGroups(m.db.Content.Root.Groups, "", 0)
	m.groupCursor = 1
	m = m.loadEntries()

	if len(m.entries) != 0 {
		t.Fatalf("the group should be empty, entries = %d", len(m.entries))
	}

	// Enter moves to the entry list.
	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)
	if m.focus != focusEntries {
		t.Fatalf("focus = %v, want focusEntries", m.focus)
	}

	// And now n creates an entry, not a group.
	after, _ = m.Update(keyRunes("n"))
	m = after.(model)
	if m.editMode != editNewEntry {
		t.Fatalf("editMode = %v, want editNewEntry", m.editMode)
	}

	m.formTitle.SetValue("First one")
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	group := findGroup(m.db.Content.Root.Groups, "General/Empty")
	if group == nil || len(group.Entries) != 1 {
		t.Fatalf("the entry should be in the empty group, group = %v", group)
	}
}

// Tab moves between the group list and the entry list.
func TestTabMovesBetweenTheLists(t *testing.T) {
	m := groupModel(t)

	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = after.(model)
	if m.focus != focusEntries {
		t.Fatalf("tab should focus the entries, focus = %v", m.focus)
	}

	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = after.(model)
	if m.focus != focusGroups {
		t.Fatalf("shift+tab should focus the groups, focus = %v", m.focus)
	}
}

// The master password must accept any character, including q and Q.
func TestUnlockAcceptsAnyCharacterInThePassword(t *testing.T) {
	m := initialModel()
	m.screen = screenUnlock

	for _, r := range "qQ password 123!#" {
		after, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = after.(model)
		if m.screen != screenUnlock {
			t.Fatalf("typing %q left the unlock screen", r)
		}
	}
	if got := m.unlockInput.Value(); got != "qQ password 123!#" {
		t.Fatalf("password = %q, want %q", got, "qQ password 123!#")
	}
}

// The cursor can move inside the password so typos can be fixed.
func TestUnlockAllowsMovingTheCursor(t *testing.T) {
	m := initialModel()
	m.screen = screenUnlock

	after, _ := m.Update(keyRunes("abc"))
	m = after.(model)
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = after.(model)
	after, _ = m.Update(keyRunes("X"))
	m = after.(model)

	if got := m.unlockInput.Value(); got != "abXc" {
		t.Fatalf("password = %q, want %q", got, "abXc")
	}
}

// Deleting a group must not delete one of its entries.
func TestDeleteOnAGroupDeletesTheGroup(t *testing.T) {
	m := groupModel(t)

	after, _ := m.Update(keyRunes("d"))
	m = after.(model)

	if m.confirmMode != "deleteGroup" {
		t.Fatalf("confirmMode = %q, want deleteGroup", m.confirmMode)
	}
	if !strings.Contains(m.confirmMsg, "General") {
		t.Fatalf("the confirmation should name the group, got %q", m.confirmMsg)
	}

	// Confirming removes the group and keeps its entries out of harm's way.
	after, _ = m.Update(keyRunes("y"))
	m = after.(model)

	if len(m.groups) != 2 || m.groups[1].name != "Work" {
		t.Fatalf("only Work should be left, groups = %v", m.groups)
	}
	if m.db.Content.Root.Groups[0].Name != "Work" {
		t.Fatalf("the wrong group was deleted: %q", m.db.Content.Root.Groups[0].Name)
	}
}

// Deleting an entry from the entry list must still work.
func TestDeleteOnAnEntryDeletesTheEntry(t *testing.T) {
	m := searchModel(t)
	m.focus = focusEntries

	after, _ := m.Update(keyRunes("d"))
	m = after.(model)

	if m.confirmMode != "deleteEntry" {
		t.Fatalf("confirmMode = %q, want deleteEntry", m.confirmMode)
	}

	after, _ = m.Update(keyRunes("y"))
	m = after.(model)

	group := findGroup(m.db.Content.Root.Groups, "General")
	if len(group.Entries) != 1 {
		t.Fatalf("expected 1 entry left, got %d", len(group.Entries))
	}
}

// The entries below the fold must stay reachable: the cursor moves and the
// list scrolls to follow it.
func TestEntryListScrollsToKeepTheCursorVisible(t *testing.T) {
	m := searchModel(t)
	m.width = 80
	m.height = 24
	m.focus = focusEntries

	group := findGroup(m.db.Content.Root.Groups, "General")
	for i := 0; i < 40; i++ {
		group.Entries = append(group.Entries, newEntry(fmt.Sprintf("Entry %02d", i), "user", "pass", "", "", 0))
	}
	m = m.loadEntries()

	// The view must never be taller than the screen.
	if lines := len(strings.Split(m.View(), "\n")); lines > m.height {
		t.Fatalf("the view is %d lines tall, the screen is %d", lines, m.height)
	}

	// Move to the last entry and check it is on screen.
	m.entryCursor = len(m.entries) - 1
	view := m.View()
	if !strings.Contains(view, "Entry 39") {
		t.Fatal("the last entry should be visible once the cursor reaches it")
	}
	if lines := len(strings.Split(view, "\n")); lines > m.height {
		t.Fatalf("the view is %d lines tall, the screen is %d", lines, m.height)
	}
}

// The same applies to the groups in the sidebar.
func TestSidebarScrollsToKeepTheCursorVisible(t *testing.T) {
	m := searchModel(t)
	m.width = 80
	m.height = 24
	m.focus = focusGroups

	for i := 0; i < 30; i++ {
		m.db.Content.Root.Groups[0].Groups = append(m.db.Content.Root.Groups[0].Groups,
			newGroup(fmt.Sprintf("Group %02d", i), 0))
	}
	m.groups = m.groupTree()

	if lines := len(strings.Split(m.View(), "\n")); lines > m.height {
		t.Fatalf("the view is %d lines tall, the screen is %d", lines, m.height)
	}

	m.groupCursor = len(m.groups) - 1
	view := m.View()
	if !strings.Contains(view, "Group 29") {
		t.Fatal("the last group should be visible once the cursor reaches it")
	}
}
