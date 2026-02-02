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

const (
	// maxFuzzyResults is the maximum number of fuzzy matches returned.
	maxFuzzyResults = 3
	// concurrencyThreshold defines the minimum data size to use parallel processing.
	// Below this threshold, sequential processing is more efficient.
	concurrencyThreshold = 50
)

// fuzzyMatch holds a match candidate with its computed distance.
type fuzzyMatch struct {
	word     string
	distance int
	lenDiff  int // absolute difference in length from query
}

// FuzzySearch performs a search using the Damerau-Levenshtein distance algorithm.
//
// The maximum distance is automatically determined based on query length:
//   - query <= 4 chars: distance 1
//   - query <= 8 chars: distance 2
//   - query > 8 chars: distance 3
//
// It returns:
//   - fuzzyMatches: Up to 3 strings from data, sorted by the distance (then by length similarity).
//   - exactMatchFound: A boolean indicating whether an exact match was found before normalization.
func FuzzySearch(data []string, query string) (fuzzyMatches []string, exactMatchFound bool) {
	if len(data) == 0 {
		return nil, false
	}

	// Exact literal match check (case-sensitive, no normalization)
	exactMatchFound = slices.Contains(data, query)

	normalizedQuery := normalize(query)
	queryRunes := []rune(normalizedQuery)
	queryLen := len(queryRunes)
	maxDist := getAdaptiveDistance(normalizedQuery)

	var matches []fuzzyMatch

	if len(data) < concurrencyThreshold {
		// Sequential processing for small datasets
		matches = searchSequential(data, queryRunes, queryLen, maxDist)
	} else {
		// Parallel processing for larger datasets
		matches = searchParallel(data, queryRunes, queryLen, maxDist)
	}

	// Sort by distance first, then by length difference (tiebreaker)
	slices.SortFunc(matches, func(a, b fuzzyMatch) int {
		if a.distance != b.distance {
			return a.distance - b.distance
		}
		return a.lenDiff - b.lenDiff
	})

	// Return top N results
	limit := min(maxFuzzyResults, len(matches))
	fuzzyMatches = make([]string, limit)
	for i := range limit {
		fuzzyMatches[i] = matches[i].word
	}

	return fuzzyMatches, exactMatchFound
}

func searchSequential(data []string, queryRunes []rune, queryLen, maxDist int) []fuzzyMatch {
	var matches []fuzzyMatch

	for _, word := range data {
		normalizedWord := normalize(word)
		wordRunes := []rune(normalizedWord)
		wordLen := len(wordRunes)

		lenDiff := abs(queryLen - wordLen)
		if lenDiff > maxDist {
			continue
		}

		dist := damerauLevenshtein(queryRunes, wordRunes)
		if dist <= maxDist {
			matches = append(matches, fuzzyMatch{
				word:     word,
				distance: dist,
				lenDiff:  lenDiff,
			})
		}
	}

	return matches
}

func searchParallel(data []string, queryRunes []rune, queryLen, maxDist int) []fuzzyMatch {
	type result struct {
		match fuzzyMatch
		ok    bool
	}

	resultsChan := make(chan result, len(data))
	var wg sync.WaitGroup

	for _, word := range data {
		wg.Go(func() {
			normalizedWord := normalize(word)
			wordRunes := []rune(normalizedWord)
			wordLen := len(wordRunes)

			lenDiff := abs(queryLen - wordLen)
			if lenDiff > maxDist {
				resultsChan <- result{ok: false}
				return
			}

			dist := damerauLevenshtein(queryRunes, wordRunes)
			if dist <= maxDist {
				resultsChan <- result{
					match: fuzzyMatch{
						word:     word,
						distance: dist,
						lenDiff:  lenDiff,
					},
					ok: true,
				}
			} else {
				resultsChan <- result{ok: false}
			}
		})
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var matches []fuzzyMatch
	for r := range resultsChan {
		if r.ok {
			matches = append(matches, r.match)
		}
	}

	return matches
}

func normalize(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

// damerauLevenshtein calculates the Damerau-Levenshtein distance using O(m) space.
// This includes insertions, deletions, substitutions, and transpositions of adjacent characters.
func damerauLevenshtein(a, b []rune) int {
	n, m := len(a), len(b)
	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}

	// Ensure a is the longer string for consistent row usage
	if n < m {
		a, b = b, a
		n, m = m, n
	}

	// Use 3 rows instead of full matrix: previous-previous, previous, current
	// This reduces space from O(n*m) to O(m)
	row0 := make([]int, m+1) // two rows back (for transposition)
	row1 := make([]int, m+1) // previous row
	row2 := make([]int, m+1) // current row

	// Initialize first row
	for j := range m + 1 {
		row1[j] = j
	}

	for i := 1; i <= n; i++ {
		row2[0] = i

		for j := 1; j <= m; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			// deletion, insertion, substitution
			row2[j] = min(
				row1[j]+1,      // deletion
				row2[j-1]+1,    // insertion
				row1[j-1]+cost, // substitution
			)

			// transposition of adjacent characters
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				row2[j] = min(row2[j], row0[j-2]+cost)
			}
		}

		// Rotate rows: row0 <- row1 <- row2
		row0, row1, row2 = row1, row2, row0
	}

	// Result is in row1 after the last rotation
	return row1[m]
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
