package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tobischo/gokeepasslib/v3"
)

func searchTestDB() *gokeepasslib.Database {
	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials("master")
	db.Content.Root.Groups = []gokeepasslib.Group{
		{
			Name: "General",
			Entries: []gokeepasslib.Entry{
				newEntry("Gmail", "user@gmail.com", "p1", "https://mail.google.com", "", 0),
				newEntry("Bank", "user@bank.com", "p2", "https://bank.com", "", 0),
			},
		},
		{
			Name: "Work",
			Entries: []gokeepasslib.Entry{
				newEntry("Jira", "user@jira.com", "p3", "https://jira.com", "", 0),
			},
		},
	}
	return db
}

func searchModel(t *testing.T) model {
	t.Helper()
	m := initialModel()
	m.db = searchTestDB()
	m.groups = flattenGroups(m.db.Content.Root.Groups, "", 0)
	m.screen = screenBrowse
	m.width = 120
	m.height = 40
	return m.loadEntries()
}

func TestSearchFindsEntriesAcrossGroups(t *testing.T) {
	m := searchModel(t)

	after, _ := m.Update(keyRunes("/"))
	m = after.(model)
	m.command = "/search gmail"
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	if m.searchQuery != "gmail" {
		t.Fatalf("searchQuery = %q, want %q", m.searchQuery, "gmail")
	}
	if len(m.entries) != 1 {
		t.Fatalf("expected 1 result, got %d", len(m.entries))
	}
	if got := getEntryValue(m.entries[0].entry, "Title"); got != "Gmail" {
		t.Fatalf("title = %q, want %q", got, "Gmail")
	}
	if m.entries[0].groupPath != "General" {
		t.Fatalf("groupPath = %q, want %q", m.entries[0].groupPath, "General")
	}
	if m.focus != focusEntries {
		t.Fatal("the search should move the focus to the results")
	}

	view := m.View()
	if !strings.Contains(view, "Results for") {
		t.Fatal("the view should show the search results header")
	}
	if !strings.Contains(view, "General") {
		t.Fatal("the results should show the group each entry comes from")
	}
}

func TestSearchMatchesUsernameAndURL(t *testing.T) {
	m := searchModel(t)

	for _, q := range []string{"bank.com", "user@jira.com", "JIRA"} {
		items := searchItems(m.db.Content.Root.Groups, q)
		if len(items) == 0 {
			t.Fatalf("query %q returned no results", q)
		}
	}
}

func TestSearchResultsPointIntoTheDatabase(t *testing.T) {
	m := searchModel(t)

	items := searchItems(m.db.Content.Root.Groups, "gmail")
	if len(items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(items))
	}

	// Editing through the result must change the entry inside the database.
	setEntryValue(items[0].entry, "Title", "Gmail renamed")
	group := findGroup(m.db.Content.Root.Groups, "General")
	if got := getEntryValue(&group.Entries[0], "Title"); got != "Gmail renamed" {
		t.Fatalf("database title = %q, want %q", got, "Gmail renamed")
	}
}

// Deleting a search result must remove the entry from the group it came from,
// not the entry at the same position in the currently selected group.
func TestDeletingASearchResultRemovesTheRightEntry(t *testing.T) {
	m := searchModel(t)
	m.focus = focusEntries

	// Search for the Work group entry while the General group is selected.
	items := searchItems(m.db.Content.Root.Groups, "jira")
	if len(items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(items))
	}
	m.entries = items
	m.entryCursor = 0

	// Confirm the delete.
	m.confirmMode = "deleteEntry"
	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = after.(model)

	work := findGroup(m.db.Content.Root.Groups, "Work")
	if len(work.Entries) != 0 {
		t.Fatalf("the Work entry should be gone, got %d entries", len(work.Entries))
	}
	general := findGroup(m.db.Content.Root.Groups, "General")
	if len(general.Entries) != 2 {
		t.Fatalf("the General entries should be untouched, got %d", len(general.Entries))
	}
	if got := getEntryValue(&general.Entries[0], "Title"); got != "Gmail" {
		t.Fatalf("General title = %q, want %q", got, "Gmail")
	}
}

func TestSearchWithoutResultsShowsNoMatches(t *testing.T) {
	m := searchModel(t)

	after, _ := m.Update(keyRunes("/"))
	m = after.(model)
	m.command = "/search nothingmatches"
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	if len(m.entries) != 0 {
		t.Fatalf("expected no results, got %d", len(m.entries))
	}
	if !strings.Contains(m.View(), "No matching entries") {
		t.Fatal("the view should say there are no matching entries")
	}
}

func TestEscapeClearsTheSearch(t *testing.T) {
	m := searchModel(t)

	after, _ := m.Update(keyRunes("/"))
	m = after.(model)
	m.command = "/search gmail"
	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = after.(model)

	if m.searchQuery != "" {
		t.Fatalf("the search should be cleared, got %q", m.searchQuery)
	}
	if m.focus != focusGroups {
		t.Fatal("escape should return the focus to the groups")
	}
	if len(m.entries) != 2 {
		t.Fatalf("the group's entries should be back, got %d", len(m.entries))
	}
}
