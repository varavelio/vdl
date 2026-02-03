package python

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
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

func toPythonType(parentName, fieldName string, t irtypes.TypeRef) string {
	switch t.Kind {
	case irtypes.TypeKindPrimitive:
		switch t.GetPrimitiveName() {
		case irtypes.PrimitiveTypeString:
			return "str"
		case irtypes.PrimitiveTypeInt:
			return "int"
		case irtypes.PrimitiveTypeFloat:
			return "float"
		case irtypes.PrimitiveTypeBool:
			return "bool"
		case irtypes.PrimitiveTypeDatetime:
			return "datetime.datetime"
		}
	case irtypes.TypeKindArray:
		itemType := toPythonType(parentName, fieldName, *t.ArrayType)
		result := itemType
		for i := int64(0); i < t.GetArrayDims(); i++ {
			result = fmt.Sprintf("List[%s]", result)
		}
		return result
	case irtypes.TypeKindMap:
		return fmt.Sprintf("Dict[str, %s]", toPythonType(parentName, fieldName, *t.MapType))
	case irtypes.TypeKindType:
		return t.GetTypeName()
	case irtypes.TypeKindEnum:
		return t.GetEnumName()
	case irtypes.TypeKindObject:
		inlineName := parentName + strutil.ToPascalCase(fieldName)
		return inlineName
	}
	return "Any"
}

func needsDictCheck(t irtypes.TypeRef) bool {
	switch t.Kind {
	case irtypes.TypeKindType, irtypes.TypeKindObject:
		return true
	}
	return false
}

func needsListCheck(t irtypes.TypeRef) bool {
	return t.Kind == irtypes.TypeKindArray
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
func renderDeprecatedPython(g *gen.Generator, deprecated *string) {
	if deprecated == nil {
		return
	}
	g.Line("    .. deprecated::")
	if *deprecated != "" {
		g.Linef("        %s", *deprecated)
	}
}
