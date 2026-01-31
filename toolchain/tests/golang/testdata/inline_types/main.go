package main

import (
	"fmt"
	"os"

	"e2e/gen"
)

func main() {
	// 1. Array of inline objects
	files := []gen.ComplexInlineTypesFiles{
		{Path: "a", Content: "b"},
	}

	// 2. Map of inline objects
	meta := map[string]gen.ComplexInlineTypesMeta{
		"v1": {CreatedAt: "2023", Author: "me"},
	}

	// 3. Map of arrays of inline objects
	groupedFiles := map[string][]gen.ComplexInlineTypesGroupedFiles{
		"group1": {{Name: "f1", Size: 10}},
	}

	// 4. Array of maps of inline objects
	configs := []map[string]gen.ComplexInlineTypesConfigs{
		{"conf1": {Key: "k", Value: "v"}},
	}

	// 5. Nested arrays of inline objects
	grid := [][]gen.ComplexInlineTypesGrid{
		{{X: 1, Y: 2}},
	}

	// 6. Simple inline object
	simple := gen.ComplexInlineTypesSimple{
		Name:    "test",
		Enabled: true,
	}

	// 7. Deeply nested inline objects
	deepNest := gen.ComplexInlineTypesDeepNest{
		Level1: "l1",
		Child: gen.ComplexInlineTypesDeepNestChild{
			Level2: 2,
			GrandChild: gen.ComplexInlineTypesDeepNestChildGrandChild{
				Level3: true,
				GreatGrandChild: gen.ComplexInlineTypesDeepNestChildGrandChildGreatGrandChild{
					Level4: 4.5,
					Data:   "end",
				},
			},
		},
	}

	output := gen.ComplexInlineTypes{
		Files:        files,
		Meta:         meta,
		GroupedFiles: groupedFiles,
		Configs:      configs,
		Grid:         grid,
		Simple:       simple,
		DeepNest:     deepNest,
	}

	// Verify integrity
	if output.Files[0].Path != "a" {
		panic("files mismatch")
	}
	if output.Meta["v1"].Author != "me" {
		panic("meta mismatch")
	}
	if output.GroupedFiles["group1"][0].Size != 10 {
		panic("groupedFiles mismatch")
	}
	if output.Configs[0]["conf1"].Value != "v" {
		panic("configs mismatch")
	}
	if output.Grid[0][0].Y != 2 {
		panic("grid mismatch")
	}
	if output.Simple.Name != "test" {
		panic("simple mismatch")
	}
	if output.DeepNest.Child.GrandChild.GreatGrandChild.Data != "end" {
		panic("deepNest mismatch")
	}

	fmt.Println("Success")
	os.Exit(0)
}
