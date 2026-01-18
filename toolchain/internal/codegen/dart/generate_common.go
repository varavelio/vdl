package dart

import (
	"fmt"
	"strings"

	"github.com/uforg/ufogenkit"
	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

// dartTypeLiteral returns the Dart type literal for a field, considering custom/built-in/inline and arrays/optionality.
func dartTypeLiteral(parentTypeName string, field schema.FieldDefinition) string {
	isNamed := field.IsNamed()
	isInline := field.IsInline()

	// default to dynamic
	typeLiteral := "dynamic"

	if isNamed && field.IsCustomType() {
		typeLiteral = *field.TypeName
	}

	if isNamed && field.IsBuiltInType() {
		switch *field.TypeName {
		case "string":
			typeLiteral = "String"
		case "int":
			typeLiteral = "int"
		case "float":
			typeLiteral = "double"
		case "bool":
			typeLiteral = "bool"
		case "datetime":
			typeLiteral = "DateTime"
		}
	}

	if isInline {
		typeLiteral = parentTypeName + strutil.ToPascalCase(field.Name)
	}

	if field.IsArray {
		typeLiteral = fmt.Sprintf("List<%s>", typeLiteral)
	}

	if field.Optional {
		typeLiteral = typeLiteral + "?"
	}

	return typeLiteral
}

// dartFromJsonExpr returns the Dart expression to parse a single field from JSON value.
func dartFromJsonExpr(parentTypeName string, field schema.FieldDefinition, jsonAccessor string) string {
	isNamed := field.IsNamed()
	isInline := field.IsInline()

	switch {
	case isNamed && field.IsCustomType():
		// Custom named type
		if field.IsArray {
			// List of custom named type
			return fmt.Sprintf("((%s as List).map((e) => %s.fromJson((e as Map).cast<String, dynamic>())).toList())", jsonAccessor, *field.TypeName)
		}
		return fmt.Sprintf("%s.fromJson((%s as Map).cast<String, dynamic>())", *field.TypeName, jsonAccessor)

	case isNamed && field.IsBuiltInType():
		// Built-in types
		switch *field.TypeName {
		case "string":
			if field.IsArray {
				return fmt.Sprintf("((%s as List).map((e) => e as String).toList())", jsonAccessor)
			}
			return fmt.Sprintf("%s as String", jsonAccessor)
		case "int":
			if field.IsArray {
				return fmt.Sprintf("((%s as List).map((e) => (e as num).toInt()).toList())", jsonAccessor)
			}
			return fmt.Sprintf("(%s as num).toInt()", jsonAccessor)
		case "float":
			if field.IsArray {
				return fmt.Sprintf("((%s as List).map((e) => (e as num).toDouble()).toList())", jsonAccessor)
			}
			return fmt.Sprintf("(%s as num).toDouble()", jsonAccessor)
		case "bool":
			if field.IsArray {
				return fmt.Sprintf("((%s as List).map((e) => e as bool).toList())", jsonAccessor)
			}
			return fmt.Sprintf("%s as bool", jsonAccessor)
		case "datetime":
			if field.IsArray {
				return fmt.Sprintf("((%s as List).map((e) => DateTime.parse(e as String)).toList())", jsonAccessor)
			}
			return fmt.Sprintf("DateTime.parse(%s as String)", jsonAccessor)
		}

	case isInline:
		namePascal := strutil.ToPascalCase(field.Name)
		if field.IsArray {
			return fmt.Sprintf("((%s as List).map((e) => %s%s.fromJson((e as Map).cast<String, dynamic>())).toList())", jsonAccessor, parentTypeName, namePascal)
		}
		return fmt.Sprintf("%s%s.fromJson((%s as Map).cast<String, dynamic>())", parentTypeName, namePascal, jsonAccessor)
	}

	// Fallback dynamic
	if field.IsArray {
		return fmt.Sprintf("List<dynamic>.from(%s as List)", jsonAccessor)
	}
	return jsonAccessor
}

// dartToJsonExpr returns the Dart expression to serialise a field to JSON.
// varName is the variable expression to serialise (must be non-nullable in the call site).
func dartToJsonExpr(field schema.FieldDefinition, varName string) string {
	isNamed := field.IsNamed()
	isInline := field.IsInline()

	switch {
	case isNamed && field.IsCustomType():
		if field.IsArray {
			return fmt.Sprintf("%s.map((e) => e.toJson()).toList()", varName)
		}
		return fmt.Sprintf("%s.toJson()", varName)
	case isNamed && field.IsBuiltInType():
		if *field.TypeName == "datetime" {
			if field.IsArray {
				return fmt.Sprintf("%s.map((e) => e.toUtc().toIso8601String()).toList()", varName)
			}
			return fmt.Sprintf("%s.toUtc().toIso8601String()", varName)
		}
		return varName
	case isInline:
		if field.IsArray {
			return fmt.Sprintf("%s.map((e) => e.toJson()).toList()", varName)
		}
		return fmt.Sprintf("%s.toJson()", varName)
	}
	if field.IsArray {
		return varName
	}
	return varName
}

// renderDartType renders a Dart class for given fields, including a short description,
// a factory constructor to hydrate from JSON and a toJson method for serialisation.
func renderDartType(parentName, name, desc string, fields []schema.FieldDefinition) string {
	name = parentName + name

	og := ufogenkit.NewGenKit().WithSpaces(2)
	if desc != "" {
		og.Line("/// " + strings.ReplaceAll(desc, "\n", "\n/// "))
	}
	og.Linef("class %s {", name)
	og.Block(func() {
		// Fields
		for _, field := range fields {
			fieldName := strutil.ToCamelCase(field.Name)
			typeLit := dartTypeLiteral(name, field)
			// Field description if present in schema
			if field.Doc != nil && strings.TrimSpace(*field.Doc) != "" {
				og.Line("/// " + strings.ReplaceAll(strings.TrimSpace(*field.Doc), "\n", "\n/// "))
			}
			og.Linef("final %s %s;", typeLit, fieldName)
		}
		og.Break()

		// Constructor
		og.Linef("/// Creates a new %s instance.", name)
		if len(fields) == 0 {
			og.Linef("const %s();", name)
		} else {
			og.Linef("const %s({", name)
			og.Block(func() {
				for _, field := range fields {
					fieldName := strutil.ToCamelCase(field.Name)
					isRequired := !field.Optional
					if isRequired {
						og.Linef("required this.%s,", fieldName)
					} else {
						og.Linef("this.%s,", fieldName)
					}
				}
			})
			og.Line("});")
		}

		og.Break()

		// fromJson factory
		og.Linef("/// Hydrates a %s from a JSON map.", name)
		og.Linef("factory %s.fromJson(Map<String, dynamic> json) {", name)
		og.Block(func() {
			for _, field := range fields {
				fieldName := strutil.ToCamelCase(field.Name)
				jsonKey := field.Name
				jsonAccessor := fmt.Sprintf("json['%s']", jsonKey)
				parseExpr := dartFromJsonExpr(name, field, jsonAccessor)
				if field.Optional {
					og.Linef("final %s = json.containsKey('%s') && %s != null ? %s : null;", fieldName, jsonKey, jsonAccessor, parseExpr)
				} else {
					og.Linef("final %s = %s;", fieldName, parseExpr)
				}
			}
			// return
			og.Linef("return %s(", name)
			og.Block(func() {
				for _, field := range fields {
					fieldName := strutil.ToCamelCase(field.Name)
					og.Linef("%s: %s,", fieldName, fieldName)
				}
			})
			og.Line(");")
		})
		og.Line("}")
		og.Break()

		// toJson method
		og.Linef("/// Serialises this %s to a JSON map compatible with the server.", name)
		og.Line("Map<String, dynamic> toJson() {")
		og.Block(func() {
			og.Line("final _data = <String, dynamic>{};")
			for _, field := range fields {
				fieldName := strutil.ToCamelCase(field.Name)
				jsonKey := field.Name
				if field.Optional {
					local := "__v_" + fieldName
					og.Linef("final %s = %s;", local, fieldName)
					ser := dartToJsonExpr(field, local)
					og.Linef("if (%s != null) _data['%s'] = %s;", local, jsonKey, ser)
				} else {
					ser := dartToJsonExpr(field, fieldName)
					og.Linef("_data['%s'] = %s;", jsonKey, ser)
				}
			}
			og.Line("return _data;")
		})
		og.Line("}")
	})
	og.Line("}")
	og.Break()

	// Children inline types
	for _, field := range fields {
		if !field.IsInline() {
			continue
		}
		// Inline type inherits description if present on the containing field
		childDesc := ""
		if field.Doc != nil {
			childDesc = strings.TrimSpace(*field.Doc)
		}
		og.Line(renderDartType(name, strutil.ToPascalCase(field.Name), childDesc, field.TypeInline.Fields))
	}

	return og.String()
}

// renderDeprecatedDart writes a deprecated doc line if provided.
func renderDeprecatedDart(g *ufogenkit.GenKit, deprecated *string) {
	if deprecated == nil {
		return
	}
	desc := "@deprecated "
	if *deprecated == "" {
		desc += "This is deprecated and should not be used in new code."
	} else {
		desc += *deprecated
	}
	g.Line("///")
	for _, line := range strings.Split(desc, "\n") {
		g.Linef("/// %s", line)
	}
}
