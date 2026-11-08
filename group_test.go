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
	m.groupCursor = 0
	return m
}

func TestGroupsAreSeparatedByABlankLine(t *testing.T) {
	m := groupModel(t)

	lines := strings.Split(strings.TrimRight(m.renderSidebar(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines for 2 groups, got %d: %q", len(lines), lines)
	}
	if lines[1] != "" {
		t.Fatalf("the groups should be separated by a blank line, got %q", lines[1])
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
