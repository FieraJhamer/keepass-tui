package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func groupModel(t *testing.T) model {
	t.Helper()
	m := searchModel(t)
	m.focus = focusGroups
	// Index 1 is the first real group, index 0 is the database root.
	m.groupCursor = 1
	return m
}

func TestGroupsAreSeparatedByABlankLine(t *testing.T) {
	m := groupModel(t)

	lines := strings.Split(strings.TrimRight(m.renderSidebar(), "\n"), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines for the root and 2 groups, got %d: %q", len(lines), lines)
	}
	// The separator is a plain blank line.
	for _, i := range []int{1, 3} {
		if lines[i] != "" {
			t.Fatalf("the groups should be separated by a blank line, got %q", lines)
		}
	}
}

// The entries of a group are separated by a blank line too.
func TestEntriesAreSeparatedByABlankLine(t *testing.T) {
	m := searchModel(t)
	m.focus = focusEntries

	lines := strings.Split(strings.TrimRight(m.renderEntryList(10), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines for 2 entries, got %d: %q", len(lines), lines)
	}
	if lines[1] != "" {
		t.Fatalf("the entries should be separated by a blank line, got %q", lines)
	}
}

// The sidebar draws the groups as a tree, like the tree command.
func TestSidebarDrawsTheGroupsAsATree(t *testing.T) {
	m := groupModel(t)

	// A group inside General.
	m.db.Content.Root.Groups[0].Groups = append(m.db.Content.Root.Groups[0].Groups, newGroup("Trabajo", 0))
	m.groups = m.groupTree()

	lines := strings.Split(m.renderSidebar(), "\n")
	joined := strings.Join(lines, "\n")

	if !strings.Contains(joined, "├───General") {
		t.Fatalf("General should have a branch, got %q", joined)
	}
	if !strings.Contains(joined, "│   └───Trabajo") {
		t.Fatalf("Trabajo should hang from General, got %q", joined)
	}
	if !strings.Contains(joined, "└───Work") {
		t.Fatalf("Work should be the last branch, got %q", joined)
	}
	// The icon goes on the right of the title.
	if !strings.Contains(joined, "General ") || !strings.Contains(joined, "General                     🔑") {
		t.Fatalf("the icon should be at the right of the title, got %q", joined)
	}
}

func TestNewOnAGroupCreatesAGroup(t *testing.T) {
	m := groupModel(t)

	after, _ := m.Update(keyRunes("n"))
	m = after.(model)

	if m.editMode != editNewGroup {
		t.Fatalf("editMode = %v, want editNewGroup", m.editMode)
	}
}

func TestNewOnAnEntryListCreatesAnEntry(t *testing.T) {
	m := searchModel(t)
	m.focus = focusEntries

	after, _ := m.Update(keyRunes("n"))
	m = after.(model)

	if m.editMode != editNewEntry {
		t.Fatalf("editMode = %v, want editNewEntry", m.editMode)
	}
}

func TestIconOnAGroupChangesTheGroupIcon(t *testing.T) {
	m := groupModel(t)

	after, _ := m.Update(keyRunes("i"))
	m = after.(model)

	if !m.iconPickMode {
		t.Fatal("the icon picker should open")
	}
	if m.iconPickTarget != "group" {
		t.Fatalf("iconPickTarget = %q, want %q", m.iconPickTarget, "group")
	}

	// Pick an icon and confirm.
	m.iconPickCursor = 5
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	group := findGroup(m.db.Content.Root.Groups, "General")
	if group.IconID != 5 {
		t.Fatalf("the group icon = %d, want 5", group.IconID)
	}
	if m.groups[m.groupCursor].iconID != 5 {
		t.Fatalf("the sidebar icon = %d, want 5", m.groups[m.groupCursor].iconID)
	}
	// The group's entries must not be touched.
	if group.Entries[0].IconID != 0 {
		t.Fatalf("the entry icon should be untouched, got %d", group.Entries[0].IconID)
	}
}

func TestIconOnAnEntryChangesTheEntryIcon(t *testing.T) {
	m := searchModel(t)
	m.focus = focusEntries

	after, _ := m.Update(keyRunes("i"))
	m = after.(model)

	if m.iconPickTarget != "entry" {
		t.Fatalf("iconPickTarget = %q, want %q", m.iconPickTarget, "entry")
	}

	m.iconPickCursor = 7
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	group := findGroup(m.db.Content.Root.Groups, "General")
	if group.Entries[0].IconID != 7 {
		t.Fatalf("the entry icon = %d, want 7", group.Entries[0].IconID)
	}
}
