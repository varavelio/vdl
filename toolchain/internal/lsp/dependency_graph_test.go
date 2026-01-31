package lsp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDependencyGraph_UpdateAndGet(t *testing.T) {
	t.Run("registers forward and reverse dependencies", func(t *testing.T) {
		g := NewDependencyGraph()

		// main.vdl imports models.vdl and utils.vdl
		g.UpdateDependencies("/main.vdl", []string{"/models.vdl", "/utils.vdl"})

		// Check that models.vdl has main.vdl as a dependent
		dependents := g.GetDependents("/models.vdl")
		require.Len(t, dependents, 1)
		require.Equal(t, "/main.vdl", dependents[0])

		// Check that utils.vdl has main.vdl as a dependent
		dependents = g.GetDependents("/utils.vdl")
		require.Len(t, dependents, 1)
		require.Equal(t, "/main.vdl", dependents[0])
	})

	t.Run("handles multiple dependents", func(t *testing.T) {
		g := NewDependencyGraph()

		// Both main.vdl and api.vdl import models.vdl
		g.UpdateDependencies("/main.vdl", []string{"/models.vdl"})
		g.UpdateDependencies("/api.vdl", []string{"/models.vdl"})

		dependents := g.GetDependents("/models.vdl")
		require.Len(t, dependents, 2)
		require.Contains(t, dependents, "/main.vdl")
		require.Contains(t, dependents, "/api.vdl")
	})

	t.Run("clears old dependencies on update", func(t *testing.T) {
		g := NewDependencyGraph()

		// Initially main.vdl imports models.vdl
		g.UpdateDependencies("/main.vdl", []string{"/models.vdl"})
		require.Len(t, g.GetDependents("/models.vdl"), 1)

		// Now main.vdl imports only utils.vdl (removed models.vdl import)
		g.UpdateDependencies("/main.vdl", []string{"/utils.vdl"})

		// models.vdl should no longer have main.vdl as dependent
		require.Empty(t, g.GetDependents("/models.vdl"))
		require.Len(t, g.GetDependents("/utils.vdl"), 1)
	})

	t.Run("returns nil for files with no dependents", func(t *testing.T) {
		g := NewDependencyGraph()

		dependents := g.GetDependents("/unknown.vdl")
		require.Nil(t, dependents)
	})
}

func TestDependencyGraph_RemoveFile(t *testing.T) {
	t.Run("removes file from graph", func(t *testing.T) {
		g := NewDependencyGraph()

		g.UpdateDependencies("/main.vdl", []string{"/models.vdl"})
		g.UpdateDependencies("/api.vdl", []string{"/models.vdl"})

		// Remove main.vdl
		g.RemoveFile("/main.vdl")

		// models.vdl should only have api.vdl as dependent now
		dependents := g.GetDependents("/models.vdl")
		require.Len(t, dependents, 1)
		require.Equal(t, "/api.vdl", dependents[0])
	})

	t.Run("cleans up empty dependency sets", func(t *testing.T) {
		g := NewDependencyGraph()

		g.UpdateDependencies("/main.vdl", []string{"/models.vdl"})

		// Remove main.vdl
		g.RemoveFile("/main.vdl")

		// models.vdl should have no dependents
		require.Nil(t, g.GetDependents("/models.vdl"))
	})
}

func TestDependencyGraph_Clear(t *testing.T) {
	t.Run("removes all entries", func(t *testing.T) {
		g := NewDependencyGraph()

		g.UpdateDependencies("/main.vdl", []string{"/models.vdl", "/utils.vdl"})
		g.UpdateDependencies("/api.vdl", []string{"/models.vdl"})

		g.Clear()

		require.Nil(t, g.GetDependents("/models.vdl"))
		require.Nil(t, g.GetDependents("/utils.vdl"))
	})
}

func TestDependencyGraph_TransitiveDependencies(t *testing.T) {
	t.Run("supports transitive dependency chains", func(t *testing.T) {
		g := NewDependencyGraph()

		// Chain: main.vdl -> api.vdl -> models.vdl
		g.UpdateDependencies("/main.vdl", []string{"/api.vdl"})
		g.UpdateDependencies("/api.vdl", []string{"/models.vdl"})

		// Direct dependents of models.vdl
		dependents := g.GetDependents("/models.vdl")
		require.Len(t, dependents, 1)
		require.Equal(t, "/api.vdl", dependents[0])

		// Direct dependents of api.vdl
		dependents = g.GetDependents("/api.vdl")
		require.Len(t, dependents, 1)
		require.Equal(t, "/main.vdl", dependents[0])
	})
}
