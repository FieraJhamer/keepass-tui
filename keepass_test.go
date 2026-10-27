package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tobischo/gokeepasslib/v3"
)

func testDB() *gokeepasslib.Database {
	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials("master")
	db.Content.Meta.DatabaseName = "test"
	db.Content.Root.Groups = []gokeepasslib.Group{
		{
			Name:    "General",
			Entries: []gokeepasslib.Entry{newEntry("Mail", "user@example.com", "s3cret", "https://example.com", "note", 0)},
		},
	}
	return db
}

func TestSaveRoundTripKeepsPasswordReadable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.kdbx")
	db := testDB()

	if err := saveDatabase(db, path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// The in-memory db must still hold plaintext, otherwise the UI shows the
	// encrypted value and a second save would double-encrypt it.
	got := getEntryValue(&db.Content.Root.Groups[0].Entries[0], "Password")
	if got != "s3cret" {
		t.Fatalf("after save, in-memory password = %q, want %q", got, "s3cret")
	}

	// Saving twice must not corrupt the value.
	if err := saveDatabase(db, path); err != nil {
		t.Fatalf("second save failed: %v", err)
	}
	got = getEntryValue(&db.Content.Root.Groups[0].Entries[0], "Password")
	if got != "s3cret" {
		t.Fatalf("after second save, in-memory password = %q, want %q", got, "s3cret")
	}

	reopened, err := openDatabase(path, "master")
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	got = getEntryValue(&reopened.Content.Root.Groups[0].Entries[0], "Password")
	if got != "s3cret" {
		t.Fatalf("reopened password = %q, want %q", got, "s3cret")
	}
}

func TestSaveKeepsOriginalOnEncodeFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.kdbx")
	db := testDB()
	if err := saveDatabase(db, path); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// A bad path makes os.Create fail; the existing file must survive.
	if err := saveDatabase(db, filepath.Join(dir, "missing", "test.kdbx")); err == nil {
		t.Fatal("expected save to fail for a missing directory")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(original) != string(after) {
		t.Fatal("failed save modified the existing database")
	}
}
