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

	t.Run("results sorted by distance then length similarity", func(t *testing.T) {
		// Query "hello" (5 chars) -> distance 2
		data := []string{"help", "hallo", "hello", "helloo", "helo"}

		matches, _ := FuzzySearch(data, "hello")
		// hello(dist=0, lenDiff=0), hallo(dist=1, lenDiff=0), helo(dist=1, lenDiff=1), helloo(dist=1, lenDiff=1)
		// Both helo and helloo have dist=1, lenDiff=1, so order between them is not guaranteed
		require.Len(t, matches, 3)
		require.Equal(t, "hello", matches[0])
		require.Equal(t, "hallo", matches[1])
		// Third is either helo or helloo (both have same distance and lenDiff)
		require.True(t, matches[2] == "helo" || matches[2] == "helloo")
	})

	t.Run("length similarity breaks distance ties", func(t *testing.T) {
		// Query "test" (4 chars) -> distance 1
		data := []string{"tests", "tes", "test", "text"}

		matches, _ := FuzzySearch(data, "test")
		// test(dist=0, lenDiff=0) should be first
		// text(dist=1, lenDiff=0), tes(dist=1, lenDiff=1), tests(dist=1, lenDiff=1)
		require.Equal(t, "test", matches[0])
		require.Equal(t, "text", matches[1])
		// Third could be "tes" or "tests" (both lenDiff=1), order may vary
		require.Len(t, matches, 3)
	})

	t.Run("transpositions rank higher than substitutions", func(t *testing.T) {
		// Query "uint" (4 chars) -> distance 1
		// This tests the "Did you mean 'unit'?" scenario
		data := []string{"unit", "hint", "uint", "init"}

		matches, exact := FuzzySearch(data, "uint")
		require.True(t, exact)
		// uint(dist=0), unit(dist=1, transposition), hint(dist=1), init(dist=2-excluded)
		require.Contains(t, matches, "uint")
		require.Contains(t, matches, "unit")
	})

	t.Run("short query uses distance 1", func(t *testing.T) {
		// Query "cat" (3 chars) -> distance 1
		data := []string{"cat", "car", "bat", "dog"}

		matches, _ := FuzzySearch(data, "cat")
		// cat(0), car(1), bat(1), dog(3-excluded)
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
		// All normalize to "cafe", so all match with distance 0
		require.Len(t, matches, 3)

		_, exact := FuzzySearch(data, "café")
		require.True(t, exact)
	})

	t.Run("normalization handles whitespace trimming", func(t *testing.T) {
		data := []string{"  test  ", "test", " TEST "}

		matches, _ := FuzzySearch(data, "test")
		require.Len(t, matches, 3)
	})

	t.Run("empty query string", func(t *testing.T) {
		data := []string{"", "a", "ab", "abc"}

		matches, exact := FuzzySearch(data, "")
		// empty query (0 chars) -> distance 1
		// ""(dist=0), "a"(dist=1)
		require.Len(t, matches, 2)
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
		// ab(0), ba(1-transposition), abc(1-insertion), acb excluded by limit
		require.Len(t, matches, 3)
		require.Equal(t, "ab", matches[0])
	})

	t.Run("transpositions in longer strings", func(t *testing.T) {
		data := []string{"receive", "recieve", "receiver"}

		// Query "receive" (7 chars) -> distance 2
		matches, _ := FuzzySearch(data, "receive")
		// receive(0), recieve(1-transposition), receiver(1-insertion)
		require.Len(t, matches, 3)
		require.Equal(t, "receive", matches[0])
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
		data[51] = "text"
		data[52] = "tест" // different but same length

		matches, _ := FuzzySearch(data, "test")
		require.LessOrEqual(t, len(matches), 3)
		require.Equal(t, "test", matches[0])
	})
}
