package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tobischo/gokeepasslib/v3"
)

func groupModel(t *testing.T) model {
	t.Helper()
	m := searchModel(t)
	m.focus = focusGroups
	// Index 1 is the first real group, index 0 is the database root.
	m.groupCursor = 1
	return m
}

// The groups take one line each, with no blank line between them.
func TestGroupsHaveNoBlankLineBetweenThem(t *testing.T) {
	m := groupModel(t)

	lines := strings.Split(strings.TrimRight(m.renderSidebar(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines for the root and 2 groups, got %d: %q", len(lines), lines)
	}
	for _, l := range lines {
		if l == "" {
			t.Fatalf("the groups should not have blank lines between them, got %q", lines)
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

// A group name can hold a slash or repeat the name of a sibling. The app
// finds groups by uuid, so those groups still show their own entries.
func TestAwkwardGroupNamesStillShowTheirEntries(t *testing.T) {
	m := searchModel(t)
	m.db.Content.Root.Groups = m.db.Content.Root.Groups[:1]
	db := &m.db.Content.Root.Groups[0]
	db.Name = "General"

	slashed := newGroup("Bancos/Internet", 0)
	slashed.Entries = []gokeepasslib.Entry{newEntry("Banco X", "u", "p", "url", "", 0)}
	twin := newGroup("General", 0)
	twin.Entries = []gokeepasslib.Entry{newEntry("Twin", "u", "p", "url", "", 0)}
	db.Groups = []gokeepasslib.Group{slashed, twin}
	m.groups = m.groupTree()

	// The tree is root, General, Bancos/Internet and the other General.
	m.groupCursor = 2
	m = m.loadEntries()
	if len(m.entries) != 1 || getEntryValue(m.entries[0].entry, "Title") != "Banco X" {
		t.Fatalf("the group with a slash should show its entry, got %d entries", len(m.entries))
	}

	m.groupCursor = 3
	m = m.loadEntries()
	if len(m.entries) != 1 || getEntryValue(m.entries[0].entry, "Title") != "Twin" {
		t.Fatalf("the group with the same name should show its own entry, got %d entries", len(m.entries))
	}
}

// Renaming a group whose name holds a slash still reaches the right group.
func TestRenamingAGroupWithASlash(t *testing.T) {
	m := searchModel(t)
	m.focus = focusGroups
	m.db.Content.Root.Groups = m.db.Content.Root.Groups[:1]
	db := &m.db.Content.Root.Groups[0]
	db.Name = "General"
	db.Groups = []gokeepasslib.Group{newGroup("Bancos/Internet", 0)}
	m.groups = m.groupTree()
	m.groupCursor = 2

	after, _ := m.Update(keyRunes("r"))
	m = after.(model)
	if m.editMode != editGroup {
		t.Fatalf("editMode = %v, want editGroup", m.editMode)
	}
	m.formGroup.SetValue("Bancos")
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	if db.Groups[0].Name != "Bancos" {
		t.Fatalf("the group name = %q, want %q", db.Groups[0].Name, "Bancos")
	}
}

// Deleting a group whose name holds a slash removes that group and not the
// first one with the same name.
func TestDeletingAGroupWithASlash(t *testing.T) {
	m := searchModel(t)
	m.focus = focusGroups
	m.db.Content.Root.Groups = m.db.Content.Root.Groups[:1]
	db := &m.db.Content.Root.Groups[0]
	db.Name = "General"
	db.Groups = []gokeepasslib.Group{newGroup("Bancos/Internet", 0)}
	m.groups = m.groupTree()
	m.groupCursor = 2

	after, _ := m.Update(keyRunes("d"))
	m = after.(model)
	if m.confirmMode != "deleteGroup" {
		t.Fatalf("confirmMode = %q, want %q", m.confirmMode, "deleteGroup")
	}
	after, _ = m.Update(keyRunes("y"))
	m = after.(model)

	if len(db.Groups) != 0 {
		t.Fatalf("the group should be gone, got %d groups", len(db.Groups))
	}
	if len(m.db.Content.Root.Groups) != 1 || m.db.Content.Root.Groups[0].Name != "General" {
		t.Fatal("the other group should be untouched")
	}
}

// A new group goes inside the selected group.
func TestNewGroupGoesInsideTheSelectedGroup(t *testing.T) {
	m := groupModel(t)

	after, _ := m.Update(keyRunes("n"))
	m = after.(model)
	m.formGroup.SetValue("Trabajo")
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	general := groupByName(m.db.Content.Root.Groups, "General")
	if len(general.Groups) != 1 || general.Groups[0].Name != "Trabajo" {
		t.Fatalf("the new group should be inside General, got %d sub groups", len(general.Groups))
	}
}

// A new group on the root goes to the top level.
func TestNewGroupOnTheRootGoesToTheTopLevel(t *testing.T) {
	m := groupModel(t)
	m.groupCursor = 0 // the database root

	after, _ := m.Update(keyRunes("n"))
	m = after.(model)
	m.formGroup.SetValue("Trabajo")
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	if len(m.db.Content.Root.Groups) != 3 {
		t.Fatalf("expected 3 top level groups, got %d", len(m.db.Content.Root.Groups))
	}
	if m.db.Content.Root.Groups[2].Name != "Trabajo" {
		t.Fatalf("the new group = %q, want %q", m.db.Content.Root.Groups[2].Name, "Trabajo")
	}
}

// A broken database can repeat a uuid or leave it empty. The path breaks the
// tie, so each group still shows its own entries.
func TestGroupsWithBrokenUUIDsStillShowTheirEntries(t *testing.T) {
	m := searchModel(t)
	m.db.Content.Root.Groups = m.db.Content.Root.Groups[:1]
	db := &m.db.Content.Root.Groups[0]
	db.Name = "General"

	first := newGroup("Primero", 0)
	first.Entries = []gokeepasslib.Entry{newEntry("One", "u", "p", "url", "", 0)}
	second := newGroup("Segundo", 0)
	second.Entries = []gokeepasslib.Entry{newEntry("Two", "u", "p", "url", "", 0)}
	second.UUID = first.UUID // the same uuid
	third := newGroup("Tercero", 0)
	third.Entries = []gokeepasslib.Entry{newEntry("Three", "u", "p", "url", "", 0)}
	third.UUID = gokeepasslib.UUID{}
	db.Groups = []gokeepasslib.Group{first, second, third}
	m.groups = m.groupTree()

	// The tree is root, General, Primero, Segundo and Tercero.
	want := []struct {
		cursor int
		title  string
	}{
		{2, "One"},
		{3, "Two"},
		{4, "Three"},
	}
	for _, w := range want {
		m.groupCursor = w.cursor
		m = m.loadEntries()
		if len(m.entries) != 1 || getEntryValue(m.entries[0].entry, "Title") != w.title {
			t.Fatalf("the group at %d should show %q, got %d entries", w.cursor, w.title, len(m.entries))
		}
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

	group := groupByName(m.db.Content.Root.Groups, "General")
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

	group := groupByName(m.db.Content.Root.Groups, "General")
	if group.Entries[0].IconID != 7 {
		t.Fatalf("the entry icon = %d, want 7", group.Entries[0].IconID)
	}
}
