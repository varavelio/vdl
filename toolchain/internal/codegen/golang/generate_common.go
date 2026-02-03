package golang

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// =============================================================================
// Type Conversion: IR TypeRef → Go Type String
// =============================================================================

// typeRefToGo converts an IR TypeRef to its Go type string representation.
// parentTypeName is used to generate names for inline object types.
func typeRefToGo(parentTypeName string, tr irtypes.TypeRef) string {
	switch tr.Kind {
	case irtypes.TypeKindPrimitive:
		return primitiveToGo(tr.GetPrimitiveName())

	case irtypes.TypeKindType:
		return tr.GetTypeName()

	case irtypes.TypeKindEnum:
		return tr.GetEnumName()

	case irtypes.TypeKindArray:
		// Build array prefix based on dimensions
		arrayPrefix := strings.Repeat("[]", int(tr.GetArrayDims()))
		elementType := typeRefToGo(parentTypeName, tr.GetArrayType())
		return arrayPrefix + elementType

	case irtypes.TypeKindMap:
		valueType := typeRefToGo(parentTypeName, tr.GetMapType())
		return "map[string]" + valueType

	case irtypes.TypeKindObject:
		// Inline objects get a generated name based on parent
		return parentTypeName
	}

	return "any"
}

// typeRefToPreGo converts an IR TypeRef to its "pre" Go type string representation.
// The "pre" types are used for unmarshaling and validation before transformation.
func typeRefToPreGo(parentTypeName string, tr irtypes.TypeRef) string {
	switch tr.Kind {
	case irtypes.TypeKindPrimitive:
		return primitiveToGo(tr.GetPrimitiveName())

	case irtypes.TypeKindType:
		return "pre" + tr.GetTypeName()

	case irtypes.TypeKindEnum:
		// Enums don't need pre-types, they're validated at parse time
		return tr.GetEnumName()

	case irtypes.TypeKindArray:
		// Build array prefix based on dimensions
		arrayPrefix := strings.Repeat("[]", int(tr.GetArrayDims()))
		elementType := typeRefToPreGo(parentTypeName, tr.GetArrayType())
		return arrayPrefix + elementType

	case irtypes.TypeKindMap:
		valueType := typeRefToPreGo(parentTypeName, tr.GetMapType())
		return "map[string]" + valueType

	case irtypes.TypeKindObject:
		// Inline objects get a generated pre-name based on parent
		return "pre" + parentTypeName
	}

	return "any"
}

// primitiveToGo converts an IR primitive type to its Go equivalent.
func primitiveToGo(p irtypes.PrimitiveType) string {
	switch p {
	case irtypes.PrimitiveTypeString:
		return "string"
	case irtypes.PrimitiveTypeInt:
		return "int64"
	case irtypes.PrimitiveTypeFloat:
		return "float64"
	case irtypes.PrimitiveTypeBool:
		return "bool"
	case irtypes.PrimitiveTypeDatetime:
		return "time.Time"
	}
	return "any"
}

// needsPreType returns true if the TypeRef requires pre-type validation.
// Primitives and enums don't need pre-types.
func needsPreType(tr irtypes.TypeRef) bool {
	switch tr.Kind {
	case irtypes.TypeKindPrimitive, irtypes.TypeKindEnum:
		return false
	case irtypes.TypeKindType, irtypes.TypeKindObject:
		return true
	case irtypes.TypeKindArray:
		return needsPreType(tr.GetArrayType())
	case irtypes.TypeKindMap:
		return needsPreType(tr.GetMapType())
	}
	return false
}

// =============================================================================
// Field Rendering
// =============================================================================

// renderField generates the Go struct field code for a single IR field.
func renderField(parentTypeName string, field irtypes.Field) string {
	namePascal := strutil.ToPascalCase(field.Name)
	nameCamel := strutil.ToCamelCase(field.Name)

	// Calculate the inline type name for objects
	inlineTypeName := parentTypeName + namePascal
	typeLiteral := typeRefToGo(inlineTypeName, field.TypeRef)

	// Optional fields use pointers
	if field.Optional {
		typeLiteral = "*" + typeLiteral
	}

	// JSON tag: optional fields get omitempty
	jsonTag := fmt.Sprintf(" `json:\"%s\"`", nameCamel)
	if field.Optional {
		jsonTag = fmt.Sprintf(" `json:\"%s,omitempty\"`", nameCamel)
	}

	doc := renderDocString(field.GetDoc(), false)
	result := fmt.Sprintf("%s %s", namePascal, typeLiteral)
	return doc + result + jsonTag
}

// renderPreField generates the Go struct field code for a pre-type field.
// All fields in pre-types are pointers for validation.
func renderPreField(parentTypeName string, field irtypes.Field) string {
	namePascal := strutil.ToPascalCase(field.Name)
	nameCamel := strutil.ToCamelCase(field.Name)

	// Calculate the inline type name for objects
	inlineTypeName := parentTypeName + namePascal
	typeLiteral := typeRefToPreGo(inlineTypeName, field.TypeRef)

	// All pre-type fields are pointers for validation
	typeLiteral = "*" + typeLiteral

	jsonTag := fmt.Sprintf(" `json:\"%s,omitempty\"`", nameCamel)
	result := fmt.Sprintf("%s %s", namePascal, typeLiteral)
	return result + jsonTag
}

// getInlineObject returns the inline object fields if the type is an object
// or contains one (recursively in arrays/maps). Returns nil otherwise.
func getInlineObject(tr irtypes.TypeRef) *[]irtypes.Field {
	switch tr.Kind {
	case irtypes.TypeKindObject:
		return tr.ObjectFields
	case irtypes.TypeKindArray:
		return getInlineObject(tr.GetArrayType())
	case irtypes.TypeKindMap:
		return getInlineObject(tr.GetMapType())
	}
	return nil
}

// =============================================================================
// Type Structure Rendering
// =============================================================================

// renderType renders a complete type definition with all its fields.
func renderType(parentName, name, desc string, fields []irtypes.Field) string {
	fullName := parentName + name

	g := gen.New().WithTabs()
	renderMultilineComment(g, desc)
	g.Linef("type %s struct {", fullName)
	g.Block(func() {
		for _, field := range fields {
			g.Line(renderField(fullName, field))
		}
	})
	g.Line("}")
	g.Break()

	// Render accessor methods for this type
	g.Line(renderAccessors(fullName, fields))

	// Render children inline types
	for _, field := range fields {
		if inlineFields := getInlineObject(field.TypeRef); inlineFields != nil {
			childName := fullName + strutil.ToPascalCase(field.Name)
			g.Line(renderType("", childName, "", *inlineFields))
		}
	}

	return g.String()
}

// renderPreType renders a pre-type definition with validation and transform methods.
func renderPreType(parentName, name string, fields []irtypes.Field) string {
	fullName := parentName + name

	g := gen.New().WithTabs()
	g.Linef("// pre%s is the version of %s previous to the required field validation", fullName, fullName)
	g.Linef("type pre%s struct {", fullName)
	g.Block(func() {
		for _, field := range fields {
			g.Line(renderPreField(fullName, field))
		}
	})
	g.Line("}")
	g.Break()

	// Render children inline pre-types
	for _, field := range fields {
		if inlineFields := getInlineObject(field.TypeRef); inlineFields != nil {
			childName := fullName + strutil.ToPascalCase(field.Name)
			g.Line(renderPreType("", childName, *inlineFields))
		}
	}

	// Render validate function
	g.Line(renderValidateFunc(fullName, fields))

	// Render transform function
	g.Line(renderTransformFunc(fullName, fields))

	return g.String()
}

// renderValidateFunc generates the validate method for a pre-type.
func renderValidateFunc(typeName string, fields []irtypes.Field) string {
	g := gen.New().WithTabs()
	g.Linef("// validate validates the required fields of %s", typeName)
	g.Linef("func (p *pre%s) validate() error {", typeName)
	g.Block(func() {
		g.Line("if p == nil {")
		g.Block(func() {
			g.Linef("return errorMissingRequiredField(\"pre%s is nil\")", typeName)
		})
		g.Line("}")
		g.Break()

		for _, field := range fields {
			fieldName := strutil.ToPascalCase(field.Name)
			isRequired := !field.Optional
			needsPre := needsPreType(field.TypeRef)

			g.Linef(`// Validation for field "%s"`, field.Name)

			if isRequired {
				g.Linef("if p.%s == nil {", fieldName)
				g.Block(func() {
					g.Linef("return errorMissingRequiredField(\"field %s is required\")", field.Name)
				})
				g.Line("}")
			}

			if needsPre {
				g.Linef("if p.%s != nil {", fieldName)
				g.Block(func() {
					source := fmt.Sprintf("p.%s", fieldName)
					// Pre-type fields are pointers, so isPointer is true
					renderNestedValidation(g, source, field.TypeRef, field.Name, true)
				})
				g.Line("}")
			}

			g.Break()
		}

		g.Line("return nil")
	})
	g.Line("}")

	return g.String()
}

// renderNestedValidation renders validation code for nested types.
// The source should be the field access expression (e.g., "p.Field" for pointer fields)
// isPointer indicates whether the source is a pointer that needs dereferencing for range operations
func renderNestedValidation(g *gen.Generator, source string, tr irtypes.TypeRef, fieldName string, isPointer bool) {
	switch tr.Kind {
	case irtypes.TypeKindType, irtypes.TypeKindObject:
		g.Linef("if err := %s.validate(); err != nil {", source)
		g.Block(func() {
			g.Linef("return errorMissingRequiredField(\"field %s: \" + err.Error())", fieldName)
		})
		g.Line("}")

	case irtypes.TypeKindArray:
		if needsPreType(tr.GetArrayType()) {
			// For pointer to slice, dereference first
			rangeSource := source
			if isPointer {
				rangeSource = "*" + source
			}
			renderArrayValidation(g, rangeSource, int(tr.GetArrayDims()), tr.GetArrayType(), fieldName)
		}

	case irtypes.TypeKindMap:
		if needsPreType(tr.GetMapType()) {
			// For pointer to map, dereference first
			rangeSource := source
			if isPointer {
				rangeSource = "*" + source
			}
			mapType := tr.GetMapType()
			// Use key in error message only for direct object types
			if mapType.Kind == irtypes.TypeKindType || mapType.Kind == irtypes.TypeKindObject {
				g.Linef("for key, value := range %s {", rangeSource)
				g.Block(func() {
					g.Line("if err := value.validate(); err != nil {")
					g.Block(func() {
						g.Linef("return errorMissingRequiredField(\"field %s[\" + key + \"]: \" + err.Error())", fieldName)
					})
					g.Line("}")
				})
				g.Line("}")
			} else {
				// For nested types (arrays, maps), we don't include the key in error messages
				// The value from range is not a pointer, so isPointer is false
				g.Linef("for _, value := range %s {", rangeSource)
				g.Block(func() {
					renderNestedValidation(g, "value", mapType, fieldName, false)
				})
				g.Line("}")
			}
		}
	}
}

// renderArrayValidation recursively validates array elements based on dimensions.
func renderArrayValidation(g *gen.Generator, source string, dims int, itemType irtypes.TypeRef, fieldName string) {
	if dims == 0 {
		// When we get here from a range, the source is not a pointer
		renderNestedValidation(g, source, itemType, fieldName, false)
		return
	}

	g.Linef("for _, item := range %s {", source)
	g.Block(func() {
		renderArrayValidation(g, "item", dims-1, itemType, fieldName)
	})
	g.Line("}")
}

// renderTransformFunc generates the transform method for a pre-type.
func renderTransformFunc(typeName string, fields []irtypes.Field) string {
	g := gen.New().WithTabs()
	g.Linef("// transform transforms the pre%s type to the final %s type", typeName, typeName)
	g.Linef("func (p *pre%s) transform() %s {", typeName, typeName)
	g.Block(func() {
		g.Line("// Transformations")
		for _, field := range fields {
			fieldName := strutil.ToPascalCase(field.Name)
			fieldNameTemp := "trans" + fieldName
			isRequired := !field.Optional
			needsPre := needsPreType(field.TypeRef)

			if !needsPre {
				// Simple extraction for primitives and enums
				if isRequired {
					g.Linef("%s := *p.%s", fieldNameTemp, fieldName)
				} else {
					g.Linef("%s := p.%s", fieldNameTemp, fieldName)
				}
				continue
			}

			// Complex transformation for types that need pre-validation
			renderFieldTransform(g, field, fieldName, fieldNameTemp, typeName)
		}

		g.Break()
		g.Line("// Assignments")
		g.Linef("return %s{", typeName)
		g.Block(func() {
			for _, field := range fields {
				fieldName := strutil.ToPascalCase(field.Name)
				fieldNameTemp := "trans" + fieldName
				g.Linef("%s: %s,", fieldName, fieldNameTemp)
			}
		})
		g.Line("}")
	})
	g.Line("}")

	return g.String()
}

// renderFieldTransform renders the transformation code for a single field.
func renderFieldTransform(g *gen.Generator, field irtypes.Field, fieldName, tempName, parentType string) {
	isRequired := !field.Optional
	goType := typeRefToGo(parentType+fieldName, field.TypeRef)

	// For Type and Object, use pointer source since preField is a pointer
	// For Arrays and Maps, use dereferenced source since they are slices/maps directly
	source := fmt.Sprintf("p.%s", fieldName)
	needsDeref := field.TypeRef.Kind == irtypes.TypeKindArray || field.TypeRef.Kind == irtypes.TypeKindMap
	if needsDeref {
		source = fmt.Sprintf("*p.%s", fieldName)
	}

	if !isRequired {
		// Optional field: use pointer type in output
		g.Linef("var %s *%s", tempName, goType)
		g.Linef("if p.%s != nil {", fieldName)
		g.Block(func() {
			valTemp := "val" + strutil.ToPascalCase(fieldName)
			g.Linef("var %s %s", valTemp, goType)
			renderValueTransform(g, source, valTemp, field.TypeRef, parentType+fieldName, "tmp")
			g.Linef("%s = &%s", tempName, valTemp)
		})
		g.Line("}")
	} else {
		g.Linef("var %s %s", tempName, goType)
		renderValueTransform(g, source, tempName, field.TypeRef, parentType+fieldName, "tmp")
	}
}

// renderValueTransform generates code to transform source (of pre-type) to dest (of final type).
func renderValueTransform(g *gen.Generator, source, dest string, tr irtypes.TypeRef, ctxName string, tempPrefix string) {
	if !needsPreType(tr) {
		g.Linef("%s = %s", dest, source)
		return
	}

	switch tr.Kind {
	case irtypes.TypeKindType, irtypes.TypeKindObject:
		g.Linef("%s = %s.transform()", dest, source)

	case irtypes.TypeKindArray:
		renderArrayTransform(g, source, dest, int(tr.GetArrayDims()), tr.GetArrayType(), ctxName, tempPrefix)

	case irtypes.TypeKindMap:
		destType := typeRefToGo(ctxName, tr)
		g.Linef("%s = make(%s)", dest, destType)
		g.Linef("for k, v := range %s {", source)
		g.Block(func() {
			itemType := tr.GetMapType()
			itemDestType := typeRefToGo(ctxName, itemType)
			tempVar := tempPrefix + "_"
			g.Linef("var %s %s", tempVar, itemDestType)
			renderValueTransform(g, "v", tempVar, itemType, ctxName, tempVar)
			g.Linef("%s[k] = %s", dest, tempVar)
		})
		g.Line("}")
	}
}

// renderArrayTransform recursively generates array transformation code handling dimensions.
func renderArrayTransform(g *gen.Generator, source, dest string, dims int, itemType irtypes.TypeRef, ctxName string, tempPrefix string) {
	if dims == 0 {
		renderValueTransform(g, source, dest, itemType, ctxName, tempPrefix)
		return
	}

	// Calculate type for destination slice at this level
	synthType := irtypes.TypeRef{
		Kind:      irtypes.TypeKindArray,
		ArrayDims: irtypes.Ptr(int64(dims)),
		ArrayType: &itemType,
	}
	destType := typeRefToGo(ctxName, synthType)

	g.Linef("%s = make(%s, len(%s))", dest, destType, source)
	g.Linef("for i, v := range %s {", source)
	g.Block(func() {
		// Next level type
		var nextLevelType string
		if dims == 1 {
			nextLevelType = typeRefToGo(ctxName, itemType)
		} else {
			synthNext := irtypes.TypeRef{
				Kind:      irtypes.TypeKindArray,
				ArrayDims: irtypes.Ptr(int64(dims - 1)),
				ArrayType: &itemType,
			}
			nextLevelType = typeRefToGo(ctxName, synthNext)
		}

		tempVar := tempPrefix + "_"
		g.Linef("var %s %s", tempVar, nextLevelType)

		renderArrayTransform(g, "v", tempVar, dims-1, itemType, ctxName, tempVar)

		g.Linef("%s[i] = %s", dest, tempVar)
	})
	g.Line("}")
}

// =============================================================================
// Safe Accessors (Getters)
// =============================================================================

// renderAccessors generates getter methods for all fields in a struct.
// Getters provide nil-safe access to fields, especially for optional pointers.
func renderAccessors(typeName string, fields []irtypes.Field) string {
	g := gen.New().WithTabs()

	for _, field := range fields {
		fieldName := strutil.ToPascalCase(field.Name)
		inlineTypeName := typeName + fieldName
		goType := typeRefToGo(inlineTypeName, field.TypeRef)

		// Generate Get{FieldName}()
		g.Linef("// Get%s returns the value of %s or the zero value if the receiver or field is nil.", fieldName, fieldName)
		g.Linef("func (x *%s) Get%s() %s {", typeName, fieldName, goType)
		g.Block(func() {
			if field.Optional {
				// Optional field: check both receiver and field
				g.Linef("if x != nil && x.%s != nil {", fieldName)
				g.Block(func() {
					g.Linef("return *x.%s", fieldName)
				})
				g.Line("}")
				g.Linef("var zero %s", goType)
				g.Line("return zero")
			} else {
				// Required field: only check receiver
				g.Line("if x != nil {")
				g.Block(func() {
					g.Linef("return x.%s", fieldName)
				})
				g.Line("}")
				g.Linef("var zero %s", goType)
				g.Line("return zero")
			}
		})
		g.Line("}")
		g.Break()

		// Generate Get{FieldName}Or(defaultVal)
		g.Linef("// Get%sOr returns the value of %s or the provided default if the receiver or field is nil.", fieldName, fieldName)
		g.Linef("func (x *%s) Get%sOr(defaultValue %s) %s {", typeName, fieldName, goType, goType)
		g.Block(func() {
			if field.Optional {
				// Optional field: check both receiver and field
				g.Linef("if x != nil && x.%s != nil {", fieldName)
				g.Block(func() {
					g.Linef("return *x.%s", fieldName)
				})
				g.Line("}")
				g.Line("return defaultValue")
			} else {
				// Required field: only check receiver
				g.Line("if x != nil {")
				g.Block(func() {
					g.Linef("return x.%s", fieldName)
				})
				g.Line("}")
				g.Line("return defaultValue")
			}
		})
		g.Line("}")
		g.Break()
	}

	return g.String()
}

// =============================================================================
// Documentation and Comments
// =============================================================================

// renderMultilineComment renders text as a multiline Go comment.
func renderMultilineComment(g *gen.Generator, text string) {
	for line := range strings.SplitSeq(text, "\n") {
		g.Linef("// %s", line)
	}
}

// renderDocString returns a documentation comment string.
func renderDocString(doc string, newLineBefore bool) string {
	if doc == "" {
		return ""
	}

	g := gen.New().WithTabs()
	renderDoc(g, doc, newLineBefore)
	return g.String()
}

// renderDoc renders documentation as Go comments.
func renderDoc(g *gen.Generator, doc string, newLineBefore bool) {
	if doc == "" {
		return
	}

	if newLineBefore {
		g.Line("//")
	}

	renderMultilineComment(g, doc)
}

// renderDeprecated renders a deprecation comment.
func renderDeprecated(g *gen.Generator, deprecated *string) {
	if deprecated == nil {
		return
	}

	desc := "Deprecated: "
	if *deprecated == "" {
		desc += "This is deprecated and should not be used in new code."
	} else {
		desc += *deprecated
	}

	g.Line("//")
	renderMultilineComment(g, desc)
}
