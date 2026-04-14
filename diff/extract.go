package diff

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ExtractC4V(cePath string) (dbPath string, cleanup func(), err error) {
	r, err := zip.OpenReader(cePath)
	if err != nil {
		// %w wraps the error so callers can use errors.Is/As to inspect it
		return "", nil, fmt.Errorf("open %s: %w", cePath, err)
	}
	defer r.Close()

	for _, f := range r.File {
		if !strings.HasSuffix(f.Name, ".c4v") {
			continue
		}

		// os.CreateTemp creates an empty file and returns it open for writing.
		// The "*" in the pattern is replaced with a random string.
		tmp, err := os.CreateTemp("", "vocab-*.c4v")
		if err != nil {
			return "", nil, fmt.Errorf("create temp file: %w", err)
		}

		rc, err := f.Open() // rc is the reader for this zip entry
		if err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			return "", nil, fmt.Errorf("open %s in archive: %w", f.Name, err)
		}

		if _, err := io.Copy(tmp, rc); err != nil {
			rc.Close()
			tmp.Close()
			os.Remove(tmp.Name())
			return "", nil, fmt.Errorf("copy %s: %w", f.Name, err)
		}
		rc.Close()
		tmp.Close()

		// Capture the name before the closure — if we used tmp.Name() inside
		// the func literal, Go would close over the variable, not the value.
		name := tmp.Name()
		return name, func() { os.Remove(name) }, nil
	}

	return "", nil, fmt.Errorf("no .c4v file found in %s", filepath.Base(cePath))
}
