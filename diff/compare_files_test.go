package diff_test

import (
	"testing"

	"github.com/izabelacg/aac-vocab-diff/diff"
)

func TestCompareFiles_NavPathsConsistentWithPages(t *testing.T) {
	// Verify wiring by checking NavPath maps are nil iff the vocab has no
	// "Home" page — independent of which fixture we use.
	dbPath, cleanup, err := diff.ExtractC4V(fixtureFile)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	pages, err := diff.LoadPages(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_, hasHome := pages[diff.NavPathRootPage]

	d, err := diff.CompareFiles(fixtureFile, fixtureFile)
	if err != nil {
		t.Fatal(err)
	}
	if hasHome && d.NavPathFromOld == nil {
		t.Error("NavPathFromOld is nil but Home page exists: CompareFiles is not wiring nav path loading")
	}
	if !hasHome && d.NavPathFromOld != nil {
		t.Error("NavPathFromOld should be nil when no Home page exists")
	}
}

func TestCompareFiles_NavPathsSameFileProduceIdenticalMaps(t *testing.T) {
	// Same file on both sides → every page must have the same path in old and new.
	// This catches the bug where old/new path maps are swapped or share a pointer.
	d, err := diff.CompareFiles(fixtureFile, fixtureFile)
	if err != nil {
		t.Fatal(err)
	}
	for page, oldPath := range d.NavPathFromOld {
		if newPath := d.NavPathFromNew[page]; oldPath != newPath {
			t.Errorf("page %q: NavPathFromOld=%q NavPathFromNew=%q — want identical for same file",
				page, oldPath, newPath)
		}
	}
	for page, newPath := range d.NavPathFromNew {
		if oldPath := d.NavPathFromOld[page]; newPath != oldPath {
			t.Errorf("page %q: NavPathFromNew=%q NavPathFromOld=%q — want identical for same file",
				page, newPath, oldPath)
		}
	}
}
