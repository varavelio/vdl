package strutil

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFuzzySearch(t *testing.T) {
	t.Run("empty data returns nil and no exact match", func(t *testing.T) {
		matches, exact := FuzzySearch(nil, "test", 2)
		require.Nil(t, matches)
		require.False(t, exact)

		matches, exact = FuzzySearch([]string{}, "test", 2)
		require.Nil(t, matches)
		require.False(t, exact)
	})

	t.Run("exact match found is case-sensitive and literal", func(t *testing.T) {
		data := []string{"Hello", "hello", "HELLO"}

		_, exact := FuzzySearch(data, "Hello", 0)
		require.True(t, exact)

		_, exact = FuzzySearch(data, "hello", 0)
		require.True(t, exact)

		_, exact = FuzzySearch(data, "HeLLo", 0)
		require.False(t, exact)

		_, exact = FuzzySearch(data, "hElLo", 0)
		require.False(t, exact)
	})

	t.Run("fuzzy matches with zero distance require normalized equality", func(t *testing.T) {
		data := []string{"apple", "banana", "cherry"}

		matches, _ := FuzzySearch(data, "apple", 0)
		require.Len(t, matches, 1)
		require.Contains(t, matches, "apple")

		matches, _ = FuzzySearch(data, "APPLE", 0)
		require.Len(t, matches, 1)
		require.Contains(t, matches, "apple")

		matches, _ = FuzzySearch(data, "grape", 0)
		require.Empty(t, matches)
	})

	t.Run("fuzzy matches with distance of 1", func(t *testing.T) {
		data := []string{"cat", "car", "bat", "cap", "cut"}

		matches, _ := FuzzySearch(data, "cat", 1)
		slices.Sort(matches)
		require.Equal(t, []string{"bat", "cap", "car", "cat", "cut"}, matches)

		matches, _ = FuzzySearch(data, "cab", 1)
		slices.Sort(matches)
		require.Equal(t, []string{"cap", "car", "cat"}, matches)
	})

	t.Run("fuzzy matches with distance of 2", func(t *testing.T) {
		data := []string{"hello", "hallo", "hullo", "help", "world"}

		matches, _ := FuzzySearch(data, "hello", 2)
		slices.Sort(matches)
		require.Equal(t, []string{"hallo", "hello", "help", "hullo"}, matches)
	})

	t.Run("no matches when distance exceeds maxDist", func(t *testing.T) {
		data := []string{"programming", "development", "engineering"}

		matches, _ := FuzzySearch(data, "xyz", 2)
		require.Empty(t, matches)
	})

	t.Run("normalization handles accents and diacritics", func(t *testing.T) {
		data := []string{"cafe", "café", "CAFÉ", "Cafe"}

		matches, _ := FuzzySearch(data, "cafe", 0)
		require.Len(t, matches, 4)

		matches, _ = FuzzySearch(data, "café", 0)
		require.Len(t, matches, 4)

		_, exact := FuzzySearch(data, "café", 0)
		require.True(t, exact)

		_, exact = FuzzySearch(data, "cafe", 0)
		require.True(t, exact)
	})

	t.Run("normalization handles whitespace trimming", func(t *testing.T) {
		data := []string{"  test  ", "test", " TEST "}

		matches, _ := FuzzySearch(data, "test", 0)
		require.Len(t, matches, 3)

		matches, _ = FuzzySearch(data, "  test  ", 0)
		require.Len(t, matches, 3)
	})

	t.Run("early return when length difference exceeds maxDist", func(t *testing.T) {
		data := []string{"a", "ab", "abcdefghij"}

		matches, _ := FuzzySearch(data, "abc", 1)
		slices.Sort(matches)
		require.Equal(t, []string{"ab"}, matches)

		matches, _ = FuzzySearch(data, "abc", 2)
		slices.Sort(matches)
		require.Equal(t, []string{"a", "ab"}, matches)
	})

	t.Run("empty query string", func(t *testing.T) {
		data := []string{"", "a", "ab", "abc"}

		matches, exact := FuzzySearch(data, "", 0)
		require.Len(t, matches, 1)
		require.Contains(t, matches, "")
		require.True(t, exact)

		matches, _ = FuzzySearch(data, "", 1)
		slices.Sort(matches)
		require.Equal(t, []string{"", "a"}, matches)

		matches, _ = FuzzySearch(data, "", 2)
		slices.Sort(matches)
		require.Equal(t, []string{"", "a", "ab"}, matches)
	})

	t.Run("empty strings in data", func(t *testing.T) {
		data := []string{"", "test", ""}

		matches, _ := FuzzySearch(data, "", 0)
		require.Len(t, matches, 2)

		matches, _ = FuzzySearch(data, "a", 1)
		slices.Sort(matches)
		require.Equal(t, []string{"", ""}, matches)
	})

	t.Run("single character strings", func(t *testing.T) {
		data := []string{"a", "b", "c", "d", "e"}

		matches, _ := FuzzySearch(data, "a", 0)
		require.Equal(t, []string{"a"}, matches)

		matches, _ = FuzzySearch(data, "a", 1)
		slices.Sort(matches)
		require.Equal(t, []string{"a", "b", "c", "d", "e"}, matches)

		matches, _ = FuzzySearch(data, "x", 1)
		slices.Sort(matches)
		require.Equal(t, []string{"a", "b", "c", "d", "e"}, matches)
	})

	t.Run("unicode characters", func(t *testing.T) {
		data := []string{"日本語", "日本", "日本人", "中国語"}

		matches, exact := FuzzySearch(data, "日本語", 0)
		require.Len(t, matches, 1)
		require.Contains(t, matches, "日本語")
		require.True(t, exact)

		matches, _ = FuzzySearch(data, "日本語", 1)
		slices.Sort(matches)
		require.Equal(t, []string{"日本", "日本人", "日本語"}, matches)
	})

	t.Run("exact match independent of fuzzy matches", func(t *testing.T) {
		data := []string{"Test", "test", "TEST"}

		matches, exact := FuzzySearch(data, "test", 0)
		require.True(t, exact)
		require.Len(t, matches, 3)

		matches, exact = FuzzySearch(data, "TeSt", 0)
		require.False(t, exact)
		require.Len(t, matches, 3)
	})

	t.Run("special characters", func(t *testing.T) {
		data := []string{"hello-world", "hello_world", "hello world", "helloworld"}

		matches, _ := FuzzySearch(data, "hello-world", 1)
		slices.Sort(matches)
		require.Equal(t, []string{"hello world", "hello-world", "hello_world", "helloworld"}, matches)
	})

	t.Run("duplicate entries in data", func(t *testing.T) {
		data := []string{"test", "test", "test"}

		matches, exact := FuzzySearch(data, "test", 0)
		require.True(t, exact)
		require.Len(t, matches, 3)
	})

	t.Run("large maxDist matches everything short enough", func(t *testing.T) {
		data := []string{"a", "bb", "ccc", "dddd"}

		matches, _ := FuzzySearch(data, "x", 10)
		require.Len(t, matches, 4)
	})
}
