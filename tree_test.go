package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// The database root cannot be deleted, renamed or given an icon.
func TestTheDatabaseRootIsProtected(t *testing.T) {
	m := searchModel(t)
	m.focus = focusGroups
	m.groupCursor = 0 // the root

	if !m.atRoot() {
		t.Fatal("the first node should be the root")
	}

	// Delete.
	after, _ := m.Update(keyRunes("d"))
	m = after.(model)
	if m.confirmMode != "" {
		t.Fatalf("the root must not be deletable, confirmMode = %q", m.confirmMode)
	}
	if !strings.Contains(m.statusMsg, "root") {
		t.Fatalf("the status should explain why, got %q", m.statusMsg)
	}

	// Rename.
	after, _ = m.Update(keyRunes("r"))
	m = after.(model)
	if m.editMode != editNone {
		t.Fatalf("the root must not be renamable, editMode = %v", m.editMode)
	}

	// Icon.
	after, _ = m.Update(keyRunes("i"))
	m = after.(model)
	if m.iconPickMode {
		t.Fatal("the root must not get an icon")
	}
}

// A group created on the root becomes a top level group.
func TestNewGroupOnTheRootIsTopLevel(t *testing.T) {
	m := searchModel(t)
	m.focus = focusGroups
	m.groupCursor = 0 // the root

	after, _ := m.Update(keyRunes("n"))
	m = after.(model)
	if m.editMode != editNewGroup {
		t.Fatalf("editMode = %v, want editNewGroup", m.editMode)
	}

	m.formGroup.SetValue("Top")
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	found := false
	for _, g := range m.db.Content.Root.Groups {
		if g.Name == "Top" {
			found = true
		}
	}
	if !found {
		t.Fatalf("the group should be at the top level, got %v", m.db.Content.Root.Groups)
	}
	// And it shows as a child of the root in the sidebar.
	if m.groups[0].path != "" {
		t.Fatalf("the root path must stay empty, got %q", m.groups[0].path)
	}
}

// The only top level group cannot be deleted, there would be no group left.
func TestTheOnlyTopLevelGroupIsProtected(t *testing.T) {
	m := searchModel(t)
	m.db.Content.Root.Groups = m.db.Content.Root.Groups[:1]
	m.groups = m.groupTree()
	m.focus = focusGroups
	m.groupCursor = 1

	after, _ := m.Update(keyRunes("d"))
	m = after.(model)
	if m.confirmMode != "" {
		t.Fatalf("the only top level group must not be deletable, confirmMode = %q", m.confirmMode)
	}
}

// The confirmation says how much is inside the group before deleting it.
func TestDeleteGroupWarningCountsTheContents(t *testing.T) {
	m := searchModel(t)
	m.db.Content.Root.Groups[0].Groups = append(m.db.Content.Root.Groups[0].Groups, newGroup("Sub", 0))
	m.db.Content.Root.Groups[0].Groups[0].Entries = append(
		m.db.Content.Root.Groups[0].Groups[0].Entries,
		newEntry("One", "u", "p", "", "", 0))
	m.groups = m.groupTree()
	m.focus = focusGroups
	m.groupCursor = 1

	after, _ := m.Update(keyRunes("d"))
	m = after.(model)

	if m.confirmMode != "deleteGroup" {
		t.Fatalf("confirmMode = %q, want deleteGroup", m.confirmMode)
	}
	if !strings.Contains(m.confirmMsg, "1 group and 3 entries") {
		t.Fatalf("the warning should count the contents, got %q", m.confirmMsg)
	}
}

// The entries panel tells which group the entries belong to.
func TestEntriesPanelShowsTheCurrentGroup(t *testing.T) {
	m := searchModel(t)
	m.focus = focusEntries

	view := m.View()
	if !strings.Contains(view, "General") {
		t.Fatalf("the panel should show the group path, got %q", view)
	}
}

// An entry cannot be created on the database root.
func TestNewEntryOnTheRootIsRefused(t *testing.T) {
	m := searchModel(t)
	m.focus = focusEntries
	m.groupCursor = 0 // the root

	after, _ := m.Update(keyRunes("n"))
	m = after.(model)

	if m.editMode != editNone {
		t.Fatalf("the form should not open on the root, editMode = %v", m.editMode)
	}
	if !strings.Contains(m.statusMsg, "Select a group") {
		t.Fatalf("the status should explain why, got %q", m.statusMsg)
	}
}
