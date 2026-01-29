package python

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/ir"
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
	return name
}

func toPythonType(t ir.TypeRef) string {
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
		return fmt.Sprintf("List[%s]", toPythonType(*t.ArrayItem))
	case ir.TypeKindMap:
		return fmt.Sprintf("Dict[str, %s]", toPythonType(*t.MapValue))
	case ir.TypeKindType:
		return t.Type
	case ir.TypeKindEnum:
		return t.Enum
	case ir.TypeKindObject:
		return "Dict[str, Any]" // Inline objects are not fully supported as named types yet, usually mapped to Dict or Any
	}
	return "Any"
}
