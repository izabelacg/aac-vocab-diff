package diff_test

import (
	"testing"

	"github.com/izabelacg/aac-vocab-diff/diff"
)

func TestCompareFiles_NavPathsPopulated(t *testing.T) {
	// With multi-source BFS, nav paths are always computed (non-nil map)
	// as long as the vocab has any navigate-to-page edges at all.
	d, err := diff.CompareFiles(fixtureFile, fixtureFile)
	if err != nil {
		t.Fatal(err)
	}
	if d.NavPathFromOld == nil {
		t.Error("NavPathFromOld is nil: CompareFiles is not wiring nav path loading")
	}
	if d.NavPathFromNew == nil {
		t.Error("NavPathFromNew is nil: CompareFiles is not wiring nav path loading")
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
