package python

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func generateDomainTypes(schema *ir.Schema, cfg *config.PythonConfig) (string, error) {
	g := gen.New()

	for _, t := range schema.Types {
		g.Raw(GenerateDataclass(t.Name, t.Doc, t.Fields))
		g.Break()
	}

	return g.String(), nil
}

func GenerateDataclass(name, doc string, fields []ir.Field) string {
	g := gen.New()

	g.Linef("@dataclass")
	g.Linef("class %s:", name)
	if doc != "" {
		g.Linef("    \"\"\"%s\"\"\"", doc)
	}

	if len(fields) == 0 {
		g.Line("    pass")
		g.Break()

		// to_dict for empty class
		g.Line("    def to_dict(self) -> Dict[str, Any]:")
		g.Line("        return {}")
		g.Break()

		// from_dict for empty class
		g.Linef("    @staticmethod")
		g.Linef("    def from_dict(d: Dict[str, Any]) -> %s:", name)
		g.Linef("        return %s()", name)

		return g.String()
	}

	// Sort fields for dataclass definition: Required first, then Optional.
	// This avoids "non-default argument follows default argument" error.
	var sortedFields []ir.Field
	var optionalFields []ir.Field
	for _, f := range fields {
		if f.Optional {
			optionalFields = append(optionalFields, f)
		} else {
			sortedFields = append(sortedFields, f)
		}
	}
	sortedFields = append(sortedFields, optionalFields...)

	// Fields
	for _, f := range sortedFields {
		fieldName := strutil.ToSnakeCase(f.Name)
		fieldName = sanitizeIdentifier(fieldName)
		pyType := toPythonType(f.Type)

		if f.Optional {
			g.Linef("    %s: Optional[%s] = None", fieldName, pyType)
		} else {
			g.Linef("    %s: %s", fieldName, pyType)
		}
	}
	g.Break()

	// to_dict
	g.Line("    def to_dict(self) -> Dict[str, Any]:")
	g.Line("        result: Dict[str, Any] = {}")
	for _, f := range fields {
		fieldName := strutil.ToSnakeCase(f.Name)
		fieldName = sanitizeIdentifier(fieldName)
		keyName := f.Name // VDL (camelCase)

		valExpr := "self." + fieldName
		convExpr := genToDictExpr(f.Type, valExpr)

		if f.Optional {
			g.Linef("        if self.%s is not None:", fieldName)
			g.Linef("            result[%q] = %s", keyName, convExpr)
		} else {
			g.Linef("        result[%q] = %s", keyName, convExpr)
		}
	}
	g.Line("        return result")
	g.Break()

	// from_dict
	g.Line("    @staticmethod")
	g.Linef("    def from_dict(d: Dict[str, Any]) -> %s:", name)
	g.Linef("        return %s(", name)
	for _, f := range fields {
		fieldName := strutil.ToSnakeCase(f.Name)
		fieldName = sanitizeIdentifier(fieldName)
		keyName := f.Name

		// d.get("key")
		getExpr := fmt.Sprintf("d.get(%q)", keyName)
		convExpr := genFromDictExpr(f.Type, getExpr)

		g.Linef("            %s=%s,", fieldName, convExpr)
	}
	g.Line("        )")

	return g.String()
}

func genToDictExpr(t ir.TypeRef, val string) string {
	switch t.Kind {
	case ir.TypeKindPrimitive:
		if t.Primitive == ir.PrimitiveDatetime {
			return fmt.Sprintf("%s.isoformat()", val)
		}
		return val
	case ir.TypeKindEnum:
		return fmt.Sprintf("%s.value", val)
	case ir.TypeKindType:
		return fmt.Sprintf("%s.to_dict()", val)
	case ir.TypeKindArray:
		inner := genToDictExpr(*t.ArrayItem, "x")
		return fmt.Sprintf("[%s for x in %s]", inner, val)
	case ir.TypeKindMap:
		inner := genToDictExpr(*t.MapValue, "v")
		return fmt.Sprintf("{k: %s for k, v in %s.items()}", inner, val)
	}
	return val
}

func genFromDictExpr(t ir.TypeRef, val string) string {
	// val is the expression to access the value, e.g. d.get("key")
	// If it's a list/map/object, we need to handle if it's None inside the expression if we are iterating?
	// But d.get returns None if missing.

	switch t.Kind {
	case ir.TypeKindPrimitive:
		if t.Primitive == ir.PrimitiveDatetime {
			return fmt.Sprintf("datetime.datetime.fromisoformat(%s) if %s is not None else None", val, val)
		}
		// Basic types: strictly speaking we might want to cast, but usually we trust JSON
		return val
	case ir.TypeKindEnum:
		return fmt.Sprintf("%s(%s) if %s is not None else None", t.Enum, val, val)
	case ir.TypeKindType:
		return fmt.Sprintf("%s.from_dict(%s) if %s is not None else None", t.Type, val, val)
	case ir.TypeKindArray:
		inner := genFromDictExpr(*t.ArrayItem, "x")
		// if val is None, this crashes: [x for x in None]
		// so we need check
		return fmt.Sprintf("[%s for x in %s] if %s is not None else None", inner, val, val)
	case ir.TypeKindMap:
		inner := genFromDictExpr(*t.MapValue, "v")
		return fmt.Sprintf("{k: %s for k, v in %s.items()} if %s is not None else None", inner, val, val)
	}
	return val
}
