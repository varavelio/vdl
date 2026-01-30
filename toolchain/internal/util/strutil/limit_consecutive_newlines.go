package strutil

import "strings"

// LimitConsecutiveNewlines collapses sequences of consecutive \n characters in the
// provided string so that they never exceed the specified max value. For example,
// if max is 1, any run of two or more newlines ("\n") will be replaced with a
// single newline. When max is 2, three or more consecutive newlines will be
// replaced with exactly two newlines and so on.
//
// If max is less than 1 the function will treat it as 1.
func LimitConsecutiveNewlines(str string, max int) string {
	if max < 1 {
		max = 1
	}

	var builder strings.Builder
	builder.Grow(len(str))

	count := 0 // number of consecutive newlines currently encountered
	for i := range len(str) {
		ch := str[i]
		if ch == '\n' {
			if count < max {
				builder.WriteByte(ch)
			}
			count++
			continue
		}

		// reset counter when we hit a non-newline byte
		count = 0
		builder.WriteByte(ch)
	}

	return builder.String()
}
