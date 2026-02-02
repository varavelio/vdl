package strutil

import (
	"slices"
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

		_, exact = FuzzySearch(data, "hElLo")
		require.False(t, exact)
	})

	t.Run("short query uses distance 1", func(t *testing.T) {
		// Query "cat" (3 chars) -> distance 1
		data := []string{"cat", "car", "bat", "cap", "dog"}

		matches, _ := FuzzySearch(data, "cat")
		slices.Sort(matches)
		// distance 1: cat(0), car(1), bat(1), cap(1), dog(3-excluded)
		require.Equal(t, []string{"bat", "cap", "car", "cat"}, matches)
	})

	t.Run("medium query uses distance 2", func(t *testing.T) {
		// Query "hello" (5 chars) -> distance 2
		data := []string{"hello", "hallo", "hullo", "help", "world", "helicopter"}

		matches, _ := FuzzySearch(data, "hello")
		slices.Sort(matches)
		// distance 2: hello(0), hallo(1), hullo(1), help(2), world(4-excluded), helicopter(5-excluded)
		require.Equal(t, []string{"hallo", "hello", "help", "hullo"}, matches)
	})

	t.Run("long query uses distance 3", func(t *testing.T) {
		// Query "programming" (11 chars) -> distance 3
		data := []string{"programming", "programing", "programmin", "programm", "developer"}

		matches, _ := FuzzySearch(data, "programming")
		slices.Sort(matches)
		// distance 3: programming(0), programing(1), programmin(1), programm(3), developer(excluded by length diff > 3)
		require.Equal(t, []string{"programing", "programm", "programmin", "programming"}, matches)
	})

	t.Run("boundary at 4 chars uses distance 1", func(t *testing.T) {
		// Query "test" (4 chars) -> distance 1
		data := []string{"test", "text", "best", "testing"}

		matches, _ := FuzzySearch(data, "test")
		slices.Sort(matches)
		// distance 1: test(0), text(1), best(1), testing(3-excluded by length diff)
		require.Equal(t, []string{"best", "test", "text"}, matches)
	})

	t.Run("boundary at 5 chars uses distance 2", func(t *testing.T) {
		// Query "tests" (5 chars) -> distance 2
		data := []string{"tests", "test", "testy", "texts", "bests", "tes"}

		matches, _ := FuzzySearch(data, "tests")
		slices.Sort(matches)
		// distance 2: tests(0), test(1), testy(1), texts(1), bests(1), tes(2)
		require.Equal(t, []string{"bests", "tes", "test", "tests", "testy", "texts"}, matches)
	})

	t.Run("boundary at 8 chars uses distance 2", func(t *testing.T) {
		// Query "function" (8 chars) -> distance 2
		data := []string{"function", "functio", "functions", "junction", "fraction"}

		matches, _ := FuzzySearch(data, "function")
		slices.Sort(matches)
		// distance 2: function(0), functio(1), functions(1), junction(1), fraction(2)
		require.Equal(t, []string{"fraction", "functio", "function", "functions", "junction"}, matches)
	})

	t.Run("boundary at 9 chars uses distance 3", func(t *testing.T) {
		// Query "functions" (9 chars) -> distance 3
		data := []string{"functions", "function", "funtions", "junctions", "fractions", "variable"}

		matches, _ := FuzzySearch(data, "functions")
		slices.Sort(matches)
		// distance 3: functions(0), function(1), funtions(2), junctions(1), fractions(2), variable(8-excluded)
		require.Equal(t, []string{"fractions", "function", "functions", "funtions", "junctions"}, matches)
	})

	t.Run("normalization handles accents and diacritics", func(t *testing.T) {
		data := []string{"cafe", "café", "CAFÉ", "Cafe"}

		matches, _ := FuzzySearch(data, "cafe")
		require.Len(t, matches, 4)

		matches, _ = FuzzySearch(data, "café")
		require.Len(t, matches, 4)

		_, exact := FuzzySearch(data, "café")
		require.True(t, exact)

		_, exact = FuzzySearch(data, "cafe")
		require.True(t, exact)
	})

	t.Run("normalization handles whitespace trimming", func(t *testing.T) {
		data := []string{"  test  ", "test", " TEST "}

		matches, _ := FuzzySearch(data, "test")
		require.Len(t, matches, 3)

		matches, _ = FuzzySearch(data, "  test  ")
		require.Len(t, matches, 3)
	})

	t.Run("empty query string", func(t *testing.T) {
		data := []string{"", "a", "ab", "abc"}

		matches, exact := FuzzySearch(data, "")
		slices.Sort(matches)
		// empty query (0 chars) -> distance 1
		require.Equal(t, []string{"", "a"}, matches)
		require.True(t, exact)
	})

	t.Run("empty strings in data", func(t *testing.T) {
		data := []string{"", "a", ""}

		// Query "a" (1 char) -> distance 1
		matches, _ := FuzzySearch(data, "a")
		slices.Sort(matches)
		require.Equal(t, []string{"", "", "a"}, matches)
	})

	t.Run("single character strings", func(t *testing.T) {
		data := []string{"a", "b", "c", "d", "e"}

		// Query "a" (1 char) -> distance 1
		matches, _ := FuzzySearch(data, "a")
		slices.Sort(matches)
		require.Equal(t, []string{"a", "b", "c", "d", "e"}, matches)

		matches, _ = FuzzySearch(data, "x")
		slices.Sort(matches)
		require.Equal(t, []string{"a", "b", "c", "d", "e"}, matches)
	})

	t.Run("unicode characters", func(t *testing.T) {
		data := []string{"日本語", "日本", "日本人", "中国語"}

		// "日本語" (3 runes) -> distance 1
		matches, exact := FuzzySearch(data, "日本語")
		slices.Sort(matches)
		require.Equal(t, []string{"日本", "日本人", "日本語"}, matches)
		require.True(t, exact)
	})

	t.Run("exact match independent of fuzzy matches", func(t *testing.T) {
		data := []string{"Test", "test", "TEST"}

		matches, exact := FuzzySearch(data, "test")
		require.True(t, exact)
		require.Len(t, matches, 3)

		matches, exact = FuzzySearch(data, "TeSt")
		require.False(t, exact)
		require.Len(t, matches, 3)
	})

	t.Run("special characters", func(t *testing.T) {
		data := []string{"hello-world", "hello_world", "hello world", "helloworld"}

		// "hello-world" (11 chars) -> distance 3
		matches, _ := FuzzySearch(data, "hello-world")
		slices.Sort(matches)
		require.Equal(t, []string{"hello world", "hello-world", "hello_world", "helloworld"}, matches)
	})

	t.Run("duplicate entries in data", func(t *testing.T) {
		data := []string{"test", "test", "test"}

		matches, exact := FuzzySearch(data, "test")
		require.True(t, exact)
		require.Len(t, matches, 3)
	})

	t.Run("no matches when all words too different", func(t *testing.T) {
		data := []string{"apple", "banana", "cherry"}

		// Query "xyz" (3 chars) -> distance 1, none match
		matches, _ := FuzzySearch(data, "xyz")
		require.Empty(t, matches)
	})
}
