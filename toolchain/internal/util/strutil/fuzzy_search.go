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

// FuzzySearch performs a search using the Levenshtein distance algorithm.
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

			if levenshtein(queryRunes, wordRunes) <= maxDist {
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

func levenshtein(a, b []rune) int {
	n, m := len(a), len(b)
	if n < m {
		a, b = b, a
		n, m = m, n
	}

	row := make([]int, m+1)
	for i := 0; i <= m; i++ {
		row[i] = i
	}

	for i := 1; i <= n; i++ {
		prev := i
		for j := 1; j <= m; j++ {
			var current int
			if a[i-1] == b[j-1] {
				current = row[j-1]
			} else {
				current = min(row[j-1]+1, min(row[j]+1, prev+1))
			}
			row[j-1] = prev
			prev = current
		}
		row[m] = prev
	}
	return row[m]
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
