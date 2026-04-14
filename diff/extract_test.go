package diff_test

import (
	"archive/zip"
	"os"
	"strings"
	"testing"

	"github.com/izabelacg/aac-vocab-diff/diff"
)

// go test runs with cwd = the package directory (diff/),
// so ../vocabs/ resolves to the repo-root vocabs/ folder.
const fixtureFile = "testdata/WordPower60 Basic SS_2026-04-08.ce"

func TestExtractC4V_ReturnsC4VPath(t *testing.T) {
	path, cleanup, err := diff.ExtractC4V(fixtureFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup()

	if !strings.HasSuffix(path, ".c4v") {
		t.Errorf("expected path ending in .c4v, got %q", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("extracted file does not exist: %v", err)
	}
}

func TestExtractC4V_FileNotFound(t *testing.T) {
	_, _, err := diff.ExtractC4V("nonexistent.ce")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestExtractC4V_NoC4VInArchive(t *testing.T) {
	// Build a minimal ZIP with no .c4v entry.
	f, _ := os.CreateTemp("", "no-c4v-*.ce")
	w := zip.NewWriter(f)
	w.Create("something.txt") // a file but not .c4v
	w.Close()
	f.Close()
	defer os.Remove(f.Name())

	_, _, err := diff.ExtractC4V(f.Name())
	if err == nil {
		t.Fatal("expected error when archive has no .c4v, got nil")
	}
}

func TestExtractC4V_CleanupRemovesFile(t *testing.T) {
	path, cleanup, err := diff.ExtractC4V(fixtureFile)
	if err != nil {
		t.Fatal(err)
	}
	cleanup()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("cleanup did not remove the temp file")
	}
}
