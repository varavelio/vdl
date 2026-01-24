package ir

import (
	"sort"

	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
)

// FromProgram builds an IR Schema from a validated analysis.Program.
// It assumes the Program has passed through analysis.Analyze() without errors.
// All spreads are expanded, docs are normalized, and collections are sorted.
func FromProgram(program *analysis.Program) *Schema {
	schema := &Schema{
		Types:     make([]Type, 0, len(program.Types)),
		Enums:     make([]Enum, 0, len(program.Enums)),
		Constants: make([]Constant, 0, len(program.Consts)),
		Patterns:  make([]Pattern, 0, len(program.Patterns)),
		RPCs:      make([]RPC, 0, len(program.RPCs)),
		Docs:      make([]Doc, 0, len(program.StandaloneDocs)),
	}

	// Convert types
	for _, typ := range program.Types {
		schema.Types = append(schema.Types, convertType(typ, program.Types, program.Enums))
	}
	sortTypes(schema.Types)

	// Convert enums
	for _, enum := range program.Enums {
		schema.Enums = append(schema.Enums, convertEnum(enum))
	}
	sortEnums(schema.Enums)

	// Convert constants
	for _, cnst := range program.Consts {
		schema.Constants = append(schema.Constants, convertConstant(cnst))
	}
	sortConstants(schema.Constants)

	// Convert patterns
	for _, pattern := range program.Patterns {
		schema.Patterns = append(schema.Patterns, convertPattern(pattern))
	}
	sortPatterns(schema.Patterns)

	// Convert services (RPCs)
	for _, rpc := range program.RPCs {
		schema.RPCs = append(schema.RPCs, convertRPC(rpc, program.Types, program.Enums))
	}
	sortRPCs(schema.RPCs)

	// Populate flattened views
	for _, rpc := range schema.RPCs {
		schema.Procedures = append(schema.Procedures, rpc.Procs...)
		schema.Streams = append(schema.Streams, rpc.Streams...)

		// Add RPC-level docs
		for _, doc := range rpc.Docs {
			schema.Docs = append(schema.Docs, Doc{
				RPCName: rpc.Name,
				Content: doc,
			})
		}
	}

	// Convert standalone docs
	for _, doc := range program.StandaloneDocs {
		normalized := normalizeDoc(&doc.Content)
		if normalized != "" {
			schema.Docs = append(schema.Docs, Doc{
				Content: normalized,
			})
		}
	}

	return schema
}

// ============================================================================
// TYPE CONVERSION
// ============================================================================

func convertType(
	typ *analysis.TypeSymbol,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
) Type {
	return Type{
		Name:       typ.Name,
		Doc:        normalizeDoc(typ.Docstring),
		Deprecated: convertDeprecation(typ.Deprecated),
		Fields:     flattenFields(typ.Fields, typ.Spreads, types, enums),
	}
}

func convertField(
	field *analysis.FieldSymbol,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
) Field {
	return Field{
		Name:     field.Name,
		Doc:      normalizeDoc(field.Docstring),
		Optional: field.Optional,
		Type:     convertFieldType(field.Type, types, enums),
	}
}

func convertFieldType(
	info *analysis.FieldTypeInfo,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
) TypeRef {
	if info == nil {
		return TypeRef{Kind: TypeKindPrimitive, Primitive: PrimitiveString}
	}

	// Get the base type first
	baseRef := convertBaseFieldType(info, types, enums)

	// If there are array dimensions, wrap the base type with array info
	if info.ArrayDims > 0 {
		return TypeRef{
			Kind:            TypeKindArray,
			ArrayItem:       &baseRef,
			ArrayDimensions: info.ArrayDims,
		}
	}

	return baseRef
}

func convertBaseFieldType(
	info *analysis.FieldTypeInfo,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
) TypeRef {
	switch info.Kind {
	case analysis.FieldTypeKindPrimitive:
		return TypeRef{
			Kind:      TypeKindPrimitive,
			Primitive: convertPrimitiveType(info.Name),
		}

	case analysis.FieldTypeKindCustom:
		// Check if this is an enum or a type
		if enum, ok := enums[info.Name]; ok {
			return TypeRef{
				Kind: TypeKindEnum,
				Enum: info.Name,
				EnumInfo: &EnumInfo{
					ValueType: convertEnumValueType(enum.ValueType),
				},
			}
		}
		// It's a custom type
		return TypeRef{
			Kind: TypeKindType,
			Type: info.Name,
		}

	case analysis.FieldTypeKindMap:
		return TypeRef{
			Kind:     TypeKindMap,
			MapValue: ptrTypeRef(convertFieldType(info.MapValue, types, enums)),
		}

	case analysis.FieldTypeKindObject:
		return TypeRef{
			Kind:   TypeKindObject,
			Object: flattenInlineObject(info.ObjectDef, types, enums),
		}

	default:
		return TypeRef{Kind: TypeKindPrimitive, Primitive: PrimitiveString}
	}
}

func convertPrimitiveType(name string) PrimitiveType {
	switch name {
	case "string":
		return PrimitiveString
	case "int":
		return PrimitiveInt
	case "float":
		return PrimitiveFloat
	case "bool":
		return PrimitiveBool
	case "datetime":
		return PrimitiveDatetime
	default:
		return PrimitiveString
	}
}

func ptrTypeRef(ref TypeRef) *TypeRef {
	return &ref
}

// ============================================================================
// ENUM CONVERSION
// ============================================================================

func convertEnum(enum *analysis.EnumSymbol) Enum {
	members := make([]EnumMember, 0, len(enum.Members))
	for _, m := range enum.Members {
		members = append(members, EnumMember{
			Name:  m.Name,
			Value: m.Value,
		})
	}

	return Enum{
		Name:       enum.Name,
		Doc:        normalizeDoc(enum.Docstring),
		Deprecated: convertDeprecation(enum.Deprecated),
		ValueType:  convertEnumValueType(enum.ValueType),
		Members:    members,
	}
}

func convertEnumValueType(vt analysis.EnumValueType) EnumValueType {
	if vt == analysis.EnumValueTypeInt {
		return EnumValueTypeInt
	}
	return EnumValueTypeString
}

// ============================================================================
// CONSTANT CONVERSION
// ============================================================================

func convertConstant(cnst *analysis.ConstSymbol) Constant {
	return Constant{
		Name:       cnst.Name,
		Doc:        normalizeDoc(cnst.Docstring),
		Deprecated: convertDeprecation(cnst.Deprecated),
		ValueType:  convertConstValueType(cnst.ValueType),
		Value:      cnst.Value,
	}
}

func convertConstValueType(vt analysis.ConstValueType) ConstValueType {
	switch vt {
	case analysis.ConstValueTypeString:
		return ConstValueTypeString
	case analysis.ConstValueTypeInt:
		return ConstValueTypeInt
	case analysis.ConstValueTypeFloat:
		return ConstValueTypeFloat
	case analysis.ConstValueTypeBool:
		return ConstValueTypeBool
	default:
		return ConstValueTypeString
	}
}

// ============================================================================
// PATTERN CONVERSION
// ============================================================================

func convertPattern(pattern *analysis.PatternSymbol) Pattern {
	return Pattern{
		Name:         pattern.Name,
		Doc:          normalizeDoc(pattern.Docstring),
		Deprecated:   convertDeprecation(pattern.Deprecated),
		Template:     pattern.Template,
		Placeholders: pattern.Placeholders,
	}
}

// ============================================================================
// RPC CONVERSION
// ============================================================================

func convertRPC(
	rpc *analysis.RPCSymbol,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
) RPC {
	procs := make([]Procedure, 0, len(rpc.Procs))
	for _, proc := range rpc.Procs {
		procs = append(procs, convertProcedure(proc, types, enums, rpc.Name))
	}
	sortProcedures(procs)

	streams := make([]Stream, 0, len(rpc.Streams))
	for _, stream := range rpc.Streams {
		streams = append(streams, convertStream(stream, types, enums, rpc.Name))
	}
	sortStreams(streams)

	docs := make([]string, 0, len(rpc.StandaloneDocs))
	for _, doc := range rpc.StandaloneDocs {
		normalized := normalizeDoc(&doc.Content)
		if normalized != "" {
			docs = append(docs, normalized)
		}
	}

	return RPC{
		Name:       rpc.Name,
		Doc:        normalizeDoc(rpc.Docstring),
		Deprecated: convertDeprecation(rpc.Deprecated),
		Procs:      procs,
		Streams:    streams,
		Docs:       docs,
	}
}

func convertProcedure(
	proc *analysis.ProcSymbol,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
	rpcName string,
) Procedure {
	return Procedure{
		RPCName:    rpcName,
		Name:       proc.Name,
		Doc:        normalizeDoc(proc.Docstring),
		Deprecated: convertDeprecation(proc.Deprecated),
		Input:      flattenBlockFields(proc.Input, types, enums),
		Output:     flattenBlockFields(proc.Output, types, enums),
	}
}

func convertStream(
	stream *analysis.StreamSymbol,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
	rpcName string,
) Stream {
	return Stream{
		RPCName:    rpcName,
		Name:       stream.Name,
		Doc:        normalizeDoc(stream.Docstring),
		Deprecated: convertDeprecation(stream.Deprecated),
		Input:      flattenBlockFields(stream.Input, types, enums),
		Output:     flattenBlockFields(stream.Output, types, enums),
	}
}

// ============================================================================
// DEPRECATION CONVERSION
// ============================================================================

func convertDeprecation(dep *analysis.DeprecationInfo) *Deprecation {
	if dep == nil {
		return nil
	}
	return &Deprecation{
		Message: dep.Message,
	}
}

// ============================================================================
// SORTING FUNCTIONS
// ============================================================================

func sortTypes(types []Type) {
	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})
}

func sortEnums(enums []Enum) {
	sort.Slice(enums, func(i, j int) bool {
		return enums[i].Name < enums[j].Name
	})
}

func sortConstants(constants []Constant) {
	sort.Slice(constants, func(i, j int) bool {
		return constants[i].Name < constants[j].Name
	})
}

func sortPatterns(patterns []Pattern) {
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Name < patterns[j].Name
	})
}

func sortRPCs(rpcs []RPC) {
	sort.Slice(rpcs, func(i, j int) bool {
		return rpcs[i].Name < rpcs[j].Name
	})
}

func sortProcedures(procs []Procedure) {
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].Name < procs[j].Name
	})
}

func sortStreams(streams []Stream) {
	sort.Slice(streams, func(i, j int) bool {
		return streams[i].Name < streams[j].Name
	})
}
