package python

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

var reservedWords = map[string]bool{
	"False": true, "None": true, "True": true, "and": true, "as": true,
	"assert": true, "async": true, "await": true, "break": true, "class": true,
	"continue": true, "def": true, "del": true, "elif": true, "else": true,
	"except": true, "finally": true, "for": true, "from": true, "global": true,
	"if": true, "import": true, "in": true, "is": true, "lambda": true,
	"nonlocal": true, "not": true, "or": true, "pass": true, "raise": true,
	"return": true, "try": true, "while": true, "with": true, "yield": true,
	"dict": true, "list": true, "str": true, "int": true, "float": true, "bool": true,
	"type": true, "format": true, "input": true, "id": true,
}

func sanitizeIdentifier(name string) string {
	if reservedWords[name] {
		return name + "_"
	}
	if name == "id" {
		return "id_"
	}
	return name
}

func toPythonType(parentName, fieldName string, t ir.TypeRef) string {
	switch t.Kind {
	case ir.TypeKindPrimitive:
		switch t.Primitive {
		case ir.PrimitiveString:
			return "str"
		case ir.PrimitiveInt:
			return "int"
		case ir.PrimitiveFloat:
			return "float"
		case ir.PrimitiveBool:
			return "bool"
		case ir.PrimitiveDatetime:
			return "datetime.datetime"
		}
	case ir.TypeKindArray:
		itemType := toPythonType(parentName, fieldName, *t.ArrayItem)
		result := itemType
		for i := 0; i < t.ArrayDimensions; i++ {
			result = fmt.Sprintf("List[%s]", result)
		}
		return result
	case ir.TypeKindMap:
		return fmt.Sprintf("Dict[str, %s]", toPythonType(parentName, fieldName, *t.MapValue))
	case ir.TypeKindType:
		return t.Type
	case ir.TypeKindEnum:
		return t.Enum
	case ir.TypeKindObject:
		inlineName := parentName + strutil.ToPascalCase(fieldName)
		return inlineName
	}
	return "Any"
}

func needsDictCheck(t ir.TypeRef) bool {
	switch t.Kind {
	case ir.TypeKindType, ir.TypeKindObject:
		return true
	}
	return false
}

func needsListCheck(t ir.TypeRef) bool {
	return t.Kind == ir.TypeKindArray
}

// renderDocstringPython renders a Python docstring.
func renderDocstringPython(g *gen.Generator, doc string) {
	if doc == "" {
		return
	}
	g.Line("    \"\"\"")
	lines := strings.Split(doc, "\n")
	for _, line := range lines {
		g.Linef("    %s", line)
	}
	g.Line("    \"\"\"")
}

// renderDeprecatedPython renders a deprecation notice for Python.
func renderDeprecatedPython(g *gen.Generator, deprecated *ir.Deprecation) {
	if deprecated == nil {
		return
	}
	g.Line("    .. deprecated::")
	if deprecated.Message != "" {
		g.Linef("        %s", deprecated.Message)
	}
}
