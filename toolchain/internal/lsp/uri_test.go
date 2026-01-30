package lsp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUriToPath(t *testing.T) {
	t.Run("unix simple path", func(t *testing.T) {
		result := uriToPathOS("file:///home/user/file.vdl", "linux")
		require.Equal(t, "/home/user/file.vdl", result)
	})

	t.Run("unix path with spaces", func(t *testing.T) {
		result := uriToPathOS("file:///home/user/my%20project/file.vdl", "linux")
		require.Equal(t, "/home/user/my project/file.vdl", result)
	})

	t.Run("unix path with special characters", func(t *testing.T) {
		result := uriToPathOS("file:///home/user/project%40work/file%231.vdl", "linux")
		require.Equal(t, "/home/user/project@work/file#1.vdl", result)
	})

	t.Run("unix path with unicode", func(t *testing.T) {
		result := uriToPathOS("file:///home/user/%E6%96%87%E4%BB%B6.vdl", "linux")
		require.Equal(t, "/home/user/\u6587\u4ef6.vdl", result)
	})

	t.Run("path without file scheme is returned as-is", func(t *testing.T) {
		result := uriToPathOS("/home/user/file.vdl", "linux")
		require.Equal(t, "/home/user/file.vdl", result)
	})

	t.Run("empty string", func(t *testing.T) {
		result := uriToPathOS("", "linux")
		require.Equal(t, "", result)
	})

	t.Run("file scheme only", func(t *testing.T) {
		// Should handle gracefully
		require.NotPanics(t, func() {
			_ = uriToPathOS("file://", "linux")
		})
	})

	t.Run("malformed URI is handled gracefully", func(t *testing.T) {
		// This should not panic
		result := uriToPathOS("file:///%zz", "linux")
		require.NotEmpty(t, result)
	})

	t.Run("relative path without scheme", func(t *testing.T) {
		result := uriToPathOS("relative/path/file.vdl", "linux")
		require.Equal(t, "relative/path/file.vdl", result)
	})

	t.Run("path with dots", func(t *testing.T) {
		result := uriToPathOS("file:///home/user/../other/./file.vdl", "linux")
		// filepath.Clean normalizes . and ..
		require.Equal(t, "/home/other/file.vdl", result)
	})

	t.Run("path already normalized", func(t *testing.T) {
		input := "/home/user/file.vdl"
		result := uriToPathOS(input, "linux")
		require.Equal(t, input, result)
	})

	t.Run("uri with query parameters is handled", func(t *testing.T) {
		result := uriToPathOS("file:///home/user/file.vdl?version=1", "linux")
		require.Equal(t, "/home/user/file.vdl", result)
	})

	t.Run("uri with fragment", func(t *testing.T) {
		result := uriToPathOS("file:///home/user/file.vdl#line10", "linux")
		require.Equal(t, "/home/user/file.vdl", result)
	})

	t.Run("path with only filename", func(t *testing.T) {
		result := uriToPathOS("file.vdl", "linux")
		require.Equal(t, "file.vdl", result)
	})

	t.Run("uri with uppercase FILE scheme", func(t *testing.T) {
		// Scheme should be case-insensitive
		result := uriToPathOS("FILE:///home/user/file.vdl", "linux")
		require.Equal(t, "/home/user/file.vdl", result)
	})

	t.Run("uri with mixed case FILE scheme", func(t *testing.T) {
		// Scheme should be case-insensitive
		result := uriToPathOS("FiLe:///home/user/file.vdl", "linux")
		require.Equal(t, "/home/user/file.vdl", result)
	})

	t.Run("path with backslashes on unix", func(t *testing.T) {
		// On Unix, backslashes are valid filename characters
		result := uriToPathOS("/home/user/weird\\name/file.vdl", "linux")
		require.Equal(t, "/home/user/weird\\name/file.vdl", result)
	})

	t.Run("uri with percent encoded forward slash", func(t *testing.T) {
		// %2F is an encoded forward slash
		result := uriToPathOS("file:///home/user/weird%2Fname/file.vdl", "linux")
		// This should decode to / in the filename (weird/name)
		require.Equal(t, "/home/user/weird/name/file.vdl", result)
	})

	t.Run("very long path", func(t *testing.T) {
		longPath := "file:///home/" + string(make([]byte, 1000)) // Very long path
		require.NotPanics(t, func() {
			_ = uriToPathOS(longPath, "linux")
		})
	})

	t.Run("path with null bytes is handled", func(t *testing.T) {
		// Null bytes in paths are problematic but shouldn't crash
		require.NotPanics(t, func() {
			_ = uriToPathOS("file:///home/user\x00file.vdl", "linux")
		})
	})

	t.Run("uri with double slashes in path", func(t *testing.T) {
		result := uriToPathOS("file:///home//user///file.vdl", "linux")
		// filepath.Clean should normalize multiple slashes
		require.Equal(t, "/home/user/file.vdl", result)
	})
}

func TestUriToPath_Windows(t *testing.T) {
	t.Run("windows path with drive letter", func(t *testing.T) {
		result := uriToPathOS("file:///C:/Users/Dev/file.vdl", "windows")
		require.Equal(t, "C:\\Users\\Dev\\file.vdl", result)
	})

	t.Run("windows path with lowercase drive", func(t *testing.T) {
		result := uriToPathOS("file:///c:/Users/Dev/file.vdl", "windows")
		require.Equal(t, "c:\\Users\\Dev\\file.vdl", result)
	})

	t.Run("windows path with spaces", func(t *testing.T) {
		result := uriToPathOS("file:///C:/Users/My%20Documents/file.vdl", "windows")
		require.Equal(t, "C:\\Users\\My Documents\\file.vdl", result)
	})

	t.Run("windows path normalizes separators", func(t *testing.T) {
		result := uriToPathOS("file:///C:/Users/Dev/project/file.vdl", "windows")
		require.Equal(t, "C:\\Users\\Dev\\project\\file.vdl", result)
	})

	t.Run("windows path with UNC", func(t *testing.T) {
		// UNC paths like \\server\share are tricky
		// file://server/share/file.vdl has "server" as host
		result := uriToPathOS("file://server/share/file.vdl", "windows")
		// The path component will be /share/file.vdl converted to \share\file.vdl
		require.Equal(t, "\\share\\file.vdl", result)
	})

	t.Run("windows relative path", func(t *testing.T) {
		result := uriToPathOS("folder\\file.vdl", "windows")
		require.Equal(t, "folder\\file.vdl", result)
	})

	t.Run("windows path with dots", func(t *testing.T) {
		result := uriToPathOS("file:///C:/Users/../Dev/./file.vdl", "windows")
		require.Equal(t, "C:\\Dev\\file.vdl", result)
	})
}

func TestPathToUri(t *testing.T) {
	t.Run("unix simple path", func(t *testing.T) {
		result := pathToUriOS("/home/user/file.vdl", "linux")
		require.Equal(t, "file:///home/user/file.vdl", result)
	})

	t.Run("unix path with spaces", func(t *testing.T) {
		result := pathToUriOS("/home/user/my project/file.vdl", "linux")
		require.Equal(t, "file:///home/user/my%20project/file.vdl", result)
	})

	t.Run("unix path with special characters", func(t *testing.T) {
		result := pathToUriOS("/home/user/project@work/file#1.vdl", "linux")
		// @ is not encoded, # is encoded as %23
		require.Equal(t, "file:///home/user/project@work/file%231.vdl", result)
	})

	t.Run("already a URI is returned as-is", func(t *testing.T) {
		result := pathToUriOS("file:///home/user/file.vdl", "linux")
		require.Equal(t, "file:///home/user/file.vdl", result)
	})

	t.Run("empty string", func(t *testing.T) {
		// Should handle gracefully
		require.NotPanics(t, func() {
			_ = pathToUriOS("", "linux")
		})
	})

	t.Run("relative path", func(t *testing.T) {
		// Relative paths should be converted to absolute
		// This test uses PathToUri directly because it needs runtime's filepath.Abs
		result := PathToUri("relative/path/file.vdl")
		require.Contains(t, result, "file://")
		// Should be converted to absolute path, so it won't literally contain "relative/path"
		// but it will contain the filename
		require.Contains(t, result, "file.vdl")
	})

	t.Run("path with dots", func(t *testing.T) {
		result := pathToUriOS("/home/user/../other/./file.vdl", "linux")
		// Dots are not resolved by pathToUriOS, they're preserved as-is
		require.Equal(t, "file:///home/user/../other/./file.vdl", result)
	})

	t.Run("path with unicode characters", func(t *testing.T) {
		result := pathToUriOS("/home/user/文件.vdl", "linux")
		// Unicode should be percent-encoded
		require.Equal(t, "file:///home/user/%E6%96%87%E4%BB%B6.vdl", result)
	})

	t.Run("path with parentheses", func(t *testing.T) {
		result := pathToUriOS("/home/user/project (copy)/file.vdl", "linux")
		// Parentheses should be encoded
		require.Equal(t, "file:///home/user/project%20%28copy%29/file.vdl", result)
	})

	t.Run("path with ampersand", func(t *testing.T) {
		result := pathToUriOS("/home/user/R&D/file.vdl", "linux")
		// Ampersand is not encoded in path segments per RFC 3986
		require.Equal(t, "file:///home/user/R&D/file.vdl", result)
	})

	t.Run("path with equals sign", func(t *testing.T) {
		result := pathToUriOS("/home/user/a=b/file.vdl", "linux")
		// Equals sign is not encoded in path segments per RFC 3986
		require.Equal(t, "file:///home/user/a=b/file.vdl", result)
	})

	t.Run("path with question mark", func(t *testing.T) {
		result := pathToUriOS("/home/user/what?/file.vdl", "linux")
		// Question mark should be encoded
		require.Equal(t, "file:///home/user/what%3F/file.vdl", result)
	})

	t.Run("single slash", func(t *testing.T) {
		result := pathToUriOS("/", "linux")
		require.Equal(t, "file:///", result)
	})

	t.Run("path with plus sign", func(t *testing.T) {
		result := pathToUriOS("/home/user/C++/file.vdl", "linux")
		// Plus sign is not encoded in path segments per RFC 3986
		require.Equal(t, "file:///home/user/C++/file.vdl", result)
	})

	t.Run("path with tilde", func(t *testing.T) {
		result := pathToUriOS("/home/user/~temp/file.vdl", "linux")
		// Tilde is unreserved and not encoded per RFC 3986
		require.Equal(t, "file:///home/user/~temp/file.vdl", result)
	})

	t.Run("path with comma", func(t *testing.T) {
		result := pathToUriOS("/home/user/a,b,c/file.vdl", "linux")
		require.Equal(t, "file:///home/user/a%2Cb%2Cc/file.vdl", result)
	})

	t.Run("path already in uri format", func(t *testing.T) {
		input := "file:///home/user/file.vdl"
		result := pathToUriOS(input, "linux")
		require.Equal(t, input, result)
	})

	t.Run("uppercase FILE scheme is preserved", func(t *testing.T) {
		// If already a URI with uppercase scheme, return as-is
		input := "FILE:///home/user/file.vdl"
		result := pathToUriOS(input, "linux")
		require.Equal(t, input, result)
	})
}

func TestPathToUri_Windows(t *testing.T) {
	t.Run("windows path with drive letter", func(t *testing.T) {
		result := pathToUriOS("C:\\Users\\Dev\\file.vdl", "windows")
		require.Equal(t, "file:///C:/Users/Dev/file.vdl", result)
	})

	t.Run("windows path with spaces", func(t *testing.T) {
		result := pathToUriOS("C:\\Users\\My Documents\\file.vdl", "windows")
		require.Equal(t, "file:///C:/Users/My%20Documents/file.vdl", result)
	})

	t.Run("windows path with forward slashes", func(t *testing.T) {
		result := pathToUriOS("C:/Users/Dev/file.vdl", "windows")
		require.Equal(t, "file:///C:/Users/Dev/file.vdl", result)
	})

	t.Run("windows relative path", func(t *testing.T) {
		// Relative paths should be converted to absolute
		// This uses PathToUri directly because it needs runtime's filepath.Abs
		result := PathToUri("folder\\subfolder\\file.vdl")
		require.Contains(t, result, "file://")
		// Should be converted to absolute path
		require.NotEqual(t, "file:///folder/subfolder/file.vdl", result)
	})

	t.Run("windows path with special characters", func(t *testing.T) {
		result := pathToUriOS("C:\\Users\\Dev (1)\\file.vdl", "windows")
		require.Equal(t, "file:///C:/Users/Dev%20%281%29/file.vdl", result)
	})

	t.Run("windows path with dots", func(t *testing.T) {
		result := pathToUriOS("C:\\Users\\..\\Dev\\.\\file.vdl", "windows")
		// Dots are preserved as-is in pathToUriOS
		require.Equal(t, "file:///C:/Users/../Dev/./file.vdl", result)
	})
}

func TestRoundTrip(t *testing.T) {
	t.Run("unix roundtrip", func(t *testing.T) {
		original := "/home/user/my project/file.vdl"
		uri := pathToUriOS(original, "linux")
		path := uriToPathOS(uri, "linux")
		require.Equal(t, original, path)
	})

	t.Run("unix uri roundtrip", func(t *testing.T) {
		original := "file:///home/user/my%20project/file.vdl"
		path := uriToPathOS(original, "linux")
		uri := pathToUriOS(path, "linux")
		require.Equal(t, original, uri)
	})

	t.Run("unix path with special chars roundtrip", func(t *testing.T) {
		original := "/home/user/my (project)/file#1.vdl"
		uri := pathToUriOS(original, "linux")
		path := uriToPathOS(uri, "linux")
		require.Equal(t, original, path)
	})

	t.Run("unix unicode roundtrip", func(t *testing.T) {
		original := "/home/user/文件夹/file.vdl"
		uri := pathToUriOS(original, "linux")
		path := uriToPathOS(uri, "linux")
		require.Equal(t, original, path)
	})
}

func TestRoundTrip_Windows(t *testing.T) {
	t.Run("windows roundtrip", func(t *testing.T) {
		original := "C:\\Users\\My Documents\\file.vdl"
		uri := pathToUriOS(original, "windows")
		path := uriToPathOS(uri, "windows")
		require.Equal(t, original, path)
	})

	t.Run("windows uri roundtrip", func(t *testing.T) {
		original := "file:///C:/Users/My%20Documents/file.vdl"
		path := uriToPathOS(original, "windows")
		uri := pathToUriOS(path, "windows")
		require.Equal(t, original, uri)
	})

	t.Run("windows path with special chars roundtrip", func(t *testing.T) {
		original := "C:\\Users\\Dev (1)\\file#1.vdl"
		uri := pathToUriOS(original, "windows")
		path := uriToPathOS(uri, "windows")
		require.Equal(t, original, path)
	})
}

func TestUriToPath_CrossPlatform(t *testing.T) {
	// These tests verify behavior that should work on any platform

	t.Run("handles double encoded spaces", func(t *testing.T) {
		// %2520 is double-encoded space (%25 = %, 20 = space)
		result := uriToPathOS("file:///home/user/my%2520project/file.vdl", "linux")
		// Should decode only once: %2520 -> %20
		require.Equal(t, "/home/user/my%20project/file.vdl", result)
	})

	t.Run("preserves trailing slash", func(t *testing.T) {
		// filepath.Clean removes trailing slashes, which is expected behavior
		result := uriToPathOS("file:///home/user/project/", "linux")
		require.Equal(t, "/home/user/project", result)
	})

	t.Run("handles file scheme with single slash", func(t *testing.T) {
		// file:/ instead of file:///
		// This is malformed but our implementation treats it as "not a file:// URI"
		result := uriToPathOS("file:/home/user/file.vdl", "linux")
		// It won't match "file://" so it returns as-is
		require.Equal(t, "file:/home/user/file.vdl", result)
	})

	t.Run("handles file scheme with double slash", func(t *testing.T) {
		// file:// instead of file:///
		result := uriToPathOS("file://home/user/file.vdl", "linux")
		// With only two slashes, "home" becomes the host, path is /user/file.vdl
		require.Equal(t, "/user/file.vdl", result)
	})

	t.Run("non-file scheme is treated as path", func(t *testing.T) {
		// http:// or other schemes should be treated as-is
		result := uriToPathOS("http://example.com/file.vdl", "linux")
		require.Equal(t, "http://example.com/file.vdl", result)
	})

	t.Run("mixed case in path", func(t *testing.T) {
		result := uriToPathOS("file:///home/User/MyFile.vdl", "linux")
		require.Equal(t, "/home/User/MyFile.vdl", result)
	})

	t.Run("path with underscore", func(t *testing.T) {
		result := uriToPathOS("file:///home/user/my_file.vdl", "linux")
		require.Equal(t, "/home/user/my_file.vdl", result)
	})

	t.Run("path with hyphen", func(t *testing.T) {
		result := uriToPathOS("file:///home/user/my-file.vdl", "linux")
		require.Equal(t, "/home/user/my-file.vdl", result)
	})

	t.Run("path with numbers", func(t *testing.T) {
		result := uriToPathOS("file:///home/user123/file456.vdl", "linux")
		require.Equal(t, "/home/user123/file456.vdl", result)
	})
}

func TestPathToUri_CrossPlatform(t *testing.T) {
	t.Run("handles path with multiple consecutive slashes", func(t *testing.T) {
		// PathToUri doesn't normalize multiple slashes, they're preserved
		result := pathToUriOS("/home//user///file.vdl", "linux")
		require.Equal(t, "file:///home//user///file.vdl", result)
	})

	t.Run("dot file", func(t *testing.T) {
		result := pathToUriOS("/home/user/.config/file.vdl", "linux")
		require.Equal(t, "file:///home/user/.config/file.vdl", result)
	})

	t.Run("hidden directory", func(t *testing.T) {
		result := pathToUriOS("/home/user/.hidden/file.vdl", "linux")
		require.Equal(t, "file:///home/user/.hidden/file.vdl", result)
	})

	t.Run("path with file extension containing numbers", func(t *testing.T) {
		result := pathToUriOS("/home/user/file.mp3", "linux")
		require.Equal(t, "file:///home/user/file.mp3", result)
	})

	t.Run("path with multiple dots", func(t *testing.T) {
		result := pathToUriOS("/home/user/file.test.vdl", "linux")
		require.Equal(t, "file:///home/user/file.test.vdl", result)
	})

	t.Run("path with no extension", func(t *testing.T) {
		result := pathToUriOS("/home/user/README", "linux")
		require.Equal(t, "file:///home/user/README", result)
	})
}

// TestEdgeCases tests additional edge cases that might occur in practice
func TestEdgeCases(t *testing.T) {
	t.Run("UriToPath with percent in filename", func(t *testing.T) {
		// A literal percent sign that's not part of encoding
		result := uriToPathOS("file:///home/user/100%25complete.vdl", "linux")
		require.Equal(t, "/home/user/100%complete.vdl", result)
	})

	t.Run("PathToUri with already encoded characters", func(t *testing.T) {
		// Path shouldn't contain encoded chars, but if it does, they should be double-encoded
		result := pathToUriOS("/home/user/file%20name.vdl", "linux")
		// The percent should be encoded as %25, so %20 becomes %2520
		require.Equal(t, "file:///home/user/file%2520name.vdl", result)
	})

	t.Run("UriToPath with localhost in URI", func(t *testing.T) {
		// file://localhost/path is valid
		result := uriToPathOS("file://localhost/home/user/file.vdl", "linux")
		// localhost as host, path is /home/user/file.vdl
		require.Equal(t, "/home/user/file.vdl", result)
	})

	t.Run("very short path", func(t *testing.T) {
		result := pathToUriOS("/a", "linux")
		require.Equal(t, "file:///a", result)
	})

	t.Run("path with only spaces in directory name", func(t *testing.T) {
		// Weird but technically valid
		result := pathToUriOS("/home/   /file.vdl", "linux")
		require.Equal(t, "file:///home/%20%20%20/file.vdl", result)
	})

	t.Run("consecutive encoded characters", func(t *testing.T) {
		result := uriToPathOS("file:///home/user/%40%23%24.vdl", "linux")
		require.Equal(t, "/home/user/@#$.vdl", result)
	})

	t.Run("path with bracket characters", func(t *testing.T) {
		result := pathToUriOS("/home/user/[test]/file.vdl", "linux")
		require.Equal(t, "file:///home/user/%5Btest%5D/file.vdl", result)
	})

	t.Run("path with curly braces", func(t *testing.T) {
		result := pathToUriOS("/home/user/{test}/file.vdl", "linux")
		require.Equal(t, "file:///home/user/%7Btest%7D/file.vdl", result)
	})

	t.Run("path with pipe character", func(t *testing.T) {
		result := pathToUriOS("/home/user/a|b/file.vdl", "linux")
		require.Equal(t, "file:///home/user/a%7Cb/file.vdl", result)
	})

	t.Run("path with less than and greater than", func(t *testing.T) {
		result := pathToUriOS("/home/user/a<b>c/file.vdl", "linux")
		require.Equal(t, "file:///home/user/a%3Cb%3Ec/file.vdl", result)
	})
}
