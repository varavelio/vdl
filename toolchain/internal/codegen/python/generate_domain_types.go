package python

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func generateDomainTypes(schema *ir.Schema, _ *config.PythonConfig) (string, error) {
	g := gen.New()

	g.Line("def _require_field(name: str, value: Any) -> Any:")
	g.Line("    if value is None:")
	g.Line("        raise TypeError(f'Missing required field: {name}')")
	g.Line("    return value")
	g.Break()

	g.Line("def _require_list(value: Any) -> List[Any]:")
	g.Line("    if not isinstance(value, list):")
	g.Line("        raise TypeError('Expected list')")
	g.Line("    return value")
	g.Break()

	g.Line("def _require_dict(value: Any) -> Dict[str, Any]:")
	g.Line("    if not isinstance(value, dict):")
	g.Line("        raise TypeError('Expected dict')")
	g.Line("    return value")
	g.Break()

	g.Line("def _ensure_list(value: Any) -> List[Any]:")
	g.Line("    if value is None:")
	g.Line("        return []")
	g.Line("    if not isinstance(value, list):")
	g.Line("        raise TypeError('Expected list')")
	g.Line("    return value")
	g.Break()

	g.Line("def _ensure_dict(value: Any) -> Dict[str, Any]:")
	g.Line("    if value is None:")
	g.Line("        return {}")
	g.Line("    if not isinstance(value, dict):")
	g.Line("        raise TypeError('Expected dict')")
	g.Line("    return value")
	g.Break()

	for _, t := range schema.Types {
		g.Raw(GenerateDataclass(t.Name, t.Doc, t.Fields))
		g.Break()
		g.Raw(renderInlineTypes(t.Name, t.Fields))
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

	g.Line("    _json_key_map = {")
	for _, f := range fields {
		fieldName := strutil.ToSnakeCase(f.Name)
		fieldName = sanitizeIdentifier(fieldName)
		g.Linef("        %q: %q,", fieldName, f.Name)
	}
	g.Line("    }")
	g.Break()

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
		pyType := toPythonType(name, f.Name, f.Type)

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
		convExpr := genToDictExpr(name, f.Name, f.Type, valExpr)

		if f.Optional {
			g.Linef("        if self.%s is not None:", fieldName)
			g.Linef("            result[%q] = %s", keyName, convExpr)
		} else {
			if needsDictCheck(f.Type) {
				g.Linef("        if not hasattr(self.%s, 'to_dict'):", fieldName)
				g.Line("            raise TypeError('Expected object with to_dict')")
			}
			if needsListCheck(f.Type) {
				g.Linef("        if not isinstance(self.%s, list):", fieldName)
				g.Line("            raise TypeError('Expected list')")
			}
			g.Linef("        result[%q] = %s", keyName, convExpr)
		}
	}
	g.Line("        return result")
	g.Break()

	// from_dict
	g.Line("    @staticmethod")
	g.Linef("    def from_dict(d: Dict[str, Any]) -> %s:", name)
	g.Line("        if not isinstance(d, dict):")
	g.Linef("            raise TypeError(%q)", "Expected dict")
	g.Linef("        return %s(", name)
	for _, f := range fields {
		fieldName := strutil.ToSnakeCase(f.Name)
		fieldName = sanitizeIdentifier(fieldName)
		keyName := f.Name

		// d.get("key")
		getExpr := fmt.Sprintf("d.get(%q)", keyName)
		convExpr := genFromDictExpr(name, f.Name, f.Type, getExpr)
		if f.Optional {
			g.Linef("            %s=%s,", fieldName, convExpr)
		} else {
			g.Linef("            %s=_require_field(%q, %s),", fieldName, keyName, convExpr)
		}
	}
	g.Line("        )")

	return g.String()
}

func genToDictExpr(parentName, fieldName string, t ir.TypeRef, val string) string {
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
		if t.ArrayDimensions > 1 {
			inner := genNestedArrayToDictExpr(parentName, fieldName, *t.ArrayItem, t.ArrayDimensions-1, "inner")
			return fmt.Sprintf("[%s for inner in _ensure_list(%s)]", inner, val)
		}
		inner := genToDictExpr(parentName, fieldName, *t.ArrayItem, "x")
		return fmt.Sprintf("[%s for x in _ensure_list(%s)]", inner, val)
	case ir.TypeKindMap:
		inner := genToDictExpr(parentName, fieldName, *t.MapValue, "v")
		return fmt.Sprintf("{k: %s for k, v in _ensure_dict(%s).items()}", inner, val)
	case ir.TypeKindObject:
		return fmt.Sprintf("%s.to_dict()", val)
	}
	return val
}

func genNestedArrayToDictExpr(parentName, fieldName string, item ir.TypeRef, remaining int, varName string) string {
	if remaining == 0 {
		return genToDictExpr(parentName, fieldName, item, varName)
	}
	inner := genNestedArrayToDictExpr(parentName, fieldName, item, remaining-1, "x")
	return fmt.Sprintf("[%s for x in _ensure_list(%s)]", inner, varName)
}

func genFromDictExpr(parentName, fieldName string, t ir.TypeRef, val string) string {
	// val is the expression to access the value, e.g. d.get("key")
	// If it's a list/map/object, we need to handle if it's None inside the expression if we are iterating?
	// But d.get returns None if missing.

	switch t.Kind {
	case ir.TypeKindPrimitive:
		if t.Primitive == ir.PrimitiveDatetime {
			return fmt.Sprintf("datetime.datetime.fromisoformat(%s) if %s is not None else None", val, val)
		}
		// Basic types: strictly speaking we might want to cast, but usually we trust JSON
		if t.Primitive == ir.PrimitiveInt {
			return fmt.Sprintf("int(%s) if %s is not None else None", val, val)
		}
		if t.Primitive == ir.PrimitiveFloat {
			return fmt.Sprintf("float(%s) if %s is not None else None", val, val)
		}
		return val
	case ir.TypeKindEnum:
		return fmt.Sprintf("%s(%s) if %s is not None else None", t.Enum, val, val)
	case ir.TypeKindType:
		return fmt.Sprintf("%s.from_dict(_ensure_dict(%s)) if %s is not None else None", t.Type, val, val)
	case ir.TypeKindArray:
		if t.ArrayDimensions > 1 {
			inner := genNestedArrayFromDictExpr(parentName, fieldName, *t.ArrayItem, t.ArrayDimensions-1, "inner")
			return fmt.Sprintf("[%s for inner in _ensure_list(%s)] if %s is not None else None", inner, val, val)
		}
		inner := genFromDictExpr(parentName, fieldName, *t.ArrayItem, "x")
		return fmt.Sprintf("[%s for x in _ensure_list(%s)] if %s is not None else None", inner, val, val)
	case ir.TypeKindMap:
		inner := genFromDictExpr(parentName, fieldName, *t.MapValue, "v")
		return fmt.Sprintf("{k: %s for k, v in _ensure_dict(%s).items()} if %s is not None else None", inner, val, val)
	case ir.TypeKindObject:
		inlineName := parentName + strutil.ToPascalCase(fieldName)
		return fmt.Sprintf("%s.from_dict(_ensure_dict(%s)) if %s is not None else None", inlineName, val, val)
	}
	return val
}

func genNestedArrayFromDictExpr(parentName, fieldName string, item ir.TypeRef, remaining int, varName string) string {
	if remaining == 0 {
		return genFromDictExpr(parentName, fieldName, item, varName)
	}
	inner := genNestedArrayFromDictExpr(parentName, fieldName, item, remaining-1, "x")
	return fmt.Sprintf("[%s for x in _ensure_list(%s)]", inner, varName)
}
