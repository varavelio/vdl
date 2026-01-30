package filepathutil

import (
	"net/url"
	"path/filepath"
	"strings"
)

// FromURI converts a URI to a file path.
// It handles file:// URIs, regular file paths, and Windows-specific paths.
// For URIs, it properly decodes escaped characters and handles drive letters on Windows.
func FromURI(uriOrPath string) string {
	// If it's not a URI, return as is
	if !strings.HasPrefix(uriOrPath, "file://") {
		return filepath.Clean(uriOrPath)
	}

	// Parse the URI
	u, err := url.Parse(uriOrPath)
	if err != nil || u.Scheme != "file" {
		// If parsing fails or it's not a file URI, fall back to simple trimming
		return filepath.Clean(strings.TrimPrefix(uriOrPath, "file://"))
	}

	// Convert the URI path to a file path
	path := u.Path

	// On Windows, handle drive letters correctly
	if len(path) >= 3 && path[0] == '/' && path[2] == ':' {
		// Remove leading slash for Windows drive paths: /C:/path -> C:/path
		path = path[1:]
	}

	return filepath.Clean(path)
}
