package ir

import (
	"sort"
	"strconv"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// FromProgram builds a flat IR schema from a validated analysis program.
func FromProgram(program *analysis.Program) *irtypes.IrSchema {
	resolver := newValueResolver(program)

	schema := &irtypes.IrSchema{
		Types:     make([]irtypes.TypeDef, 0, len(program.Types)),
		Enums:     make([]irtypes.EnumDef, 0, len(program.Enums)),
		Constants: make([]irtypes.ConstantDef, 0, len(program.Consts)),
		Docs:      make([]irtypes.DocDef, 0, len(program.StandaloneDocs)),
	}

	for _, typ := range program.Types {
		schema.Types = append(schema.Types, convertType(typ, program.Types, program.Enums, resolver))
	}
	for _, enum := range program.Enums {
		schema.Enums = append(schema.Enums, convertEnum(enum, program.Enums, resolver))
	}
	for _, cnst := range program.Consts {
		schema.Constants = append(schema.Constants, convertConstant(cnst, program, resolver))
	}
	for _, doc := range program.StandaloneDocs {
		normalized := normalizeDoc(&doc.Content)
		if normalized == "" {
			continue
		}
		schema.Docs = append(schema.Docs, irtypes.DocDef{Content: normalized})
	}

	sort.Slice(schema.Types, func(i, j int) bool { return schema.Types[i].Name < schema.Types[j].Name })
	sort.Slice(schema.Enums, func(i, j int) bool { return schema.Enums[i].Name < schema.Enums[j].Name })
	sort.Slice(schema.Constants, func(i, j int) bool { return schema.Constants[i].Name < schema.Constants[j].Name })

	return schema
}

func convertType(
	typ *analysis.TypeSymbol,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
	resolver *valueResolver,
) irtypes.TypeDef {
	return irtypes.TypeDef{
		Name:        typ.Name,
		Doc:         normalizeDocPtr(typ.Docstring),
		Annotations: convertAnnotations(typ.Annotations, resolver),
		Fields:      flattenTypeFields(typ, types, enums, resolver),
	}
}

func convertEnum(
	enum *analysis.EnumSymbol,
	enums map[string]*analysis.EnumSymbol,
	resolver *valueResolver,
) irtypes.EnumDef {
	members := expandEnumMembers(enum, enums, map[string]bool{})
	irMembers := make([]irtypes.EnumDefMember, 0, len(members))

	for _, member := range members {
		irMembers = append(irMembers, irtypes.EnumDefMember{
			Name:        member.Name,
			Value:       member.Value,
			Doc:         normalizeDocPtr(member.Docstring),
			Annotations: convertAnnotations(member.Annotations, resolver),
		})
	}

	return irtypes.EnumDef{
		Name:        enum.Name,
		Doc:         normalizeDocPtr(enum.Docstring),
		Annotations: convertAnnotations(enum.Annotations, resolver),
		EnumType:    convertEnumType(enum.ValueType),
		Members:     irMembers,
	}
}

func convertConstant(
	cnst *analysis.ConstSymbol,
	program *analysis.Program,
	resolver *valueResolver,
) irtypes.ConstantDef {
	value, ok := resolver.resolveConstValue(cnst.Name)
	if !ok {
		value = irtypes.Value{
			Kind:        irtypes.ValueKindString,
			StringValue: irtypes.Ptr(""),
		}
	}

	return irtypes.ConstantDef{
		Name:        cnst.Name,
		Doc:         normalizeDocPtr(cnst.Docstring),
		Annotations: convertAnnotations(cnst.Annotations, resolver),
		TypeRef:     inferConstTypeRef(cnst, value, program.Types, program.Enums),
		Value:       value,
	}
}

func convertField(
	field *analysis.FieldSymbol,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
	resolver *valueResolver,
) irtypes.Field {
	return irtypes.Field{
		Name:        field.Name,
		Doc:         normalizeDocPtr(field.Docstring),
		Optional:    field.Optional,
		Annotations: convertAnnotations(field.Annotations, resolver),
		TypeRef:     convertFieldType(field.Type, types, enums, resolver),
	}
}

func convertFieldType(
	info *analysis.FieldTypeInfo,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
	resolver *valueResolver,
) irtypes.TypeRef {
	if info == nil {
		return primitiveTypeRef(irtypes.PrimitiveTypeString)
	}

	baseRef := convertBaseFieldType(info, types, enums, resolver)
	if info.ArrayDims <= 0 {
		return baseRef
	}

	dims := int64(info.ArrayDims)
	return irtypes.TypeRef{
		Kind:      irtypes.TypeKindArray,
		ArrayType: &baseRef,
		ArrayDims: &dims,
	}
}

func convertBaseFieldType(
	info *analysis.FieldTypeInfo,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
	resolver *valueResolver,
) irtypes.TypeRef {
	switch info.Kind {
	case analysis.FieldTypeKindPrimitive:
		return primitiveTypeRef(convertPrimitiveType(info.Name))

	case analysis.FieldTypeKindCustom:
		if enum, ok := enums[info.Name]; ok {
			enumType := convertEnumType(enum.ValueType)
			return irtypes.TypeRef{
				Kind:     irtypes.TypeKindEnum,
				EnumName: &info.Name,
				EnumType: &enumType,
			}
		}
		if _, ok := types[info.Name]; ok {
			return irtypes.TypeRef{Kind: irtypes.TypeKindType, TypeName: &info.Name}
		}
		return irtypes.TypeRef{Kind: irtypes.TypeKindType, TypeName: &info.Name}

	case analysis.FieldTypeKindMap:
		mapValue := convertFieldType(info.MapValue, types, enums, resolver)
		return irtypes.TypeRef{Kind: irtypes.TypeKindMap, MapType: &mapValue}

	case analysis.FieldTypeKindObject:
		return irtypes.TypeRef{
			Kind:         irtypes.TypeKindObject,
			ObjectFields: flattenInlineObjectFields(info.ObjectDef, types, enums, resolver),
		}

	default:
		return primitiveTypeRef(irtypes.PrimitiveTypeString)
	}
}

func convertPrimitiveType(name string) irtypes.PrimitiveType {
	switch name {
	case "string":
		return irtypes.PrimitiveTypeString
	case "int":
		return irtypes.PrimitiveTypeInt
	case "float":
		return irtypes.PrimitiveTypeFloat
	case "bool":
		return irtypes.PrimitiveTypeBool
	case "datetime":
		return irtypes.PrimitiveTypeDatetime
	default:
		return irtypes.PrimitiveTypeString
	}
}

func convertEnumType(vt analysis.EnumValueType) irtypes.EnumType {
	if vt == analysis.EnumValueTypeInt {
		return irtypes.EnumTypeInt
	}
	return irtypes.EnumTypeString
}

func primitiveTypeRef(primitive irtypes.PrimitiveType) irtypes.TypeRef {
	return irtypes.TypeRef{
		Kind:          irtypes.TypeKindPrimitive,
		PrimitiveName: &primitive,
	}
}

func convertAnnotations(annotations []*analysis.AnnotationRef, resolver *valueResolver) []irtypes.Annotation {
	if len(annotations) == 0 {
		return []irtypes.Annotation{}
	}

	result := make([]irtypes.Annotation, 0, len(annotations))
	for _, ann := range annotations {
		if ann == nil {
			continue
		}

		converted := irtypes.Annotation{Name: ann.Name}
		if ann.Argument != nil {
			if value, ok := resolver.resolveDataLiteral(ann.Argument); ok {
				converted.Argument = &value
			}
		}

		result = append(result, converted)
	}

	if len(result) == 0 {
		return []irtypes.Annotation{}
	}
	return result
}

func inferConstTypeRef(
	cnst *analysis.ConstSymbol,
	value irtypes.Value,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
) irtypes.TypeRef {
	if cnst.ExplicitTypeName != nil {
		typeName := *cnst.ExplicitTypeName
		if ast.IsPrimitiveType(typeName) {
			return primitiveTypeRef(convertPrimitiveType(typeName))
		}
		if enum, ok := enums[typeName]; ok {
			enumType := convertEnumType(enum.ValueType)
			return irtypes.TypeRef{
				Kind:     irtypes.TypeKindEnum,
				EnumName: &typeName,
				EnumType: &enumType,
			}
		}
		if _, ok := types[typeName]; ok {
			return irtypes.TypeRef{Kind: irtypes.TypeKindType, TypeName: &typeName}
		}
	}

	return inferTypeRefFromValue(value)
}

func inferTypeRefFromValue(value irtypes.Value) irtypes.TypeRef {
	switch value.Kind {
	case irtypes.ValueKindString:
		return primitiveTypeRef(irtypes.PrimitiveTypeString)
	case irtypes.ValueKindInt:
		return primitiveTypeRef(irtypes.PrimitiveTypeInt)
	case irtypes.ValueKindFloat:
		return primitiveTypeRef(irtypes.PrimitiveTypeFloat)
	case irtypes.ValueKindBool:
		return primitiveTypeRef(irtypes.PrimitiveTypeBool)

	case irtypes.ValueKindObject:
		entries := value.GetObjectEntries()
		fields := make([]irtypes.Field, 0, len(entries))
		for _, entry := range entries {
			fields = append(fields, irtypes.Field{
				Name:        entry.Key,
				Optional:    false,
				Annotations: []irtypes.Annotation{},
				TypeRef:     inferTypeRefFromValue(entry.Value),
			})
		}
		return irtypes.TypeRef{Kind: irtypes.TypeKindObject, ObjectFields: &fields}

	case irtypes.ValueKindArray:
		items := value.GetArrayItems()
		if len(items) == 0 {
			dims := int64(1)
			base := primitiveTypeRef(irtypes.PrimitiveTypeString)
			return irtypes.TypeRef{Kind: irtypes.TypeKindArray, ArrayType: &base, ArrayDims: &dims}
		}

		elemType := inferTypeRefFromValue(items[0])
		if elemType.Kind == irtypes.TypeKindArray && elemType.ArrayDims != nil && elemType.ArrayType != nil {
			dims := *elemType.ArrayDims + 1
			return irtypes.TypeRef{Kind: irtypes.TypeKindArray, ArrayType: elemType.ArrayType, ArrayDims: &dims}
		}

		dims := int64(1)
		return irtypes.TypeRef{Kind: irtypes.TypeKindArray, ArrayType: &elemType, ArrayDims: &dims}
	}

	return primitiveTypeRef(irtypes.PrimitiveTypeString)
}

type valueResolver struct {
	consts      map[string]*analysis.ConstSymbol
	enums       map[string]*analysis.EnumSymbol
	constValues map[string]irtypes.Value
	resolving   map[string]bool
}

func newValueResolver(program *analysis.Program) *valueResolver {
	return &valueResolver{
		consts:      program.Consts,
		enums:       program.Enums,
		constValues: make(map[string]irtypes.Value, len(program.Consts)),
		resolving:   make(map[string]bool, len(program.Consts)),
	}
}

func (r *valueResolver) resolveConstValue(name string) (irtypes.Value, bool) {
	if v, ok := r.constValues[name]; ok {
		return v, true
	}
	if r.resolving[name] {
		return irtypes.Value{}, false
	}

	cnst := r.consts[name]
	if cnst == nil || cnst.AST == nil || cnst.AST.Value == nil {
		return irtypes.Value{}, false
	}

	r.resolving[name] = true
	defer delete(r.resolving, name)

	v, ok := r.resolveDataLiteral(cnst.AST.Value)
	if ok {
		r.constValues[name] = v
	}
	return v, ok
}

func (r *valueResolver) resolveDataLiteral(lit *ast.DataLiteral) (irtypes.Value, bool) {
	if lit == nil {
		return irtypes.Value{}, false
	}

	if lit.Scalar != nil {
		return r.resolveScalarLiteral(lit.Scalar)
	}

	if lit.Object != nil {
		entries := make([]irtypes.ObjectEntry, 0, len(lit.Object.Entries))
		for _, entry := range lit.Object.Entries {
			if entry == nil {
				continue
			}

			if entry.Spread != nil {
				if entry.Spread.Ref.Member != nil {
					continue
				}
				spreadValue, ok := r.resolveConstValue(entry.Spread.Ref.Name)
				if !ok || spreadValue.Kind != irtypes.ValueKindObject {
					continue
				}
				entries = append(entries, spreadValue.GetObjectEntries()...)
				continue
			}

			value, ok := r.resolveDataLiteral(entry.Value)
			if !ok {
				continue
			}
			entries = append(entries, irtypes.ObjectEntry{Key: entry.Key, Value: value})
		}

		return irtypes.Value{
			Kind:          irtypes.ValueKindObject,
			ObjectEntries: &entries,
		}, true
	}

	if lit.Array != nil {
		items := make([]irtypes.Value, 0, len(lit.Array.Elements))
		for _, element := range lit.Array.Elements {
			value, ok := r.resolveDataLiteral(element)
			if !ok {
				continue
			}
			items = append(items, value)
		}

		return irtypes.Value{
			Kind:       irtypes.ValueKindArray,
			ArrayItems: &items,
		}, true
	}

	return irtypes.Value{}, false
}

func (r *valueResolver) resolveScalarLiteral(s *ast.ScalarLiteral) (irtypes.Value, bool) {
	if s.Str != nil {
		value := string(*s.Str)
		return irtypes.Value{Kind: irtypes.ValueKindString, StringValue: &value}, true
	}
	if s.Int != nil {
		n, err := strconv.ParseInt(*s.Int, 10, 64)
		if err != nil {
			return irtypes.Value{}, false
		}
		return irtypes.Value{Kind: irtypes.ValueKindInt, IntValue: &n}, true
	}
	if s.Float != nil {
		f, err := strconv.ParseFloat(*s.Float, 64)
		if err != nil {
			return irtypes.Value{}, false
		}
		return irtypes.Value{Kind: irtypes.ValueKindFloat, FloatValue: &f}, true
	}
	if s.True {
		b := true
		return irtypes.Value{Kind: irtypes.ValueKindBool, BoolValue: &b}, true
	}
	if s.False {
		b := false
		return irtypes.Value{Kind: irtypes.ValueKindBool, BoolValue: &b}, true
	}
	if s.Ref != nil {
		if s.Ref.Member == nil {
			return r.resolveConstValue(s.Ref.Name)
		}
		return lookupEnumMemberValue(r.enums, s.Ref.Name, *s.Ref.Member)
	}

	return irtypes.Value{}, false
}

func normalizeDoc(raw *string) string {
	if raw == nil {
		return ""
	}
	return strings.TrimSpace(strutil.NormalizeIndent(*raw))
}

func normalizeDocPtr(raw *string) *string {
	if raw == nil {
		return nil
	}
	normalized := strings.TrimSpace(strutil.NormalizeIndent(*raw))
	if normalized == "" {
		return nil
	}
	return &normalized
}
