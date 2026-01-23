package pieces

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestStructure includes all nullable types
type TestStructure struct {
	Text    Optional[string]  `json:"text"`
	Number  Optional[int]     `json:"number"`
	Decimal Optional[float64] `json:"decimal"`
	Flag    Optional[bool]    `json:"flag"`
	Generic Optional[[]int]   `json:"generic"`
}

func TestStringOptional(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected Optional[string]
		wantErr  bool
	}{
		{
			name:     "normal value",
			json:     `"hello"`,
			expected: Optional[string]{Value: "hello", Present: true},
		},
		{
			name:     "empty value",
			json:     `""`,
			expected: Optional[string]{Value: "", Present: true},
		},
		{
			name:     "null value",
			json:     `null`,
			expected: Optional[string]{Present: false},
		},
		{
			name:    "invalid value",
			json:    `123`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Optional[string]
			err := json.Unmarshal([]byte(tt.json), &got)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestIntOptional(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected Optional[int]
		wantErr  bool
	}{
		{
			name:     "positive number",
			json:     `42`,
			expected: Optional[int]{Value: 42, Present: true},
		},
		{
			name:     "negative number",
			json:     `-42`,
			expected: Optional[int]{Value: -42, Present: true},
		},
		{
			name:     "zero",
			json:     `0`,
			expected: Optional[int]{Value: 0, Present: true},
		},
		{
			name:     "null value",
			json:     `null`,
			expected: Optional[int]{Present: false},
		},
		{
			name:    "invalid string value",
			json:    `"123"`,
			wantErr: true,
		},
		{
			name:    "invalid float value",
			json:    `123.45`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Optional[int]
			err := json.Unmarshal([]byte(tt.json), &got)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestFloat64Optional(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected Optional[float64]
		wantErr  bool
	}{
		{
			name:     "integer number",
			json:     `42`,
			expected: Optional[float64]{Value: 42.0, Present: true},
		},
		{
			name:     "decimal number",
			json:     `42.5`,
			expected: Optional[float64]{Value: 42.5, Present: true},
		},
		{
			name:     "negative number",
			json:     `-42.5`,
			expected: Optional[float64]{Value: -42.5, Present: true},
		},
		{
			name:     "zero",
			json:     `0`,
			expected: Optional[float64]{Value: 0, Present: true},
		},
		{
			name:     "null value",
			json:     `null`,
			expected: Optional[float64]{Present: false},
		},
		{
			name:    "invalid value",
			json:    `"123.45"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Optional[float64]
			err := json.Unmarshal([]byte(tt.json), &got)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestBoolOptional(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected Optional[bool]
		wantErr  bool
	}{
		{
			name:     "true",
			json:     `true`,
			expected: Optional[bool]{Value: true, Present: true},
		},
		{
			name:     "false",
			json:     `false`,
			expected: Optional[bool]{Value: false, Present: true},
		},
		{
			name:     "null value",
			json:     `null`,
			expected: Optional[bool]{Present: false},
		},
		{
			name:    "invalid number value",
			json:    `1`,
			wantErr: true,
		},
		{
			name:    "invalid string value",
			json:    `"true"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Optional[bool]
			err := json.Unmarshal([]byte(tt.json), &got)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestOptionalTime(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected Optional[time.Time]
		wantErr  bool
	}{
		{
			name:     "valid time",
			json:     `"2023-05-01T12:34:56Z"`,
			expected: Optional[time.Time]{Value: time.Date(2023, 5, 1, 12, 34, 56, 0, time.UTC), Present: true},
		},
		{
			name:     "null value",
			json:     `null`,
			expected: Optional[time.Time]{Present: false},
		},
		{
			name:    "invalid time format",
			json:    `"2023-05-01"`,
			wantErr: true,
		},
		{
			name:    "invalid type",
			json:    `123`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Optional[time.Time]
			err := json.Unmarshal([]byte(tt.json), &got)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.Present, got.Present)
			require.True(t, got.Value.Equal(tt.expected.Value))
		})
	}
}

func TestComplexStructure(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected TestStructure
		wantErr  bool
	}{
		{
			name: "all fields with value",
			json: `{
                "text": "hello",
                "number": 42,
                "decimal": 42.5,
                "flag": true,
                "generic": [1,2,3]
            }`,
			expected: TestStructure{
				Text:    Optional[string]{Value: "hello", Present: true},
				Number:  Optional[int]{Value: 42, Present: true},
				Decimal: Optional[float64]{Value: 42.5, Present: true},
				Flag:    Optional[bool]{Value: true, Present: true},
				Generic: Optional[[]int]{Value: []int{1, 2, 3}, Present: true},
			},
		},
		{
			name: "all fields null",
			json: `{
                "text": null,
                "number": null,
                "decimal": null,
                "flag": null,
                "generic": null
            }`,
			expected: TestStructure{
				Text:    Optional[string]{Present: false},
				Number:  Optional[int]{Present: false},
				Decimal: Optional[float64]{Present: false},
				Flag:    Optional[bool]{Present: false},
				Generic: Optional[[]int]{Present: false},
			},
		},
		{
			name: "some fields null",
			json: `{
                "text": "hello",
                "number": null,
                "decimal": 42.5,
                "flag": null,
                "generic": [1,2,3]
            }`,
			expected: TestStructure{
				Text:    Optional[string]{Value: "hello", Present: true},
				Number:  Optional[int]{Present: false},
				Decimal: Optional[float64]{Value: 42.5, Present: true},
				Flag:    Optional[bool]{Present: false},
				Generic: Optional[[]int]{Value: []int{1, 2, 3}, Present: true},
			},
		},
		{
			name: "missing fields",
			json: `{
                "text": "hello",
                "decimal": 42.5
            }`,
			expected: TestStructure{
				Text:    Optional[string]{Value: "hello", Present: true},
				Number:  Optional[int]{Present: false},
				Decimal: Optional[float64]{Value: 42.5, Present: true},
				Flag:    Optional[bool]{Present: false},
				Generic: Optional[[]int]{Present: false},
			},
		},
		{
			name: "empty json",
			json: `{}`,
			expected: TestStructure{
				Text:    Optional[string]{Present: false},
				Number:  Optional[int]{Present: false},
				Decimal: Optional[float64]{Present: false},
				Flag:    Optional[bool]{Present: false},
				Generic: Optional[[]int]{Present: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TestStructure
			err := json.Unmarshal([]byte(tt.json), &got)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    TestStructure
		expected string
	}{
		{
			name: "all fields with value",
			input: TestStructure{
				Text:    Optional[string]{Value: "hello", Present: true},
				Number:  Optional[int]{Value: 42, Present: true},
				Decimal: Optional[float64]{Value: 42.5, Present: true},
				Flag:    Optional[bool]{Value: true, Present: true},
				Generic: Optional[[]int]{Value: []int{1, 2, 3}, Present: true},
			},
			expected: `{"text":"hello","number":42,"decimal":42.5,"flag":true,"generic":[1,2,3]}`,
		},
		{
			name: "all fields null",
			input: TestStructure{
				Text:    Optional[string]{Present: false},
				Number:  Optional[int]{Present: false},
				Decimal: Optional[float64]{Present: false},
				Flag:    Optional[bool]{Present: false},
				Generic: Optional[[]int]{Present: false},
			},
			expected: `{"text":null,"number":null,"decimal":null,"flag":null,"generic":null}`,
		},
		{
			name: "mix of values and null",
			input: TestStructure{
				Text:    Optional[string]{Value: "hello", Present: true},
				Number:  Optional[int]{Present: false},
				Decimal: Optional[float64]{Value: 42.5, Present: true},
				Flag:    Optional[bool]{Present: false},
				Generic: Optional[[]int]{Value: []int{1, 2, 3}, Present: true},
			},
			expected: `{"text":"hello","number":null,"decimal":42.5,"flag":null,"generic":[1,2,3]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, string(got))

			// Verify that it can be deserialized back
			var roundTrip TestStructure
			err = json.Unmarshal(got, &roundTrip)
			require.NoError(t, err)
		})
	}
}

func TestSome(t *testing.T) {
	t.Run("string value", func(t *testing.T) {
		opt := Some("hello")
		require.True(t, opt.Present)
		require.Equal(t, "hello", opt.Value)
	})

	t.Run("int value", func(t *testing.T) {
		opt := Some(42)
		require.True(t, opt.Present)
		require.Equal(t, 42, opt.Value)
	})

	t.Run("zero value", func(t *testing.T) {
		opt := Some(0)
		require.True(t, opt.Present)
		require.Equal(t, 0, opt.Value)
	})

	t.Run("empty string", func(t *testing.T) {
		opt := Some("")
		require.True(t, opt.Present)
		require.Equal(t, "", opt.Value)
	})

	t.Run("slice value", func(t *testing.T) {
		opt := Some([]int{1, 2, 3})
		require.True(t, opt.Present)
		require.Len(t, opt.Value, 3)
	})

	t.Run("nil slice", func(t *testing.T) {
		opt := Some[[]int](nil)
		require.True(t, opt.Present)
		require.Nil(t, opt.Value)
	})
}

func TestNone(t *testing.T) {
	t.Run("string type", func(t *testing.T) {
		opt := None[string]()
		require.False(t, opt.Present)
		require.Equal(t, "", opt.Value)
	})

	t.Run("int type", func(t *testing.T) {
		opt := None[int]()
		require.False(t, opt.Present)
		require.Equal(t, 0, opt.Value)
	})

	t.Run("bool type", func(t *testing.T) {
		opt := None[bool]()
		require.False(t, opt.Present)
		require.Equal(t, false, opt.Value)
	})

	t.Run("slice type", func(t *testing.T) {
		opt := None[[]int]()
		require.False(t, opt.Present)
		require.Nil(t, opt.Value)
	})

	t.Run("struct type", func(t *testing.T) {
		type TestStruct struct {
			Field string
		}
		opt := None[TestStruct]()
		require.False(t, opt.Present)
		require.Equal(t, "", opt.Value.Field)
	})
}

func TestOr(t *testing.T) {
	t.Run("present value returns value", func(t *testing.T) {
		opt := Some(42)
		result := opt.Or(100)
		require.Equal(t, 42, result)
	})

	t.Run("absent value returns default", func(t *testing.T) {
		opt := None[int]()
		result := opt.Or(100)
		require.Equal(t, 100, result)
	})

	t.Run("present zero value returns zero", func(t *testing.T) {
		opt := Some(0)
		result := opt.Or(100)
		require.Equal(t, 0, result)
	})

	t.Run("present empty string returns empty", func(t *testing.T) {
		opt := Some("")
		result := opt.Or("default")
		require.Equal(t, "", result)
	})

	t.Run("absent string returns default", func(t *testing.T) {
		opt := None[string]()
		result := opt.Or("default")
		require.Equal(t, "default", result)
	})

	t.Run("present false bool returns false", func(t *testing.T) {
		opt := Some(false)
		result := opt.Or(true)
		require.Equal(t, false, result)
	})

	t.Run("absent bool returns default", func(t *testing.T) {
		opt := None[bool]()
		result := opt.Or(true)
		require.Equal(t, true, result)
	})

	t.Run("present slice returns slice", func(t *testing.T) {
		opt := Some([]int{1, 2, 3})
		result := opt.Or([]int{4, 5, 6})
		require.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("absent slice returns default", func(t *testing.T) {
		opt := None[[]int]()
		result := opt.Or([]int{4, 5, 6})
		require.Equal(t, []int{4, 5, 6}, result)
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("unmarshal empty json", func(t *testing.T) {
		var opt Optional[string]
		err := json.Unmarshal([]byte(""), &opt)
		require.Error(t, err)
	})

	t.Run("unmarshal invalid json", func(t *testing.T) {
		var opt Optional[string]
		err := json.Unmarshal([]byte("{invalid}"), &opt)
		require.Error(t, err)
	})

	t.Run("marshal unmarshalable type", func(t *testing.T) {
		type UnmarshalableType struct {
			Ch chan int
		}
		opt := Some(UnmarshalableType{Ch: make(chan int)})
		_, err := json.Marshal(opt)
		require.Error(t, err)
	})

	t.Run("unmarshal very long null-like string", func(t *testing.T) {
		var opt Optional[string]
		err := json.Unmarshal([]byte(`"nullish"`), &opt)
		require.NoError(t, err)
		require.True(t, opt.Present)
		require.Equal(t, "nullish", opt.Value)
	})

	t.Run("unmarshal short string that is not null", func(t *testing.T) {
		var opt Optional[string]
		err := json.Unmarshal([]byte(`"nul"`), &opt)
		require.NoError(t, err)
		require.True(t, opt.Present)
		require.Equal(t, "nul", opt.Value)
	})

	t.Run("unmarshal array of optionals", func(t *testing.T) {
		type ArrayStruct struct {
			Values []Optional[int] `json:"values"`
		}

		data := `{"values": [1, null, 3, null, 5]}`
		var result ArrayStruct
		err := json.Unmarshal([]byte(data), &result)
		require.NoError(t, err)
		require.Len(t, result.Values, 5)
		require.Equal(t, Optional[int]{Value: 1, Present: true}, result.Values[0])
		require.Equal(t, Optional[int]{Present: false}, result.Values[1])
		require.Equal(t, Optional[int]{Value: 3, Present: true}, result.Values[2])
		require.Equal(t, Optional[int]{Present: false}, result.Values[3])
		require.Equal(t, Optional[int]{Value: 5, Present: true}, result.Values[4])
	})
}
