package ir

import (
	"sort"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// FromProgram builds an IR Schema from a validated analysis.Program.
// It assumes the Program has passed through analysis.Analyze() without errors.
// All spreads are expanded, docs are normalized, and collections are sorted.
func FromProgram(program *analysis.Program) *irtypes.IrSchema {
	schema := &irtypes.IrSchema{
		Types:      make([]irtypes.TypeDef, 0, len(program.Types)),
		Enums:      make([]irtypes.EnumDef, 0, len(program.Enums)),
		Constants:  make([]irtypes.ConstantDef, 0, len(program.Consts)),
		Patterns:   make([]irtypes.PatternDef, 0, len(program.Patterns)),
		Rpcs:       make([]irtypes.RpcDef, 0, len(program.RPCs)),
		Procedures: make([]irtypes.ProcedureDef, 0),
		Streams:    make([]irtypes.StreamDef, 0),
		Docs:       make([]irtypes.DocDef, 0, len(program.StandaloneDocs)),
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

	// Convert services (RPCs) - now we flatten procs/streams separately
	for _, rpc := range program.RPCs {
		// Add RPC definition
		schema.Rpcs = append(schema.Rpcs, convertRPCDef(rpc))

		// Add procedures
		for _, proc := range rpc.Procs {
			schema.Procedures = append(schema.Procedures, convertProcedure(proc, program.Types, program.Enums, rpc.Name))
		}

		// Add streams
		for _, stream := range rpc.Streams {
			schema.Streams = append(schema.Streams, convertStream(stream, program.Types, program.Enums, rpc.Name))
		}

		// Add RPC-level docs
		for _, doc := range rpc.StandaloneDocs {
			normalized := normalizeDoc(&doc.Content)
			if normalized != "" {
				rpcName := rpc.Name
				schema.Docs = append(schema.Docs, irtypes.DocDef{
					RpcName: &rpcName,
					Content: normalized,
				})
			}
		}
	}
	sortRPCs(schema.Rpcs)
	sortProcedures(schema.Procedures)
	sortStreams(schema.Streams)

	// Convert standalone docs
	for _, doc := range program.StandaloneDocs {
		normalized := normalizeDoc(&doc.Content)
		if normalized != "" {
			schema.Docs = append(schema.Docs, irtypes.DocDef{
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
) irtypes.TypeDef {
	return irtypes.TypeDef{
		Name:        typ.Name,
		Doc:         normalizeDocPtr(typ.Docstring),
		Deprecation: convertDeprecation(typ.Deprecated),
		Fields:      flattenFields(typ.Fields, typ.Spreads, types, enums),
	}
}

func convertField(
	field *analysis.FieldSymbol,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
) irtypes.Field {
	return irtypes.Field{
		Name:     field.Name,
		Doc:      normalizeDocPtr(field.Docstring),
		Optional: field.Optional,
		TypeRef:  convertFieldType(field.Type, types, enums),
	}
}

func convertFieldType(
	info *analysis.FieldTypeInfo,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
) irtypes.TypeRef {
	if info == nil {
		return irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeString)}
	}

	// Get the base type first
	baseRef := convertBaseFieldType(info, types, enums)

	// If there are array dimensions, wrap the base type with array info
	if info.ArrayDims > 0 {
		dims := int64(info.ArrayDims)
		return irtypes.TypeRef{
			Kind:      irtypes.TypeKindArray,
			ArrayType: &baseRef,
			ArrayDims: &dims,
		}
	}

	return baseRef
}

func convertBaseFieldType(
	info *analysis.FieldTypeInfo,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
) irtypes.TypeRef {
	switch info.Kind {
	case analysis.FieldTypeKindPrimitive:
		primType := convertPrimitiveType(info.Name)
		return irtypes.TypeRef{
			Kind:          irtypes.TypeKindPrimitive,
			PrimitiveName: &primType,
		}

	case analysis.FieldTypeKindCustom:
		// Check if this is an enum or a type
		if enum, ok := enums[info.Name]; ok {
			enumType := convertEnumType(enum.ValueType)
			return irtypes.TypeRef{
				Kind:     irtypes.TypeKindEnum,
				EnumName: &info.Name,
				EnumType: &enumType,
			}
		}
		// It's a custom type
		return irtypes.TypeRef{
			Kind:     irtypes.TypeKindType,
			TypeName: &info.Name,
		}

	case analysis.FieldTypeKindMap:
		mapType := convertFieldType(info.MapValue, types, enums)
		return irtypes.TypeRef{
			Kind:    irtypes.TypeKindMap,
			MapType: &mapType,
		}

	case analysis.FieldTypeKindObject:
		return irtypes.TypeRef{
			Kind:         irtypes.TypeKindObject,
			ObjectFields: flattenInlineObjectFields(info.ObjectDef, types, enums),
		}

	default:
		return irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeString)}
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

// ============================================================================
// ENUM CONVERSION
// ============================================================================

func convertEnum(enum *analysis.EnumSymbol) irtypes.EnumDef {
	members := make([]irtypes.EnumDefMember, 0, len(enum.Members))
	for _, m := range enum.Members {
		members = append(members, irtypes.EnumDefMember{
			Name:  m.Name,
			Value: m.Value,
		})
	}

	return irtypes.EnumDef{
		Name:        enum.Name,
		Doc:         normalizeDocPtr(enum.Docstring),
		Deprecation: convertDeprecation(enum.Deprecated),
		EnumType:    convertEnumType(enum.ValueType),
		Members:     members,
	}
}

func convertEnumType(vt analysis.EnumValueType) irtypes.EnumType {
	if vt == analysis.EnumValueTypeInt {
		return irtypes.EnumTypeInt
	}
	return irtypes.EnumTypeString
}

// ============================================================================
// CONSTANT CONVERSION
// ============================================================================

func convertConstant(cnst *analysis.ConstSymbol) irtypes.ConstantDef {
	return irtypes.ConstantDef{
		Name:        cnst.Name,
		Doc:         normalizeDocPtr(cnst.Docstring),
		Deprecation: convertDeprecation(cnst.Deprecated),
		ConstType:   convertConstType(cnst.ValueType),
		Value:       cnst.Value,
	}
}

func convertConstType(vt analysis.ConstValueType) irtypes.ConstType {
	switch vt {
	case analysis.ConstValueTypeString:
		return irtypes.ConstTypeString
	case analysis.ConstValueTypeInt:
		return irtypes.ConstTypeInt
	case analysis.ConstValueTypeFloat:
		return irtypes.ConstTypeFloat
	case analysis.ConstValueTypeBool:
		return irtypes.ConstTypeBool
	default:
		return irtypes.ConstTypeString
	}
}

// ============================================================================
// PATTERN CONVERSION
// ============================================================================

func convertPattern(pattern *analysis.PatternSymbol) irtypes.PatternDef {
	placeholders := pattern.Placeholders
	if placeholders == nil {
		placeholders = []string{}
	}
	return irtypes.PatternDef{
		Name:         pattern.Name,
		Doc:          normalizeDocPtr(pattern.Docstring),
		Deprecation:  convertDeprecation(pattern.Deprecated),
		Template:     pattern.Template,
		Placeholders: placeholders,
	}
}

// ============================================================================
// RPC CONVERSION
// ============================================================================

func convertRPCDef(rpc *analysis.RPCSymbol) irtypes.RpcDef {
	return irtypes.RpcDef{
		Name:        rpc.Name,
		Doc:         normalizeDocPtr(rpc.Docstring),
		Deprecation: convertDeprecation(rpc.Deprecated),
	}
}

func convertProcedure(
	proc *analysis.ProcSymbol,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
	rpcName string,
) irtypes.ProcedureDef {
	return irtypes.ProcedureDef{
		RpcName:      rpcName,
		Name:         proc.Name,
		Doc:          normalizeDocPtr(proc.Docstring),
		Deprecation:  convertDeprecation(proc.Deprecated),
		InputFields:  flattenBlockFields(proc.Input, types, enums),
		OutputFields: flattenBlockFields(proc.Output, types, enums),
	}
}

func convertStream(
	stream *analysis.StreamSymbol,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
	rpcName string,
) irtypes.StreamDef {
	return irtypes.StreamDef{
		RpcName:      rpcName,
		Name:         stream.Name,
		Doc:          normalizeDocPtr(stream.Docstring),
		Deprecation:  convertDeprecation(stream.Deprecated),
		InputFields:  flattenBlockFields(stream.Input, types, enums),
		OutputFields: flattenBlockFields(stream.Output, types, enums),
	}
}

// ============================================================================
// DEPRECATION CONVERSION
// ============================================================================

func convertDeprecation(dep *analysis.DeprecationInfo) *string {
	if dep == nil {
		return nil
	}
	return &dep.Message
}

// ============================================================================
// SORTING FUNCTIONS
// ============================================================================

func sortTypes(types []irtypes.TypeDef) {
	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})
}

func sortEnums(enums []irtypes.EnumDef) {
	sort.Slice(enums, func(i, j int) bool {
		return enums[i].Name < enums[j].Name
	})
}

func sortConstants(constants []irtypes.ConstantDef) {
	sort.Slice(constants, func(i, j int) bool {
		return constants[i].Name < constants[j].Name
	})
}

func sortPatterns(patterns []irtypes.PatternDef) {
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Name < patterns[j].Name
	})
}

func sortRPCs(rpcs []irtypes.RpcDef) {
	sort.Slice(rpcs, func(i, j int) bool {
		return rpcs[i].Name < rpcs[j].Name
	})
}

func sortProcedures(procs []irtypes.ProcedureDef) {
	sort.Slice(procs, func(i, j int) bool {
		if procs[i].RpcName != procs[j].RpcName {
			return procs[i].RpcName < procs[j].RpcName
		}
		return procs[i].Name < procs[j].Name
	})
}

func sortStreams(streams []irtypes.StreamDef) {
	sort.Slice(streams, func(i, j int) bool {
		if streams[i].RpcName != streams[j].RpcName {
			return streams[i].RpcName < streams[j].RpcName
		}
		return streams[i].Name < streams[j].Name
	})
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
