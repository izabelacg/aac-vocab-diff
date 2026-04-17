package diff

import "path/filepath"

// CompareFiles runs the full Extract → Load → Compare pipeline on two .ce
// file paths and returns a Diff with OldLabel and NewLabel set from the file
// basenames. The caller may override those fields afterwards when the paths
// are not meaningful (e.g. temp files written by an HTTP upload handler).
func CompareFiles(oldCE, newCE string) (Diff, error) {
	oldDB, cleanOld, err := ExtractC4V(oldCE)
	if err != nil {
		return Diff{}, err
	}
	defer cleanOld()

	newDB, cleanNew, err := ExtractC4V(newCE)
	if err != nil {
		return Diff{}, err
	}
	defer cleanNew()

	oldPages, err := LoadPages(oldDB)
	if err != nil {
		return Diff{}, err
	}
	newPages, err := LoadPages(newDB)
	if err != nil {
		return Diff{}, err
	}
	oldBtns, err := LoadButtons(oldDB)
	if err != nil {
		return Diff{}, err
	}
	newBtns, err := LoadButtons(newDB)
	if err != nil {
		return Diff{}, err
	}

	d := ComputeDiff(oldBtns, newBtns, oldPages, newPages)
	d.OldLabel = filepath.Base(oldCE)
	d.NewLabel = filepath.Base(newCE)

	gOld, err := LoadPageNavGraph(oldDB)
	if err != nil {
		return Diff{}, err
	}
	gNew, err := LoadPageNavGraph(newDB)
	if err != nil {
		return Diff{}, err
	}

	d.NavPathFromOld = AllShortestPathsFromHome(gOld, oldPages, NavPathRootPage)
	d.NavPathFromNew = AllShortestPathsFromHome(gNew, newPages, NavPathRootPage)
	return d, nil
}
