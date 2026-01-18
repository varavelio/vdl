package golang

import (
	"fmt"
	"strings"

	"github.com/uforg/ufogenkit"
	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

// renderField generates the code for a field
func renderField(parentTypeName string, field schema.FieldDefinition) string {
	name := field.Name
	isNamed := field.IsNamed()
	isInline := field.IsInline()

	// Protect against empty fields
	if !isNamed && !isInline {
		return ""
	}

	namePascal := strutil.ToPascalCase(name)
	nameCamel := strutil.ToCamelCase(name)
	isOptional := field.Optional
	isCustomType := field.IsCustomType()
	isBuiltInType := field.IsBuiltInType()

	typeLiteral := "any"

	if isNamed && isCustomType {
		typeLiteral = *field.TypeName
	}

	if isNamed && isBuiltInType {
		switch *field.TypeName {
		case "string":
			typeLiteral = "string"
		case "int":
			typeLiteral = "int"
		case "float":
			typeLiteral = "float64"
		case "bool":
			typeLiteral = "bool"
		case "datetime":
			typeLiteral = "time.Time"
		}
	}

	if isInline {
		typeLiteral = parentTypeName + namePascal
	}

	if field.IsArray {
		typeLiteral = fmt.Sprintf("[]%s", typeLiteral)
	}

	if isOptional {
		typeLiteral = fmt.Sprintf("Optional[%s]", typeLiteral)
	}

	jsonTag := fmt.Sprintf(" `json:\"%s\"`", nameCamel)
	if isOptional {
		jsonTag = fmt.Sprintf(" `json:\"%s,omitempty\"`", nameCamel)
	}

	doc := renderDocString(field.Doc, false)
	result := fmt.Sprintf("%s %s", namePascal, typeLiteral)
	return doc + result + jsonTag
}

// renderType renders a type definition with all its fields
func renderType(
	parentName string,
	name string,
	desc string,
	fields []schema.FieldDefinition,
) string {
	name = parentName + name

	og := ufogenkit.NewGenKit().WithTabs()
	renderMultilineComment(og, desc)
	og.Linef("type %s struct {", name)
	og.Block(func() {
		for _, fieldDef := range fields {
			og.Line(renderField(name, fieldDef))
		}
	})
	og.Line("}")
	og.Break()

	// Render children inline types
	for _, fieldDef := range fields {
		if !fieldDef.IsInline() {
			continue
		}

		og.Line(renderType(name, strutil.ToPascalCase(fieldDef.Name), "", fieldDef.TypeInline.Fields))
	}

	return og.String()
}

// renderPreField generates the code for a field in a pre type
func renderPreField(parentTypeName string, field schema.FieldDefinition) string {
	name := field.Name
	isNamed := field.IsNamed()
	isInline := field.IsInline()

	// Protect against empty fields
	if !isNamed && !isInline {
		return ""
	}

	namePascal := strutil.ToPascalCase(name)
	nameCamel := strutil.ToCamelCase(name)
	isCustomType := field.IsCustomType()
	isBuiltInType := field.IsBuiltInType()

	typeLiteral := "any"

	if isNamed && isCustomType {
		typeLiteral = "pre" + *field.TypeName
	}

	if isNamed && isBuiltInType {
		switch *field.TypeName {
		case "string":
			typeLiteral = "string"
		case "int":
			typeLiteral = "int"
		case "float":
			typeLiteral = "float64"
		case "bool":
			typeLiteral = "bool"
		case "datetime":
			typeLiteral = "time.Time"
		}
	}

	if isInline {
		typeLiteral = "pre" + parentTypeName + namePascal
	}

	if field.IsArray {
		typeLiteral = fmt.Sprintf("[]%s", typeLiteral)
	}

	typeLiteral = fmt.Sprintf("Optional[%s]", typeLiteral)

	jsonTag := fmt.Sprintf(" `json:\"%s,omitempty\"`", nameCamel)
	result := fmt.Sprintf("%s %s", namePascal, typeLiteral)
	return result + jsonTag
}

// renderPreType renders a type definition with all its fields marked as optional
// and helpers to validate the required fields and transform to the final type
func renderPreType(
	parentName string,
	name string,
	fields []schema.FieldDefinition,
) string {
	name = parentName + name

	og := ufogenkit.NewGenKit().WithTabs()
	og.Linef("// pre%s is the version of %s previous to the required field validation", name, name)
	og.Linef("type pre%s struct {", name)
	og.Block(func() {
		for _, fieldDef := range fields {
			og.Line(renderPreField(name, fieldDef))
		}
	})
	og.Line("}")
	og.Break()

	// Render children inline types
	for _, fieldDef := range fields {
		if !fieldDef.IsInline() {
			continue
		}

		og.Line(renderPreType(name, strutil.ToPascalCase(fieldDef.Name), fieldDef.TypeInline.Fields))
	}

	// Render validate function
	og.Linef("// validate validates the required fields of %s", name)
	og.Linef("func (p *pre%s) validate() error {", name)
	og.Block(func() {
		og.Line("if p == nil {")
		og.Block(func() {
			og.Linef("return errorMissingRequiredField(\"pre%s is nil\")", name)
		})
		og.Line("}")
		og.Break()

		// Required fields
		for _, fieldDef := range fields {
			fieldName := strutil.ToPascalCase(fieldDef.Name)
			isRequired := !fieldDef.Optional
			isCustomType := fieldDef.IsCustomType()
			isInline := fieldDef.IsInline()
			isArray := fieldDef.IsArray

			og.Linef(`// Required validations for field "%s"`, fieldDef.Name)

			if isRequired {
				og.Linef("if !p.%s.Present {", fieldName)
				og.Block(func() {
					og.Linef("return errorMissingRequiredField(\"field %s is required\")", fieldDef.Name)
				})
				og.Line("}")
			}

			if (isCustomType || isInline) && !isArray {
				og.Linef("if p.%s.Present {", fieldName)
				og.Block(func() {
					og.Linef("if err := p.%s.Value.validate(); err != nil {", fieldName)
					og.Block(func() {
						og.Linef("return errorMissingRequiredField(\"field %s: \" + err.Error())", fieldDef.Name)
					})
					og.Line("}")
				})
				og.Line("}")
			}

			if (isCustomType || isInline) && isArray {
				og.Linef("if p.%s.Present {", fieldName)

				og.Block(func() {
					og.Linef("for _, item := range p.%s.Value {", fieldName)
					og.Block(func() {
						og.Linef("if err := item.validate(); err != nil {")
						og.Block(func() {
							og.Linef("return errorMissingRequiredField(\"field %s: \" + err.Error())", fieldDef.Name)
						})
						og.Line("}")
					})
					og.Line("}")
				})

				og.Line("}")
			}

			og.Break()
		}

		og.Line("return nil")
	})
	og.Line("}")
	og.Break()

	// Render transform function
	og.Linef("// transform transforms the pre%s type to the final %s type", name, name)
	og.Linef("func (p *pre%s) transform() %s {", name, name)
	og.Block(func() {
		og.Line("// Transformations")
		for _, fieldDef := range fields {
			fieldName := strutil.ToPascalCase(fieldDef.Name)
			fieldNameTemp := "trans" + fieldName
			isRequired := !fieldDef.Optional
			isBuiltinType := fieldDef.IsBuiltInType()
			isCustomType := fieldDef.IsCustomType()
			isInline := fieldDef.IsInline()
			isArray := fieldDef.IsArray

			// Process fields with builtin types
			if isBuiltinType {
				if isRequired {
					og.Linef("%s := p.%s.Value", fieldNameTemp, fieldName)
				} else {
					og.Linef("%s := p.%s", fieldNameTemp, fieldName)
				}
				continue
			}

			// Process fields with custom types or inline fields (non-arrays)
			if (isCustomType || isInline) && !isArray {
				typeName := ""
				if isCustomType {
					typeName = *fieldDef.TypeName
				}
				if isInline {
					typeName = name + fieldName
				}

				if isRequired {
					og.Linef("%s := p.%s.Value.transform()", fieldNameTemp, fieldName)
				} else {
					og.Linef("%s := Optional[%s]{Present: p.%s.Present, Value: p.%s.Value.transform()}",
						fieldNameTemp,
						typeName,
						fieldName,
						fieldName,
					)
				}
				continue
			}

			// Process fields with custom types or inline fields (arrays)
			if (isCustomType || isInline) && isArray {
				typeName := ""
				if isCustomType {
					typeName = *fieldDef.TypeName
				}
				if isInline {
					typeName = name + fieldName
				}

				fieldNameTempArr := "items" + strutil.ToPascalCase(fieldNameTemp)
				og.Linef(
					"%s := make([]%s, len(p.%s.Value))",
					fieldNameTempArr,
					typeName,
					fieldName,
				)

				og.Linef("for index, preItem := range p.%s.Value {", fieldName)
				og.Block(func() {
					og.Linef("%s[index] = preItem.transform()", fieldNameTempArr)
				})
				og.Line("}")

				if isRequired {
					og.Linef("%s := %s", fieldNameTemp, fieldNameTempArr)
				} else {
					og.Linef(
						"%s := Optional[[]%s]{Present: p.%s.Present, Value: %s}",
						fieldNameTemp,
						typeName,
						fieldName,
						fieldNameTempArr,
					)
				}
				continue
			}
		}

		og.Break()
		og.Line("// Assignments")
		og.Linef("return %s{", name)
		og.Block(func() {
			for _, fieldDef := range fields {
				fieldName := strutil.ToPascalCase(fieldDef.Name)
				fieldNameTemp := "trans" + fieldName
				og.Linef("%s: %s,", fieldName, fieldNameTemp)
			}
		})
		og.Line("}")
	})
	og.Line("}")
	og.Break()

	return og.String()
}

// renderMultilineComment receives a text and renders it to the given genkit.GenKit
// as a multiline comment.
func renderMultilineComment(g *ufogenkit.GenKit, text string) {
	for line := range strings.SplitSeq(text, "\n") {
		g.Linef("// %s", line)
	}
}

// renderDocString is the same as renderDoc but it returns a string instead of
// rendering to the given genkit.GenKit.
func renderDocString(doc *string, newLineBefore bool) string {
	if doc == nil {
		return ""
	}

	og := ufogenkit.NewGenKit().WithTabs()
	renderDoc(og, doc, newLineBefore)
	return og.String()
}

// renderDoc receives a pointer to a string and if it is not nil, it will
// render a comment with the documentation to the given genkit.GenKit.
//
// It will normalize the indent and trim the trailing and leading whitespace.
func renderDoc(g *ufogenkit.GenKit, doc *string, newLineBefore bool) {
	if doc == nil {
		return
	}

	if newLineBefore {
		g.Line("//")
	}

	renderMultilineComment(g, strings.TrimSpace(strutil.NormalizeIndent(*doc)))
}

// renderDeprecated receives a pointer to a string and if it is not nil, it will
// render a comment with the deprecated message to the given genkit.GenKit.
func renderDeprecated(g *ufogenkit.GenKit, deprecated *string) {
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
