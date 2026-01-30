package filepathutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// NormalizeFromWD converts a relative path to a canonical absolute path
// using the current working directory as the base.
//
// If relativePath is already absolute, it is returned as is.
func NormalizeFromWD(relativePath string) (string, error) {
	// Get working dir
	wd, err := os.Getwd()
	if err != nil {
		return relativePath, fmt.Errorf("failed to get working directory: %w", err)
	}

	// Required because Normalize requires a file path and
	// not a directory path
	dummyFile := filepath.Join(wd, "dummy.txt")

	// Join and normalize the working dir and the relative path
	return Normalize(dummyFile, relativePath)
}
