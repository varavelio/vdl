package strutil

import (
	"slices"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// FuzzySearch performs a search using the Damerau-Levenshtein distance algorithm.
//
// The maximum distance is automatically determined based on query length:
//   - query <= 4 chars: distance 1
//   - query <= 8 chars: distance 2
//   - query > 8 chars: distance 3
//
// It returns:
//   - fuzzyMatches: A slice of strings from data that fall within the distance threshold after normalization.
//   - exactMatchFound: A boolean indicating whether the exactMatch was identified before normalization (LITERALLY EXACT MATCH).
func FuzzySearch(data []string, query string) (fuzzyMatches []string, exactMatchFound bool) {
	if len(data) == 0 {
		return nil, false
	}

	// Exact literal match check (case-sensitive, no normalization)
	exactMatchFound = slices.Contains(data, query)

	normalizedQuery := normalize(query)
	queryRunes := []rune(normalizedQuery)
	maxDist := getAdaptiveDistance(normalizedQuery)

	resultsChan := make(chan string, len(data))
	var wg sync.WaitGroup

	for _, word := range data {
		wg.Go(func() {
			normalizedWord := normalize(word)
			wordRunes := []rune(normalizedWord)

			if abs(len(queryRunes)-len(wordRunes)) > maxDist {
				return
			}

			if damerauLevenshtein(queryRunes, wordRunes) <= maxDist {
				resultsChan <- word
			}
		})
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for m := range resultsChan {
		fuzzyMatches = append(fuzzyMatches, m)
	}

	return fuzzyMatches, exactMatchFound
}

func normalize(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

// damerauLevenshtein calculates the Damerau-Levenshtein distance between two rune slices.
// This includes insertions, deletions, substitutions, and transpositions of adjacent characters.
func damerauLevenshtein(a, b []rune) int {
	n, m := len(a), len(b)
	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}

	// Create matrix with dimensions (n+1) x (m+1)
	d := make([][]int, n+1)
	for i := range d {
		d[i] = make([]int, m+1)
	}

	// Initialize first column and row
	for i := 0; i <= n; i++ {
		d[i][0] = i
	}
	for j := 0; j <= m; j++ {
		d[0][j] = j
	}

	// Fill the matrix
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			d[i][j] = min(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)

			// Transposition
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+cost)
			}
		}
	}

	return d[n][m]
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func getAdaptiveDistance(query string) int {
	length := len([]rune(query))
	if length <= 4 {
		return 1
	}
	if length <= 8 {
		return 2
	}
	return 3
}
