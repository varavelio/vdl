package lsp

import "sync"

// DependencyGraph maintains a reverse dependency graph for tracking which files
// depend on which other files. This enables propagating changes when an imported
// file is modified.
//
// The graph maps imported files (children) to their importers (parents).
// Example: If main.vdl imports models.vdl, the graph stores:
//
//	models.vdl -> {main.vdl}
//
// This allows us to find all files that need re-analysis when models.vdl changes.
type DependencyGraph struct {
	mu sync.RWMutex
	// reverseDeps maps a file path to the set of files that import it.
	// Key: imported file (child), Value: set of importing files (parents)
	reverseDeps map[string]map[string]struct{}
	// forwardDeps maps a file path to the set of files it imports.
	// Key: importing file (parent), Value: set of imported files (children)
	// This is used to clean up stale dependencies when a file is re-analyzed.
	forwardDeps map[string]map[string]struct{}
}

// NewDependencyGraph creates a new empty dependency graph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		reverseDeps: make(map[string]map[string]struct{}),
		forwardDeps: make(map[string]map[string]struct{}),
	}
}

// UpdateDependencies updates the dependency graph for a given file.
// It clears any previous dependencies for the file and registers the new ones.
//
// Parameters:
//   - filePath: The absolute path of the file being analyzed (the importer/parent)
//   - imports: The absolute paths of files that this file imports (the children)
func (g *DependencyGraph) UpdateDependencies(filePath string, imports []string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// First, remove old dependencies for this file
	if oldImports, exists := g.forwardDeps[filePath]; exists {
		for oldImport := range oldImports {
			if parents, ok := g.reverseDeps[oldImport]; ok {
				delete(parents, filePath)
				// Clean up empty sets
				if len(parents) == 0 {
					delete(g.reverseDeps, oldImport)
				}
			}
		}
	}

	// Create new forward deps set
	newImports := make(map[string]struct{}, len(imports))
	for _, imp := range imports {
		newImports[imp] = struct{}{}
	}
	g.forwardDeps[filePath] = newImports

	// Register new reverse dependencies
	for _, imp := range imports {
		if g.reverseDeps[imp] == nil {
			g.reverseDeps[imp] = make(map[string]struct{})
		}
		g.reverseDeps[imp][filePath] = struct{}{}
	}
}

// GetDependents returns all files that depend on the given file (its parents/importers).
// The returned slice contains absolute file paths.
func (g *DependencyGraph) GetDependents(filePath string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	parents, exists := g.reverseDeps[filePath]
	if !exists {
		return nil
	}

	result := make([]string, 0, len(parents))
	for parent := range parents {
		result = append(result, parent)
	}
	return result
}

// GetAllDependents returns all files that depend on the given file, directly or indirectly.
// It returns a flattened list of all dependents.
func (g *DependencyGraph) GetAllDependents(filePath string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	visited := make(map[string]struct{})
	var result []string

	// Queue of files to inspect for their parents
	queue := []string{filePath}

	// We mark filePath as visited so we don't process it if we encounter a cycle back to it
	visited[filePath] = struct{}{}

	head := 0
	for head < len(queue) {
		current := queue[head]
		head++

		if parents, ok := g.reverseDeps[current]; ok {
			for p := range parents {
				if _, seen := visited[p]; !seen {
					visited[p] = struct{}{}
					queue = append(queue, p)
					result = append(result, p)
				}
			}
		}
	}

	return result
}

// RemoveFile removes a file and all its dependencies from the graph.
// This should be called when a file is closed.
func (g *DependencyGraph) RemoveFile(filePath string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Remove this file from reverse deps of its imports
	if imports, exists := g.forwardDeps[filePath]; exists {
		for imp := range imports {
			if parents, ok := g.reverseDeps[imp]; ok {
				delete(parents, filePath)
				if len(parents) == 0 {
					delete(g.reverseDeps, imp)
				}
			}
		}
		delete(g.forwardDeps, filePath)
	}

	// Also remove this file as a dependency target
	delete(g.reverseDeps, filePath)
}

// Clear removes all entries from the dependency graph.
func (g *DependencyGraph) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()

	clear(g.reverseDeps)
	clear(g.forwardDeps)
}
