package analysis_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Golden test suite for semantic analysis.
//
// How discovery works:
// - Every `.vdl` under `testdata/` is a potential test case.
// - If a directory contains `main.vdl`, that directory is treated as one multi-file
//   case (entrypoint is `main.vdl`, sibling `.vdl` + `.md` are loaded into VFS).
// - Any `.vdl` not inside a `main.vdl` directory runs as a single-file case.
//
// Directives (put them at the top of the entry `.vdl` file as `//` comments):
// - `@expect: EXXX`
//   Expect the analysis result to include diagnostic code `EXXX`.
//   You can declare this multiple times.
// - `@expect-no-errors`
//   Expect zero diagnostics.
//
// Assertions:
// - `@assert: has-type <TypeName>`
// - `@assert: has-enum <EnumName>`
// - `@assert: has-const <ConstName>`
// - `@assert: file-count <N>`
// - `@assert: standalone-docs <N>`
// - `@assert: type-spread <TypeName> <SpreadName>`
// - `@assert: field-resolved-type <TypeName> <fieldName> <ResolvedTypeName>`
// - `@assert: field-resolved-enum <TypeName> <fieldName> <ResolvedEnumName>`
// - `@assert: type-doc-contains <TypeName> <free text...>`
// - `@assert: field-doc-contains <TypeName> <fieldName> <free text...>`
// - `@assert: standalone-doc-contains <free text...>`
// - `@assert: const-kind <ConstName> <string|int|float|bool|object|array|reference|unknown>`
// - `@assert: diagnostic-message-contains <Code> <free text...>`
//
// To add new coverage, prefer adding new files under `testdata/` with directives
// instead of editing Go test code.

func TestGoldenCases(t *testing.T) {
	cases := discoverGoldenCases(t)
	require.NotEmpty(t, cases, "no golden cases found in testdata")

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			program, diagnostics := runGoldenCase(t, tc)
			assertGoldenResult(t, tc, program, diagnostics)
		})
	}
}
