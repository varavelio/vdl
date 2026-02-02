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

// matchKind represents the type of match found, ordered by relevance (lower = better).
type matchKind int

const (
	matchExact    matchKind = iota // Normalized strings are identical
	matchPrefix                    // Query is a prefix of the word
	matchSuffix                    // Query is a suffix of the word
	matchContains                  // Query is contained within the word
	matchEdit                      // Match by edit distance only
)

// fuzzyMatch holds a match candidate with its computed score.
type fuzzyMatch struct {
	word     string
	kind     matchKind
	distance int // edit distance (0 for exact/prefix/suffix/contains)
	lenDiff  int // absolute difference in length from query
}

// FuzzySearch performs an intelligent fuzzy search combining structural matching
// (prefix, suffix, contains) with Damerau-Levenshtein edit distance.
//
// Match priority (highest to lowest):
//  1. Exact match (after normalization)
//  2. Prefix match (query is prefix of word)
//  3. Suffix match (query is suffix of word)
//  4. Contains match (query appears within word)
//  5. Edit distance match (within adaptive threshold)
//
// The maximum edit distance is automatically determined based on query length:
//   - query <= 4 chars: distance 1
//   - query <= 8 chars: distance 2
//   - query > 8 chars: distance 3
//
// It returns:
//   - fuzzyMatches: Up to 3 strings from data, sorted by relevance.
//   - exactMatchFound: A boolean indicating whether an exact match was found before normalization.
func FuzzySearch(data []string, query string) (fuzzyMatches []string, exactMatchFound bool) {
	if len(data) == 0 {
		return nil, false
	}

	// Exact literal match check (case-sensitive, no normalization)
	exactMatchFound = slices.Contains(data, query)

	normalizedQuery := normalize(query)
	if normalizedQuery == "" {
		// Empty query: only return exact empty matches
		return collectEmptyMatches(data), exactMatchFound
	}

	queryRunes := []rune(normalizedQuery)
	queryLen := len(queryRunes)
	maxDist := getAdaptiveDistance(normalizedQuery)

	var matches []fuzzyMatch

	if len(data) < concurrencyThreshold {
		matches = searchSequential(data, normalizedQuery, queryRunes, queryLen, maxDist)
	} else {
		matches = searchParallel(data, normalizedQuery, queryRunes, queryLen, maxDist)
	}

	// Sort by: kind (lower=better), then distance, then length similarity
	slices.SortFunc(matches, compareMatches)

	// Return top N results
	limit := min(maxFuzzyResults, len(matches))
	fuzzyMatches = make([]string, limit)
	for i := range limit {
		fuzzyMatches[i] = matches[i].word
	}

	return fuzzyMatches, exactMatchFound
}

// compareMatches defines the sorting order for matches.
func compareMatches(a, b fuzzyMatch) int {
	// Primary: match kind (exact > prefix > suffix > contains > edit)
	if a.kind != b.kind {
		return int(a.kind) - int(b.kind)
	}
	// Secondary: edit distance (for edit matches)
	if a.distance != b.distance {
		return a.distance - b.distance
	}
	// Tertiary: length similarity
	return a.lenDiff - b.lenDiff
}

// evaluateMatch determines if and how a word matches the query.
// Returns the match and whether it qualifies.
// Uses short-circuit evaluation: structural matches skip edit distance calculation.
func evaluateMatch(word, normalizedQuery string, queryRunes []rune, queryLen, maxDist int) (fuzzyMatch, bool) {
	normalizedWord := normalize(word)
	wordLen := len([]rune(normalizedWord))
	lenDiff := abs(queryLen - wordLen)

	// Fast path: exact normalized match
	if normalizedWord == normalizedQuery {
		return fuzzyMatch{
			word:     word,
			kind:     matchExact,
			distance: 0,
			lenDiff:  0,
		}, true
	}

	// Structural matches: prefix, suffix, contains
	// These skip expensive edit distance calculation
	if strings.HasPrefix(normalizedWord, normalizedQuery) {
		return fuzzyMatch{
			word:     word,
			kind:     matchPrefix,
			distance: 0,
			lenDiff:  lenDiff,
		}, true
	}

	if strings.HasSuffix(normalizedWord, normalizedQuery) {
		return fuzzyMatch{
			word:     word,
			kind:     matchSuffix,
			distance: 0,
			lenDiff:  lenDiff,
		}, true
	}

	if strings.Contains(normalizedWord, normalizedQuery) {
		return fuzzyMatch{
			word:     word,
			kind:     matchContains,
			distance: 0,
			lenDiff:  lenDiff,
		}, true
	}

	// Edit distance match: only if length difference is within bounds
	if lenDiff > maxDist {
		return fuzzyMatch{}, false
	}

	wordRunes := []rune(normalizedWord)
	dist := damerauLevenshtein(queryRunes, wordRunes)
	if dist <= maxDist {
		return fuzzyMatch{
			word:     word,
			kind:     matchEdit,
			distance: dist,
			lenDiff:  lenDiff,
		}, true
	}

	return fuzzyMatch{}, false
}

func searchSequential(data []string, normalizedQuery string, queryRunes []rune, queryLen, maxDist int) []fuzzyMatch {
	matches := make([]fuzzyMatch, 0, min(len(data), maxFuzzyResults*2))

	for _, word := range data {
		if m, ok := evaluateMatch(word, normalizedQuery, queryRunes, queryLen, maxDist); ok {
			matches = append(matches, m)
		}
	}

	return matches
}

func searchParallel(data []string, normalizedQuery string, queryRunes []rune, queryLen, maxDist int) []fuzzyMatch {
	type result struct {
		match fuzzyMatch
		ok    bool
	}

	resultsChan := make(chan result, len(data))
	var wg sync.WaitGroup

	for _, word := range data {
		wg.Go(func() {
			m, ok := evaluateMatch(word, normalizedQuery, queryRunes, queryLen, maxDist)
			resultsChan <- result{match: m, ok: ok}
		})
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	matches := make([]fuzzyMatch, 0, min(len(data), maxFuzzyResults*2))
	for r := range resultsChan {
		if r.ok {
			matches = append(matches, r.match)
		}
	}

	return matches
}

// collectEmptyMatches handles the edge case of empty query.
func collectEmptyMatches(data []string) []string {
	var matches []string
	for _, word := range data {
		if normalize(word) == "" {
			matches = append(matches, word)
			if len(matches) >= maxFuzzyResults {
				break
			}
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
	row0 := make([]int, m+1)
	row1 := make([]int, m+1)
	row2 := make([]int, m+1)

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

			row2[j] = min(
				row1[j]+1,
				row2[j-1]+1,
				row1[j-1]+cost,
			)

			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				row2[j] = min(row2[j], row0[j-2]+cost)
			}
		}

		row0, row1, row2 = row1, row2, row0
	}

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
