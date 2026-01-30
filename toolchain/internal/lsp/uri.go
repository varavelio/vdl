package lsp

import (
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
)

// UriToPath converts a file:// URI to an absolute file path.
// It handles URL decoding and platform-specific path normalization.
// If the URI doesn't have a file:// prefix, it returns the URI as-is.
func UriToPath(uri string) string {
	return uriToPathOS(uri, runtime.GOOS)
}

// PathToUri converts an absolute file path to a file:// URI.
// It handles URL encoding and platform-specific path normalization.
func PathToUri(path string) string {
	return pathToUriOS(path, runtime.GOOS)
}

// uriToPathOS converts a URI to a path with explicit OS specification.
// This is used internally and for testing across platforms.
func uriToPathOS(uri, goos string) string {
	const scheme = "file://"

	// Check for file:// scheme (case-insensitive)
	lowerURI := strings.ToLower(uri)
	if !strings.HasPrefix(lowerURI, scheme) {
		return uri
	}

	// Preserve the rest of the URI but skip the scheme part
	uri = uri[len(scheme):]

	// Parse the URI
	parsed, err := url.Parse("file://" + uri)
	if err != nil {
		// If parsing fails, fall back to simple processing
		path := uri
		// On Windows, handle the leading slash before drive letter
		if goos == "windows" && len(path) >= 3 {
			if path[0] == '/' && len(path) >= 3 && path[2] == ':' {
				path = path[1:]
			}
		}
		// Normalize separators
		if goos == "windows" {
			path = strings.ReplaceAll(path, "/", "\\")
		}
		return cleanPath(path, goos)
	}

	// Get the path component (handles URL decoding automatically)
	path := parsed.Path

	// On Windows, handle the leading slash before drive letter
	// e.g., /C:/Users/... -> C:/Users/...
	if goos == "windows" && len(path) >= 3 {
		// Check for pattern like /C: or /c:
		if path[0] == '/' && len(path) >= 3 && path[2] == ':' {
			path = path[1:]
		}
	}

	// Normalize the path separators for the target OS
	if goos == "windows" {
		// Convert forward slashes to backslashes
		path = strings.ReplaceAll(path, "/", "\\")
	}

	// Clean the path to remove any redundant separators or dots
	path = cleanPath(path, goos)

	return path
}

// pathToUriOS converts a path to a URI with explicit OS specification.
// This is used internally and for testing across platforms.
func pathToUriOS(path, goos string) string {
	const scheme = "file://"

	// If already a URI (case-insensitive check), return as-is
	lowerPath := strings.ToLower(path)
	if strings.HasPrefix(lowerPath, scheme) {
		return path
	}

	// For testing, we need to handle paths as they would be on the target OS
	// Normalize slashes based on target OS
	if goos == "windows" {
		// On Windows, paths use backslashes
		path = strings.ReplaceAll(path, "\\", "/")
	}

	// Check if path is absolute for the target OS
	isAbs := false
	if goos == "windows" {
		// Windows absolute: C:\ or C:/ or \\server\share
		isAbs = (len(path) >= 2 && path[1] == ':') || strings.HasPrefix(path, "//")
	} else {
		// Unix absolute: starts with /
		isAbs = strings.HasPrefix(path, "/")
	}

	// If not absolute, we can't reliably convert without the actual filesystem
	// In tests, we'll just work with what we have
	if !isAbs && goos == runtime.GOOS {
		// Only try to make absolute if we're on the same OS
		absPath, err := filepath.Abs(path)
		if err == nil {
			path = absPath
			// Re-normalize after Abs
			if goos == "windows" {
				path = strings.ReplaceAll(path, "\\", "/")
			}
		}
	}

	// On Windows, ensure the path starts with a slash for proper URI format
	// e.g., C:/Users/... -> /C:/Users/...
	if goos == "windows" && len(path) >= 2 && path[1] == ':' {
		path = "/" + path
	}

	// Ensure Unix paths start with /
	if goos != "windows" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Build the URI manually to control encoding
	// We need to encode special characters but preserve the path structure
	var result strings.Builder
	result.WriteString(scheme)

	for _, c := range path {
		switch c {
		case '/':
			// Don't encode path separators
			result.WriteByte('/')
		case ':':
			// Don't encode colons (for Windows drive letters)
			result.WriteByte(':')
		default:
			// Encode if needed using url.PathEscape for individual characters
			encoded := url.PathEscape(string(c))
			result.WriteString(encoded)
		}
	}

	return result.String()
}

// cleanPath cleans a path for the specified OS, removing redundant separators and resolving . and ..
func cleanPath(path, goos string) string {
	if path == "" {
		return "."
	}

	separator := "/"
	if goos == "windows" {
		separator = "\\"
	}

	// Split path into components
	components := strings.Split(path, separator)
	var cleaned []string

	for _, comp := range components {
		if comp == "" || comp == "." {
			// Skip empty and current directory
			continue
		}
		if comp == ".." {
			// Go up one directory
			if len(cleaned) > 0 && cleaned[len(cleaned)-1] != ".." {
				cleaned = cleaned[:len(cleaned)-1]
			} else {
				cleaned = append(cleaned, comp)
			}
		} else {
			cleaned = append(cleaned, comp)
		}
	}

	if len(cleaned) == 0 {
		if strings.HasPrefix(path, separator) {
			return separator
		}
		return "."
	}

	result := strings.Join(cleaned, separator)

	// Preserve leading separator for absolute paths
	if strings.HasPrefix(path, separator) {
		result = separator + result
	}

	// On Windows, preserve drive letter format
	if goos == "windows" && len(path) >= 2 && path[1] == ':' {
		// Ensure we don't add extra separator after drive letter
		if !strings.HasPrefix(result, string(path[0])+":") {
			// Path was something like C:.. which got cleaned
			result = string(path[0]) + ":" + separator + result
		}
	}

	return result
}
