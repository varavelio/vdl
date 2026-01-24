package golang

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// =============================================================================
// Type Conversion: IR TypeRef → Go Type String
// =============================================================================

// typeRefToGo converts an IR TypeRef to its Go type string representation.
// parentTypeName is used to generate names for inline object types.
func typeRefToGo(parentTypeName string, tr ir.TypeRef) string {
	switch tr.Kind {
	case ir.TypeKindPrimitive:
		return primitiveToGo(tr.Primitive)

	case ir.TypeKindType:
		return tr.Type

	case ir.TypeKindEnum:
		return tr.Enum

	case ir.TypeKindArray:
		// Build array prefix based on dimensions
		arrayPrefix := strings.Repeat("[]", tr.ArrayDimensions)
		elementType := typeRefToGo(parentTypeName, *tr.ArrayItem)
		return arrayPrefix + elementType

	case ir.TypeKindMap:
		valueType := typeRefToGo(parentTypeName, *tr.MapValue)
		return "map[string]" + valueType

	case ir.TypeKindObject:
		// Inline objects get a generated name based on parent
		return parentTypeName
	}

	return "any"
}

// typeRefToPreGo converts an IR TypeRef to its "pre" Go type string representation.
// The "pre" types are used for unmarshaling and validation before transformation.
func typeRefToPreGo(parentTypeName string, tr ir.TypeRef) string {
	switch tr.Kind {
	case ir.TypeKindPrimitive:
		return primitiveToGo(tr.Primitive)

	case ir.TypeKindType:
		return "pre" + tr.Type

	case ir.TypeKindEnum:
		// Enums don't need pre-types, they're validated at parse time
		return tr.Enum

	case ir.TypeKindArray:
		// Build array prefix based on dimensions
		arrayPrefix := strings.Repeat("[]", tr.ArrayDimensions)
		elementType := typeRefToPreGo(parentTypeName, *tr.ArrayItem)
		return arrayPrefix + elementType

	case ir.TypeKindMap:
		valueType := typeRefToPreGo(parentTypeName, *tr.MapValue)
		return "map[string]" + valueType

	case ir.TypeKindObject:
		// Inline objects get a generated pre-name based on parent
		return "pre" + parentTypeName
	}

	return "any"
}

// primitiveToGo converts an IR primitive type to its Go equivalent.
func primitiveToGo(p ir.PrimitiveType) string {
	switch p {
	case ir.PrimitiveString:
		return "string"
	case ir.PrimitiveInt:
		return "int64"
	case ir.PrimitiveFloat:
		return "float64"
	case ir.PrimitiveBool:
		return "bool"
	case ir.PrimitiveDatetime:
		return "time.Time"
	}
	return "any"
}

// needsPreType returns true if the TypeRef requires pre-type validation.
// Primitives and enums don't need pre-types.
func needsPreType(tr ir.TypeRef) bool {
	switch tr.Kind {
	case ir.TypeKindPrimitive, ir.TypeKindEnum:
		return false
	case ir.TypeKindType, ir.TypeKindObject:
		return true
	case ir.TypeKindArray:
		return needsPreType(*tr.ArrayItem)
	case ir.TypeKindMap:
		return needsPreType(*tr.MapValue)
	}
	return false
}

// =============================================================================
// Field Rendering
// =============================================================================

// renderField generates the Go struct field code for a single IR field.
func renderField(parentTypeName string, field ir.Field) string {
	namePascal := strutil.ToPascalCase(field.Name)
	nameCamel := strutil.ToCamelCase(field.Name)

	// Calculate the inline type name for objects
	inlineTypeName := parentTypeName + namePascal
	typeLiteral := typeRefToGo(inlineTypeName, field.Type)

	if field.Optional {
		typeLiteral = fmt.Sprintf("Optional[%s]", typeLiteral)
	}

	// JSON tag
	jsonTag := fmt.Sprintf(" `json:\"%s\"`", nameCamel)
	if field.Optional {
		jsonTag = fmt.Sprintf(" `json:\"%s,omitzero\"`", nameCamel)
	}

	doc := renderDocString(field.Doc, false)
	result := fmt.Sprintf("%s %s", namePascal, typeLiteral)
	return doc + result + jsonTag
}

// renderPreField generates the Go struct field code for a pre-type field.
// All fields in pre-types are wrapped in Optional for validation.
func renderPreField(parentTypeName string, field ir.Field) string {
	namePascal := strutil.ToPascalCase(field.Name)
	nameCamel := strutil.ToCamelCase(field.Name)

	// Calculate the inline type name for objects
	inlineTypeName := parentTypeName + namePascal
	typeLiteral := typeRefToPreGo(inlineTypeName, field.Type)

	// All pre-type fields are optional for validation
	typeLiteral = fmt.Sprintf("Optional[%s]", typeLiteral)

	jsonTag := fmt.Sprintf(" `json:\"%s,omitzero\"`", nameCamel)
	result := fmt.Sprintf("%s %s", namePascal, typeLiteral)
	return result + jsonTag
}

// getInlineObject returns the inline object definition if the type is an object
// or contains one (recursively in arrays/maps). Returns nil otherwise.
func getInlineObject(tr ir.TypeRef) *ir.InlineObject {
	switch tr.Kind {
	case ir.TypeKindObject:
		return tr.Object
	case ir.TypeKindArray:
		return getInlineObject(*tr.ArrayItem)
	case ir.TypeKindMap:
		return getInlineObject(*tr.MapValue)
	}
	return nil
}

// =============================================================================
// Type Structure Rendering
// =============================================================================

// renderType renders a complete type definition with all its fields.
func renderType(parentName, name, desc string, fields []ir.Field) string {
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

	// Render children inline types
	for _, field := range fields {
		if inlineObj := getInlineObject(field.Type); inlineObj != nil {
			childName := fullName + strutil.ToPascalCase(field.Name)
			g.Line(renderType("", childName, "", inlineObj.Fields))
		}
	}

	return g.String()
}

// renderPreType renders a pre-type definition with validation and transform methods.
func renderPreType(parentName, name string, fields []ir.Field) string {
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
		if inlineObj := getInlineObject(field.Type); inlineObj != nil {
			childName := fullName + strutil.ToPascalCase(field.Name)
			g.Line(renderPreType("", childName, inlineObj.Fields))
		}
	}

	// Render validate function
	g.Line(renderValidateFunc(fullName, fields))

	// Render transform function
	g.Line(renderTransformFunc(fullName, fields))

	return g.String()
}

// renderValidateFunc generates the validate method for a pre-type.
func renderValidateFunc(typeName string, fields []ir.Field) string {
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
			needsPre := needsPreType(field.Type)

			g.Linef(`// Validation for field "%s"`, field.Name)

			if isRequired {
				g.Linef("if !p.%s.Present {", fieldName)
				g.Block(func() {
					g.Linef("return errorMissingRequiredField(\"field %s is required\")", field.Name)
				})
				g.Line("}")
			}

			if needsPre {
				g.Linef("if p.%s.Present {", fieldName)
				g.Block(func() {
					source := fmt.Sprintf("p.%s.Value", fieldName)
					renderNestedValidation(g, source, field.Type, field.Name)
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
func renderNestedValidation(g *gen.Generator, source string, tr ir.TypeRef, fieldName string) {
	switch tr.Kind {
	case ir.TypeKindType, ir.TypeKindObject:
		g.Linef("if err := %s.validate(); err != nil {", source)
		g.Block(func() {
			g.Linef("return errorMissingRequiredField(\"field %s: \" + err.Error())", fieldName)
		})
		g.Line("}")

	case ir.TypeKindArray:
		if needsPreType(*tr.ArrayItem) {
			renderArrayValidation(g, source, tr.ArrayDimensions, *tr.ArrayItem, fieldName)
		}

	case ir.TypeKindMap:
		if needsPreType(*tr.MapValue) {
			g.Linef("for key, value := range %s {", source)
			g.Block(func() {
				// Optimization: if it's a direct object, use the key in the error message
				if tr.MapValue.Kind == ir.TypeKindType || tr.MapValue.Kind == ir.TypeKindObject {
					g.Line("if err := value.validate(); err != nil {")
					g.Block(func() {
						g.Linef("return errorMissingRequiredField(\"field %s[\" + key + \"]: \" + err.Error())", fieldName)
					})
					g.Line("}")
				} else {
					renderNestedValidation(g, "value", *tr.MapValue, fieldName)
				}
			})
			g.Line("}")
		}
	}
}

// renderArrayValidation recursively validates array elements based on dimensions.
func renderArrayValidation(g *gen.Generator, source string, dims int, itemType ir.TypeRef, fieldName string) {
	if dims == 0 {
		renderNestedValidation(g, source, itemType, fieldName)
		return
	}

	g.Linef("for _, item := range %s {", source)
	g.Block(func() {
		renderArrayValidation(g, "item", dims-1, itemType, fieldName)
	})
	g.Line("}")
}

// renderTransformFunc generates the transform method for a pre-type.
func renderTransformFunc(typeName string, fields []ir.Field) string {
	g := gen.New().WithTabs()
	g.Linef("// transform transforms the pre%s type to the final %s type", typeName, typeName)
	g.Linef("func (p *pre%s) transform() %s {", typeName, typeName)
	g.Block(func() {
		g.Line("// Transformations")
		for _, field := range fields {
			fieldName := strutil.ToPascalCase(field.Name)
			fieldNameTemp := "trans" + fieldName
			isRequired := !field.Optional
			needsPre := needsPreType(field.Type)

			if !needsPre {
				// Simple extraction for primitives and enums
				if isRequired {
					g.Linef("%s := p.%s.Value", fieldNameTemp, fieldName)
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
func renderFieldTransform(g *gen.Generator, field ir.Field, fieldName, tempName, parentType string) {
	isRequired := !field.Optional
	goType := typeRefToGo(parentType+fieldName, field.Type)

	source := fmt.Sprintf("p.%s.Value", fieldName)
	if !isRequired {
		g.Linef("%s := Optional[%s]{Present: p.%s.Present}", tempName, goType, fieldName)
		g.Linef("if p.%s.Present {", fieldName)
		g.Block(func() {
			valTemp := "val" + strutil.ToPascalCase(fieldName)
			g.Linef("var %s %s", valTemp, goType)
			renderValueTransform(g, source, valTemp, field.Type, parentType+fieldName, "tmp")
			g.Linef("%s.Value = %s", tempName, valTemp)
		})
		g.Line("}")
	} else {
		g.Linef("var %s %s", tempName, goType)
		renderValueTransform(g, source, tempName, field.Type, parentType+fieldName, "tmp")
	}
}

// renderValueTransform generates code to transform source (of pre-type) to dest (of final type).
func renderValueTransform(g *gen.Generator, source, dest string, tr ir.TypeRef, ctxName string, tempPrefix string) {
	if !needsPreType(tr) {
		g.Linef("%s = %s", dest, source)
		return
	}

	switch tr.Kind {
	case ir.TypeKindType, ir.TypeKindObject:
		g.Linef("%s = %s.transform()", dest, source)

	case ir.TypeKindArray:
		renderArrayTransform(g, source, dest, tr.ArrayDimensions, *tr.ArrayItem, ctxName, tempPrefix)

	case ir.TypeKindMap:
		destType := typeRefToGo(ctxName, tr)
		g.Linef("%s = make(%s)", dest, destType)
		g.Linef("for k, v := range %s {", source)
		g.Block(func() {
			itemType := *tr.MapValue
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
func renderArrayTransform(g *gen.Generator, source, dest string, dims int, itemType ir.TypeRef, ctxName string, tempPrefix string) {
	if dims == 0 {
		renderValueTransform(g, source, dest, itemType, ctxName, tempPrefix)
		return
	}

	// Calculate type for destination slice at this level
	synthType := ir.TypeRef{
		Kind:            ir.TypeKindArray,
		ArrayDimensions: dims,
		ArrayItem:       &itemType,
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
			synthNext := ir.TypeRef{
				Kind:            ir.TypeKindArray,
				ArrayDimensions: dims - 1,
				ArrayItem:       &itemType,
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
func renderDeprecated(g *gen.Generator, deprecated *ir.Deprecation) {
	if deprecated == nil {
		return
	}

	desc := "Deprecated: "
	if deprecated.Message == "" {
		desc += "This is deprecated and should not be used in new code."
	} else {
		desc += deprecated.Message
	}

	g.Line("//")
	renderMultilineComment(g, desc)
}
