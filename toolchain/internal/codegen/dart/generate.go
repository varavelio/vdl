package dart

import (
	_ "embed"
	"strings"

	"github.com/uforg/ufogenkit"
	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

//go:embed pieces/pubspec.yaml
var pubspecRawPiece string

//go:embed pieces/pubspec.lock
var pubspecLockRawPiece string

//go:embed pieces/.gitignore
var gitignoreRawPiece string

// OutputFile represents a single generated file.
type OutputFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// Output represents multiple generated files for the Dart package.
type Output struct {
	Files []OutputFile `json:"files"`
}

// Generate takes a schema and a config and generates the Dart package files.
func Generate(sch schema.Schema, config Config) (Output, error) {
	subGenerators := []func(schema.Schema, Config) (string, error){
		generateCore,
		generateDomainTypes,
		generateProcedureTypes,
		generateStreamTypes,
		generateClient,
	}

	// 1) Generate lib/main.dart
	g := ufogenkit.NewGenKit().WithSpaces(2)
	for _, generator := range subGenerators {
		codeChunk, err := generator(sch, config)
		if err != nil {
			return Output{}, err
		}

		codeChunk = strings.TrimSpace(codeChunk)
		if codeChunk != "" {
			g.Raw(codeChunk)
			g.Break()
			g.Break()
		}
	}
	libClientContent := g.String()
	libClientContent = strutil.LimitConsecutiveNewlines(libClientContent, 2)

	dartClient := OutputFile{
		Path:    "lib/client.dart",
		Content: libClientContent,
	}

	// 2) Generate pubspec.yaml
	pubspecRawPiece = strings.ReplaceAll(pubspecRawPiece, "{{ package_name }}", config.PackageName)
	pubspec := OutputFile{
		Path:    "pubspec.yaml",
		Content: pubspecRawPiece,
	}

	// 3) Generate pubspec.lock
	pubspecLock := OutputFile{
		Path:    "pubspec.lock",
		Content: pubspecLockRawPiece,
	}

	// 4) Generate .gitignore
	gitignore := OutputFile{
		Path:    ".gitignore",
		Content: gitignoreRawPiece,
	}

	return Output{
		Files: []OutputFile{
			dartClient,
			pubspec,
			pubspecLock,
			gitignore,
		},
	}, nil
}
