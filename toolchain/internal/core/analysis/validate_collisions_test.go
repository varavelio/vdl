package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

func TestReservationRule_Apply(t *testing.T) {
	t.Run("prefix only", func(t *testing.T) {
		rule := ReservationRule{Kind: SyntheticNameKindEnum, Prefix: "is"}
		assert.Equal(t, "isColor", rule.Apply("Color"))
		assert.Equal(t, "isStatus", rule.Apply("Status"))
	})

	t.Run("suffix only", func(t *testing.T) {
		rule := ReservationRule{Kind: SyntheticNameKindEnum, Suffix: "List"}
		assert.Equal(t, "ColorList", rule.Apply("Color"))
		assert.Equal(t, "StatusList", rule.Apply("Status"))
	})

	t.Run("prefix and suffix", func(t *testing.T) {
		rule := ReservationRule{Kind: SyntheticNameKindEnum, Prefix: "get", Suffix: "Handler"}
		assert.Equal(t, "getColorHandler", rule.Apply("Color"))
	})
}

func TestReservationRule_Description(t *testing.T) {
	t.Run("prefix only", func(t *testing.T) {
		rule := ReservationRule{Prefix: "is"}
		assert.Equal(t, "is<Name>", rule.Description())
	})

	t.Run("suffix only", func(t *testing.T) {
		rule := ReservationRule{Suffix: "List"}
		assert.Equal(t, "<Name>List", rule.Description())
	})

	t.Run("prefix and suffix", func(t *testing.T) {
		rule := ReservationRule{Prefix: "get", Suffix: "Handler"}
		assert.Equal(t, "get<Name>Handler", rule.Description())
	})
}

func TestValidateCollisions(t *testing.T) {
	t.Run("no collision when no user types match synthetic names", func(t *testing.T) {
		symbols := newSymbolTable()
		symbols.enums["Color"] = &EnumSymbol{
			Symbol: Symbol{Name: "Color", File: "test.vdl"},
		}
		symbols.types["User"] = &TypeSymbol{
			Symbol: Symbol{Name: "User", File: "test.vdl"},
		}

		diagnostics := validateCollisions(symbols)
		assert.Empty(t, diagnostics)
	})

	t.Run("detects enum List suffix collision with type", func(t *testing.T) {
		symbols := newSymbolTable()
		symbols.enums["Color"] = &EnumSymbol{
			Symbol: Symbol{Name: "Color", File: "test.vdl", Pos: ast.Position{Line: 1, Column: 1}},
		}
		symbols.types["ColorList"] = &TypeSymbol{
			Symbol: Symbol{Name: "ColorList", File: "test.vdl", Pos: ast.Position{Line: 5, Column: 1}},
		}

		diagnostics := validateCollisions(symbols)
		require.Len(t, diagnostics, 1)
		assert.Equal(t, CodeSyntheticNameCollision, diagnostics[0].Code)
		assert.Contains(t, diagnostics[0].Message, "ColorList")
		assert.Contains(t, diagnostics[0].Message, "Color")
		assert.Contains(t, diagnostics[0].Message, "<Name>List")
	})

	t.Run("no collision with is prefix due to case difference", func(t *testing.T) {
		// The synthetic name isStatus (camelCase) doesn't match IsStatus (PascalCase)
		// because VDL types must be PascalCase, they can't match is<Name> patterns
		symbols := newSymbolTable()
		symbols.enums["Status"] = &EnumSymbol{
			Symbol: Symbol{Name: "Status", File: "test.vdl", Pos: ast.Position{Line: 1, Column: 1}},
		}
		symbols.types["IsStatus"] = &TypeSymbol{
			Symbol: Symbol{Name: "IsStatus", File: "test.vdl", Pos: ast.Position{Line: 5, Column: 1}},
		}

		diagnostics := validateCollisions(symbols)
		// No collision because "isStatus" != "IsStatus" (case-sensitive)
		assert.Empty(t, diagnostics)
	})

	t.Run("detects enum Value suffix collision with type", func(t *testing.T) {
		symbols := newSymbolTable()
		symbols.enums["Priority"] = &EnumSymbol{
			Symbol: Symbol{Name: "Priority", File: "test.vdl", Pos: ast.Position{Line: 1, Column: 1}},
		}
		symbols.types["PriorityValue"] = &TypeSymbol{
			Symbol: Symbol{Name: "PriorityValue", File: "test.vdl", Pos: ast.Position{Line: 5, Column: 1}},
		}

		diagnostics := validateCollisions(symbols)
		require.Len(t, diagnostics, 1)
		assert.Equal(t, CodeSyntheticNameCollision, diagnostics[0].Code)
		assert.Contains(t, diagnostics[0].Message, "PriorityValue")
	})

	t.Run("detects proc Input suffix collision with type", func(t *testing.T) {
		symbols := newSymbolTable()
		symbols.rpcs["Users"] = &RPCSymbol{
			Symbol: Symbol{Name: "Users", File: "test.vdl"},
			Procs: map[string]*ProcSymbol{
				"GetUser": {
					Symbol: Symbol{Name: "GetUser", File: "test.vdl", Pos: ast.Position{Line: 2, Column: 1}},
				},
			},
			Streams: map[string]*StreamSymbol{},
		}
		symbols.types["GetUserInput"] = &TypeSymbol{
			Symbol: Symbol{Name: "GetUserInput", File: "test.vdl", Pos: ast.Position{Line: 10, Column: 1}},
		}

		diagnostics := validateCollisions(symbols)
		require.Len(t, diagnostics, 1)
		assert.Equal(t, CodeSyntheticNameCollision, diagnostics[0].Code)
		assert.Contains(t, diagnostics[0].Message, "GetUserInput")
		assert.Contains(t, diagnostics[0].Message, "GetUser")
		assert.Contains(t, diagnostics[0].Message, "<Name>Input")
	})

	t.Run("detects proc Output suffix collision with type", func(t *testing.T) {
		symbols := newSymbolTable()
		symbols.rpcs["Users"] = &RPCSymbol{
			Symbol: Symbol{Name: "Users", File: "test.vdl"},
			Procs: map[string]*ProcSymbol{
				"CreateUser": {
					Symbol: Symbol{Name: "CreateUser", File: "test.vdl", Pos: ast.Position{Line: 2, Column: 1}},
				},
			},
			Streams: map[string]*StreamSymbol{},
		}
		symbols.types["CreateUserOutput"] = &TypeSymbol{
			Symbol: Symbol{Name: "CreateUserOutput", File: "test.vdl", Pos: ast.Position{Line: 10, Column: 1}},
		}

		diagnostics := validateCollisions(symbols)
		require.Len(t, diagnostics, 1)
		assert.Equal(t, CodeSyntheticNameCollision, diagnostics[0].Code)
		assert.Contains(t, diagnostics[0].Message, "CreateUserOutput")
		assert.Contains(t, diagnostics[0].Message, "CreateUser")
		assert.Contains(t, diagnostics[0].Message, "<Name>Output")
	})

	t.Run("detects stream Input collision with type", func(t *testing.T) {
		symbols := newSymbolTable()
		symbols.rpcs["Events"] = &RPCSymbol{
			Symbol: Symbol{Name: "Events", File: "test.vdl"},
			Procs:  map[string]*ProcSymbol{},
			Streams: map[string]*StreamSymbol{
				"WatchUpdates": {
					Symbol: Symbol{Name: "WatchUpdates", File: "test.vdl", Pos: ast.Position{Line: 2, Column: 1}},
				},
			},
		}
		symbols.types["WatchUpdatesInput"] = &TypeSymbol{
			Symbol: Symbol{Name: "WatchUpdatesInput", File: "test.vdl", Pos: ast.Position{Line: 10, Column: 1}},
		}

		diagnostics := validateCollisions(symbols)
		require.Len(t, diagnostics, 1)
		assert.Equal(t, CodeSyntheticNameCollision, diagnostics[0].Code)
		assert.Contains(t, diagnostics[0].Message, "WatchUpdatesInput")
		assert.Contains(t, diagnostics[0].Message, "WatchUpdates")
	})

	t.Run("detects collision with pattern", func(t *testing.T) {
		symbols := newSymbolTable()
		symbols.enums["Status"] = &EnumSymbol{
			Symbol: Symbol{Name: "Status", File: "test.vdl", Pos: ast.Position{Line: 1, Column: 1}},
		}
		symbols.patterns["StatusList"] = &PatternSymbol{
			Symbol: Symbol{Name: "StatusList", File: "test.vdl", Pos: ast.Position{Line: 5, Column: 1}},
		}

		diagnostics := validateCollisions(symbols)
		require.Len(t, diagnostics, 1)
		assert.Equal(t, CodeSyntheticNameCollision, diagnostics[0].Code)
		assert.Contains(t, diagnostics[0].Message, "pattern")
		assert.Contains(t, diagnostics[0].Message, "StatusList")
	})

	t.Run("multiple collisions reported", func(t *testing.T) {
		symbols := newSymbolTable()
		symbols.enums["Color"] = &EnumSymbol{
			Symbol: Symbol{Name: "Color", File: "test.vdl", Pos: ast.Position{Line: 1, Column: 1}},
		}
		symbols.types["ColorList"] = &TypeSymbol{
			Symbol: Symbol{Name: "ColorList", File: "test.vdl", Pos: ast.Position{Line: 5, Column: 1}},
		}
		symbols.types["ColorValue"] = &TypeSymbol{
			Symbol: Symbol{Name: "ColorValue", File: "test.vdl", Pos: ast.Position{Line: 15, Column: 1}},
		}

		diagnostics := validateCollisions(symbols)
		assert.Len(t, diagnostics, 2)
		for _, d := range diagnostics {
			assert.Equal(t, CodeSyntheticNameCollision, d.Code)
		}
	})

	t.Run("no collision when source and user-defined are the same", func(t *testing.T) {
		// Edge case: if user somehow defines both "Color" and they're the same
		// This shouldn't produce a collision (handled by duplicate detection instead)
		symbols := newSymbolTable()
		symbols.enums["Color"] = &EnumSymbol{
			Symbol: Symbol{Name: "Color", File: "test.vdl"},
		}
		// "Color" enum generates "isColor", "ColorList", "ColorValue"
		// but NOT "Color" itself

		diagnostics := validateCollisions(symbols)
		assert.Empty(t, diagnostics)
	})
}

func TestSyntheticNameRulesConfiguration(t *testing.T) {
	t.Run("enum rules are configured", func(t *testing.T) {
		enumRules := []string{}
		for _, rule := range syntheticNameRules {
			if rule.Kind == SyntheticNameKindEnum {
				enumRules = append(enumRules, rule.Apply("Test"))
			}
		}

		assert.Contains(t, enumRules, "isTest", "should have is<Name> rule")
		assert.Contains(t, enumRules, "TestList", "should have <Name>List rule")
		assert.Contains(t, enumRules, "TestValue", "should have <Name>Value rule")
	})

	t.Run("proc rules are configured", func(t *testing.T) {
		procRules := []string{}
		for _, rule := range syntheticNameRules {
			if rule.Kind == SyntheticNameKindProc {
				procRules = append(procRules, rule.Apply("Test"))
			}
		}

		assert.Contains(t, procRules, "TestInput", "should have <Name>Input rule")
		assert.Contains(t, procRules, "TestOutput", "should have <Name>Output rule")
	})
}
