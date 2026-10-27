package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func TestCreateDatabaseProducesReopenableKDBX(t *testing.T) {
	path := filepath.Join(t.TempDir(), "New.kdbx")

	db, err := createDatabase(path, "masterpw")
	if err != nil {
		t.Fatalf("createDatabase failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("database file was not created: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("database file is empty")
	}
	if !db.Header.IsKdbx4() {
		t.Fatal("expected a KDBX 4 database")
	}

	// The on-disk signature must match what other KeePass tools look for:
	// base signature, secondary signature, then minor (uint16) and major (uint16).
	head, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x03, 0xd9, 0xa2, 0x9a, 0x67, 0xfb, 0x4b, 0xb5, 0x00, 0x00, 0x04, 0x00}
	if len(head) < len(want) || string(head[:len(want)]) != string(want) {
		t.Fatalf("unexpected kdbx signature %x, want %x", head[:len(want)], want)
	}

	// The file on disk must reopen with the same master password.
	reopened, err := openDatabase(path, "masterpw")
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	if len(reopened.Content.Root.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(reopened.Content.Root.Groups))
	}
	if reopened.Content.Root.Groups[0].Name != "General" {
		t.Fatalf("group name = %q, want %q", reopened.Content.Root.Groups[0].Name, "General")
	}

	// A wrong master password must be rejected.
	if _, err := openDatabase(path, "wrong"); err == nil {
		t.Fatal("expected opening with a wrong password to fail")
	}
}

func TestCreateDatabaseEntryRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "New.kdbx")
	db, err := createDatabase(path, "masterpw")
	if err != nil {
		t.Fatalf("createDatabase failed: %v", err)
	}

	group := findGroup(db.Content.Root.Groups, "General")
	if group == nil {
		t.Fatal("General group not found")
	}
	group.Entries = append(group.Entries, newEntry("Mail", "u@e.com", "s3cret", "https://e.com", "", 0))
	if err := saveDatabase(db, path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	reopened, err := openDatabase(path, "masterpw")
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	g := findGroup(reopened.Content.Root.Groups, "General")
	if g == nil || len(g.Entries) != 1 {
		t.Fatalf("expected 1 entry in General, got %v", g)
	}
	if got := getEntryValue(&g.Entries[0], "Password"); got != "s3cret" {
		t.Fatalf("password = %q, want %q", got, "s3cret")
	}
}

func TestPickerCtrlNOpensNewDatabaseForm(t *testing.T) {
	m := initialModel()

	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	got := after.(model)
	if got.editMode != editNewDB {
		t.Fatalf("ctrl+n should open the new database form, got editMode %v", got.editMode)
	}
	if len(got.newDBInputs) != 3 {
		t.Fatalf("expected 3 inputs, got %d", len(got.newDBInputs))
	}
}

func TestPickerFirstItemOpensNewDatabaseForm(t *testing.T) {
	m := initialModel()
	m.fileCursor = 0

	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := after.(model)
	if got.editMode != editNewDB {
		t.Fatal("enter on the first item should open the new database form")
	}
}

// TestNewDBFormCreatesDatabaseAndOpensIt walks the whole flow the user sees:
// Ctrl+N, fill the form, Enter, and end up browsing the fresh database.
func TestNewDBFormCreatesDatabaseAndOpensIt(t *testing.T) {
	t.Chdir(t.TempDir())

	m := initialModel()

	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	m = after.(model)
	if m.editMode != editNewDB {
		t.Fatalf("ctrl+n should open the new database form, got %v", m.editMode)
	}

	m.newDBInputs[0].SetValue("Fresh")
	m.newDBInputs[1].SetValue("masterpw")
	m.newDBInputs[2].SetValue("masterpw")

	after, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = after.(model)

	if m.editMode != editNone {
		t.Fatalf("the form should close after creating, got %v", m.editMode)
	}
	if m.screen != screenBrowse {
		t.Fatalf("expected the browse screen, got %v", m.screen)
	}
	if m.filePath != "Fresh.kdbx" {
		t.Fatalf("filePath = %q, want %q", m.filePath, "Fresh.kdbx")
	}
	if m.db == nil {
		t.Fatal("the new database should be loaded")
	}
	if len(m.groups) != 1 || m.groups[0].name != "General" {
		t.Fatalf("expected a single General group, got %v", m.groups)
	}
	if _, err := os.Stat("Fresh.kdbx"); err != nil {
		t.Fatalf("the database file should exist: %v", err)
	}
}

func TestNewDBFormCtrlPTogglesPasswordVisibility(t *testing.T) {
	m := initialModel().startNewDB()

	if m.newDBShowPass {
		t.Fatal("the password must start hidden")
	}
	if m.newDBInputs[1].EchoMode != textinput.EchoPassword {
		t.Fatal("the password field must start masked")
	}

	// Ctrl+P reveals both password fields.
	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	got := after.(model)
	if !got.newDBShowPass {
		t.Fatal("ctrl+p should reveal the password")
	}
	if got.newDBInputs[1].EchoMode != textinput.EchoNormal {
		t.Fatal("ctrl+p should unmask the password field")
	}
	if got.newDBInputs[2].EchoMode != textinput.EchoNormal {
		t.Fatal("ctrl+p should unmask the confirmation field")
	}
	// The file name field is never masked.
	if got.newDBInputs[0].EchoMode != textinput.EchoNormal {
		t.Fatal("the file field should stay plain")
	}

	// Ctrl+P again hides them.
	after, _ = got.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	got = after.(model)
	if got.newDBShowPass {
		t.Fatal("ctrl+p should hide the password again")
	}
	if got.newDBInputs[1].EchoMode != textinput.EchoPassword {
		t.Fatal("the password field should be masked again")
	}
}

func TestNewDBFormCtrlPKeepsTheTypedValue(t *testing.T) {
	m := initialModel().startNewDB()
	m.newDBInputs[1].SetValue("S3cret!")

	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	got := after.(model)

	if got.newDBInputs[1].Value() != "S3cret!" {
		t.Fatalf("value = %q, want %q", got.newDBInputs[1].Value(), "S3cret!")
	}
}

// A plain p must still be typed as text, never toggle the visibility.
func TestNewDBFormPlainPIsTypedNotToggled(t *testing.T) {
	m := initialModel().startNewDB()
	m.newDBCursor = 1

	after, _ := m.Update(keyRunes("p"))
	got := after.(model)

	if got.newDBShowPass {
		t.Fatal("a plain p must not toggle password visibility")
	}
	if got.newDBInputs[1].Value() != "p" {
		t.Fatalf("value = %q, want %q", got.newDBInputs[1].Value(), "p")
	}
}

func TestNewDBFormRejectsMismatchedPasswords(t *testing.T) {
	t.Chdir(t.TempDir())

	m := initialModel().startNewDB()
	m.newDBInputs[0].SetValue("x")
	m.newDBInputs[1].SetValue("aaa")
	m.newDBInputs[2].SetValue("bbb")

	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := after.(model)

	if got.editMode != editNewDB {
		t.Fatal("the form should stay open when the passwords do not match")
	}
	if _, err := os.Stat("x.kdbx"); err == nil {
		t.Fatal("no file should be created when the passwords do not match")
	}
}

func TestNewDBFormRejectsPaths(t *testing.T) {
	t.Chdir(t.TempDir())

	m := initialModel().startNewDB()
	m.newDBInputs[0].SetValue(filepath.Join("sub", "x.kdbx"))
	m.newDBInputs[1].SetValue("aaa")
	m.newDBInputs[2].SetValue("aaa")

	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := after.(model)

	if got.editMode != editNewDB {
		t.Fatal("the form should stay open when the name contains a path")
	}
	if _, err := os.Stat("sub"); err == nil {
		t.Fatal("no file should be created outside the program's directory")
	}
}

func TestNewDBFormRefusesToOverwriteExistingFile(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.WriteFile("x.kdbx", []byte("precious"), 0o600); err != nil {
		t.Fatal(err)
	}

	m := initialModel().startNewDB()
	m.newDBInputs[0].SetValue("x")
	m.newDBInputs[1].SetValue("aaa")
	m.newDBInputs[2].SetValue("aaa")

	after, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := after.(model)

	if got.editMode != editNewDB {
		t.Fatal("the form should stay open when the file already exists")
	}
	data, err := os.ReadFile("x.kdbx")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "precious" {
		t.Fatalf("existing file was overwritten: %q", string(data))
	}
}
