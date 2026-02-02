package strutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFuzzySearch(t *testing.T) {
	t.Run("empty data returns nil and no exact match", func(t *testing.T) {
		matches, exact := FuzzySearch(nil, "test")
		require.Nil(t, matches)
		require.False(t, exact)

		matches, exact = FuzzySearch([]string{}, "test")
		require.Nil(t, matches)
		require.False(t, exact)
	})

	t.Run("exact match found is case-sensitive and literal", func(t *testing.T) {
		data := []string{"Hello", "hello", "HELLO"}

		_, exact := FuzzySearch(data, "Hello")
		require.True(t, exact)

		_, exact = FuzzySearch(data, "hello")
		require.True(t, exact)

		_, exact = FuzzySearch(data, "HeLLo")
		require.False(t, exact)
	})

	t.Run("returns at most 3 results", func(t *testing.T) {
		data := []string{"cat", "car", "bat", "cap", "can", "cab"}

		matches, _ := FuzzySearch(data, "cat")
		require.Len(t, matches, maxFuzzyResults)
	})

	t.Run("exact normalized match ranks first", func(t *testing.T) {
		data := []string{"helloo", "hallo", "hello", "help", "helo"}

		matches, _ := FuzzySearch(data, "hello")
		require.Len(t, matches, 3)
		require.Equal(t, "hello", matches[0]) // exact match
		// helloo is prefix match (hello is prefix of helloo), ranks second
		require.Equal(t, "helloo", matches[1])
	})

	t.Run("prefix matches rank higher than edit distance", func(t *testing.T) {
		// Query "test" - "tests" and "testing" are prefix matches
		data := []string{"text", "tests", "testing", "test"}

		matches, _ := FuzzySearch(data, "test")
		require.Equal(t, "test", matches[0])    // exact
		require.Equal(t, "tests", matches[1])   // prefix
		require.Equal(t, "testing", matches[2]) // prefix (longer)
	})

	t.Run("suffix matches included even beyond edit distance", func(t *testing.T) {
		// Query "User" (4 chars, maxDist=1)
		// "BaseUser" is 8 chars, diff=4 > maxDist, but it's a suffix match
		data := []string{"BaseUser", "Users", "User", "Usr"}

		matches, _ := FuzzySearch(data, "User")
		require.Contains(t, matches, "User")
		require.Contains(t, matches, "Users")    // prefix match
		require.Contains(t, matches, "BaseUser") // suffix match
	})

	t.Run("contains matches included", func(t *testing.T) {
		// Query "User" contained in "SuperUserAdmin"
		data := []string{"SuperUserAdmin", "UserProfile", "User", "Admin"}

		matches, _ := FuzzySearch(data, "User")
		require.Equal(t, "User", matches[0]) // exact
		require.Contains(t, matches, "UserProfile")
		require.Contains(t, matches, "SuperUserAdmin")
	})

	t.Run("transpositions rank higher than substitutions", func(t *testing.T) {
		// Query "uint" (4 chars) -> distance 1
		// This tests the "Did you mean 'unit'?" scenario
		data := []string{"unit", "hint", "uint", "init"}

		matches, exact := FuzzySearch(data, "uint")
		require.True(t, exact)
		require.Contains(t, matches, "uint")
		require.Contains(t, matches, "unit")
	})

	t.Run("short query uses distance 1", func(t *testing.T) {
		// Query "cat" (3 chars) -> distance 1
		data := []string{"cat", "car", "bat", "dog"}

		matches, _ := FuzzySearch(data, "cat")
		require.Len(t, matches, 3)
		require.Equal(t, "cat", matches[0])
	})

	t.Run("medium query uses distance 2", func(t *testing.T) {
		// Query "hello" (5 chars) -> distance 2
		data := []string{"hello", "hallo", "help", "world"}

		matches, _ := FuzzySearch(data, "hello")
		require.Len(t, matches, 3)
		require.Equal(t, "hello", matches[0])
	})

	t.Run("long query uses distance 3", func(t *testing.T) {
		// Query "programming" (11 chars) -> distance 3
		data := []string{"programming", "programing", "programmin", "developer"}

		matches, _ := FuzzySearch(data, "programming")
		require.Len(t, matches, 3)
		require.Equal(t, "programming", matches[0])
	})

	t.Run("normalization handles accents and diacritics", func(t *testing.T) {
		data := []string{"cafe", "café", "CAFÉ", "Cafe"}

		matches, _ := FuzzySearch(data, "cafe")
		// All normalize to "cafe", so all match
		require.Len(t, matches, 3)

		_, exact := FuzzySearch(data, "café")
		require.True(t, exact)
	})

	t.Run("normalization handles whitespace trimming", func(t *testing.T) {
		data := []string{"  test  ", "test", " TEST "}

		matches, _ := FuzzySearch(data, "test")
		require.Len(t, matches, 3)
	})

	t.Run("empty query string matches only empty strings", func(t *testing.T) {
		data := []string{"", "a", "ab", "abc"}

		matches, exact := FuzzySearch(data, "")
		require.Len(t, matches, 1)
		require.Equal(t, "", matches[0])
		require.True(t, exact)
	})

	t.Run("single character strings", func(t *testing.T) {
		data := []string{"a", "b", "c", "d", "e"}

		// Query "a" (1 char) -> distance 1, all single chars match
		matches, _ := FuzzySearch(data, "a")
		require.Len(t, matches, 3)
		require.Equal(t, "a", matches[0]) // exact match first
	})

	t.Run("unicode characters", func(t *testing.T) {
		data := []string{"日本語", "日本", "日本人", "中国語"}

		// "日本語" (3 runes) -> distance 1
		matches, exact := FuzzySearch(data, "日本語")
		require.Len(t, matches, 3)
		require.Equal(t, "日本語", matches[0])
		require.True(t, exact)
	})

	t.Run("transpositions count as distance 1", func(t *testing.T) {
		// This is the key difference between Damerau-Levenshtein and Levenshtein
		data := []string{"ab", "ba", "abc", "acb"}

		// Query "ab" (2 chars) -> distance 1
		matches, _ := FuzzySearch(data, "ab")
		require.Len(t, matches, 3)
		require.Equal(t, "ab", matches[0])
		// abc is a prefix match
		require.Equal(t, "abc", matches[1])
	})

	t.Run("transpositions in longer strings", func(t *testing.T) {
		data := []string{"receive", "recieve", "receiver"}

		// Query "receive" (7 chars) -> distance 2
		matches, _ := FuzzySearch(data, "receive")
		require.Len(t, matches, 3)
		require.Equal(t, "receive", matches[0])
		// receiver is prefix match
		require.Equal(t, "receiver", matches[1])
		require.Contains(t, matches, "recieve")
	})

	t.Run("no matches when all words too different", func(t *testing.T) {
		data := []string{"apple", "banana", "cherry"}

		// Query "xyz" (3 chars) -> distance 1, none match
		matches, _ := FuzzySearch(data, "xyz")
		require.Empty(t, matches)
	})

	t.Run("fewer than 3 results when data is small", func(t *testing.T) {
		data := []string{"test"}

		matches, _ := FuzzySearch(data, "test")
		require.Len(t, matches, 1)
		require.Equal(t, []string{"test"}, matches)
	})

	t.Run("duplicate entries in data", func(t *testing.T) {
		data := []string{"test", "test", "test"}

		matches, exact := FuzzySearch(data, "test")
		require.True(t, exact)
		require.Len(t, matches, 3)
	})

	t.Run("special characters", func(t *testing.T) {
		data := []string{"hello-world", "hello_world", "hello world", "helloworld"}

		// "hello-world" (11 chars) -> distance 3
		matches, _ := FuzzySearch(data, "hello-world")
		require.Len(t, matches, 3)
		require.Equal(t, "hello-world", matches[0])
	})

	t.Run("concurrent processing for large datasets", func(t *testing.T) {
		// Generate data larger than concurrencyThreshold
		data := make([]string, 100)
		for i := range data {
			data[i] = "word"
		}
		data[50] = "test"
		data[51] = "testing" // prefix match

		matches, _ := FuzzySearch(data, "test")
		require.LessOrEqual(t, len(matches), 3)
		require.Equal(t, "test", matches[0])
	})

	// New tests for prefix/suffix/contains functionality

	t.Run("prefix match skips edit distance calculation", func(t *testing.T) {
		// "UserRepository" starts with "User" - should match even though
		// length diff (14-4=10) exceeds maxDist (1 for 4 chars)
		data := []string{"UserRepository", "AdminService", "User"}

		matches, _ := FuzzySearch(data, "User")
		require.Contains(t, matches, "User")
		require.Contains(t, matches, "UserRepository")
	})

	t.Run("suffix match skips edit distance calculation", func(t *testing.T) {
		// "IUserService" ends with "Service" - should match
		data := []string{"IUserService", "AdminController", "Service"}

		matches, _ := FuzzySearch(data, "Service")
		require.Contains(t, matches, "Service")
		require.Contains(t, matches, "IUserService")
	})

	t.Run("contains match skips edit distance calculation", func(t *testing.T) {
		// "getUserById" contains "User" - should match
		data := []string{"getUserById", "AdminService", "User"}

		matches, _ := FuzzySearch(data, "User")
		require.Contains(t, matches, "User")
		require.Contains(t, matches, "getUserById")
	})

	t.Run("priority order exact > prefix > suffix > contains > edit", func(t *testing.T) {
		data := []string{
			"MyUserHelper", // contains "User"
			"SuperUser",    // suffix "User"
			"UserService",  // prefix "User"
			"User",         // exact
			"Usr",          // edit distance
		}

		matches, _ := FuzzySearch(data, "User")
		require.Len(t, matches, 3)
		require.Equal(t, "User", matches[0])        // exact
		require.Equal(t, "UserService", matches[1]) // prefix
		require.Equal(t, "SuperUser", matches[2])   // suffix
	})

	t.Run("abbreviation as prefix finds full names", func(t *testing.T) {
		// User types "Cfg" looking for "Config" or "Configuration"
		data := []string{"Configuration", "Config", "Settings", "CfgManager"}

		matches, _ := FuzzySearch(data, "Cfg")
		// CfgManager starts with Cfg
		require.Contains(t, matches, "CfgManager")
	})

	t.Run("case insensitive structural matching", func(t *testing.T) {
		data := []string{"UserService", "USERSERVICE", "userservice"}

		matches, _ := FuzzySearch(data, "user")
		// All should match as prefix (normalized)
		require.Len(t, matches, 3)
	})

	t.Run("structural matches with diacritics", func(t *testing.T) {
		data := []string{"CaféService", "CafeManager", "Café"}

		matches, _ := FuzzySearch(data, "cafe")
		require.Len(t, matches, 3)
		// All normalize and match
	})

	t.Run("edit distance still works when no structural match", func(t *testing.T) {
		// "Sttring" has typo, no structural match with "String"
		data := []string{"String", "Integer", "Boolean"}

		matches, _ := FuzzySearch(data, "Sttring")
		require.Contains(t, matches, "String") // edit distance match
	})

	t.Run("short abbreviation finds longer matches", func(t *testing.T) {
		// Common pattern: user types short form
		data := []string{"StringBuilder", "StringBuffer", "String", "Integer"}

		matches, _ := FuzzySearch(data, "Str")
		// All String* should match as prefix
		require.Contains(t, matches, "String")
		require.Contains(t, matches, "StringBuilder")
		require.Contains(t, matches, "StringBuffer")
	})
}
