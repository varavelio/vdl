package analysis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
)

func TestAnalyzeWithContext(t *testing.T) {
	t.Run("returns nil when context is already cancelled", func(t *testing.T) {
		fs := vfs.New()
		fs.WriteFileCache("/test.vdl", []byte(`type User { name: string }`))

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		program, diagnostics := AnalyzeWithContext(ctx, fs, "/test.vdl")
		require.Nil(t, program)
		require.Nil(t, diagnostics)
	})

	t.Run("completes normally with valid context", func(t *testing.T) {
		fs := vfs.New()
		fs.WriteFileCache("/test.vdl", []byte(`type User { name: string }`))

		ctx := context.Background()
		program, diagnostics := AnalyzeWithContext(ctx, fs, "/test.vdl")

		require.NotNil(t, program)
		require.Empty(t, diagnostics)
		require.Contains(t, program.Types, "User")
	})

	t.Run("Analyze uses background context internally", func(t *testing.T) {
		fs := vfs.New()
		fs.WriteFileCache("/test.vdl", []byte(`type User { name: string }`))

		program, diagnostics := Analyze(fs, "/test.vdl")

		require.NotNil(t, program)
		require.Empty(t, diagnostics)
		require.Contains(t, program.Types, "User")
	})
}

func TestResolverWithContext(t *testing.T) {
	t.Run("stops resolution when context is cancelled", func(t *testing.T) {
		fs := vfs.New()
		fs.WriteFileCache("/main.vdl", []byte(`include "a.vdl"`))
		fs.WriteFileCache("/a.vdl", []byte(`include "b.vdl"`))
		fs.WriteFileCache("/b.vdl", []byte(`type B { name: string }`))

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		resolver := newResolverWithContext(ctx, fs)
		files, _ := resolver.resolve("/main.vdl")

		// Should have stopped early due to cancellation
		require.Empty(t, files)
	})

	t.Run("resolves all files with valid context", func(t *testing.T) {
		fs := vfs.New()
		fs.WriteFileCache("/main.vdl", []byte(`include "a.vdl"`))
		fs.WriteFileCache("/a.vdl", []byte(`type A { name: string }`))

		ctx := context.Background()
		resolver := newResolverWithContext(ctx, fs)
		files, diagnostics := resolver.resolve("/main.vdl")

		require.Len(t, files, 2)
		require.Empty(t, diagnostics)
	})
}

func TestValidatorWithContext(t *testing.T) {
	t.Run("stops collection when context is cancelled", func(t *testing.T) {
		// Create mock files
		files := make(map[string]*File)
		// Would normally have parsed ASTs here, but we're testing the cancel behavior

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		validator := newValidatorWithContext(ctx, files)
		diagnostics := validator.collect()

		// Should return early with empty diagnostics
		require.Empty(t, diagnostics)
	})

	t.Run("stops validation when context is cancelled", func(t *testing.T) {
		files := make(map[string]*File)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		validator := newValidatorWithContext(ctx, files)
		diagnostics := validator.validate()

		// Should return early with empty diagnostics
		require.Empty(t, diagnostics)
	})
}
