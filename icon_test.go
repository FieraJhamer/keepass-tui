package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func entryFormModel(t *testing.T) model {
	t.Helper()
	m := searchModel(t)
	m.focus = focusEntries
	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)
	after, _ = m.Update(keyRunes("e"))
	return after.(model)
}

func TestNewEntryFormOpensTheIconPicker(t *testing.T) {
	m := searchModel(t)
	m.focus = focusEntries

	after, _ := m.Update(keyRunes("n"))
	m = after.(model)
	if m.editMode != editNewEntry {
		t.Fatalf("editMode = %v, want editNewEntry", m.editMode)
	}

	// Tab five times to reach the icon field.
	for i := 0; i < 5; i++ {
		after, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
		m = after.(model)
	}
	if m.formField != 5 {
		t.Fatalf("formField = %d, want 5", m.formField)
	}

	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)
	if !m.iconPickMode {
		t.Fatal("the icon picker should open on the icon field")
	}
	if m.iconPickTarget != "form" {
		t.Fatalf("iconPickTarget = %q, want %q", m.iconPickTarget, "form")
	}
}

func TestIconChosenInTheFormIsAppliedToTheNewEntry(t *testing.T) {
	m := searchModel(t)
	m.focus = focusEntries

	after, _ := m.Update(keyRunes("n"))
	m = after.(model)

	m.formTitle.SetValue("With icon")
	for i := 0; i < 5; i++ {
		after, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
		m = after.(model)
	}
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	// Pick the second icon and confirm.
	m.iconPickCursor = 1
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	if m.iconPickMode {
		t.Fatal("the icon picker should close after picking")
	}
	if m.formIconID != 1 {
		t.Fatalf("formIconID = %d, want 1", m.formIconID)
	}

	// Save the entry.
	m.formField = 0
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	group := findGroup(m.db.Content.Root.Groups, "General")
	if len(group.Entries) != 3 {
		t.Fatalf("expected 3 entries in General, got %d", len(group.Entries))
	}
	if group.Entries[2].IconID != 1 {
		t.Fatalf("entry icon = %d, want 1", group.Entries[2].IconID)
	}
}

func TestEditingAnEntryLoadsItsIcon(t *testing.T) {
	m := entryFormModel(t)
	if m.editMode != editEntry {
		t.Fatalf("editMode = %v, want editEntry", m.editMode)
	}
	if m.formIconID != 0 {
		t.Fatalf("formIconID = %d, want 0", m.formIconID)
	}

	// Change the icon and save.
	m.formIconID = 7
	m.formField = 0
	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	if e := m.currentEntry(); e == nil || e.IconID != 7 {
		t.Fatalf("the entry icon should be 7, got %v", e)
	}
}

func TestIconPickerCursorIsClampedToTheKnownIcons(t *testing.T) {
	// Real databases can carry icon IDs outside this app's mapping.
	if got := iconCursorFor(9999); got != 0 {
		t.Fatalf("iconCursorFor(9999) = %d, want 0", got)
	}
	if got := iconCursorFor(-1); got != 0 {
		t.Fatalf("iconCursorFor(-1) = %d, want 0", got)
	}
	if got := iconCursorFor(3); got != 3 {
		t.Fatalf("iconCursorFor(3) = %d, want 3", got)
	}
}
