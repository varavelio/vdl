package pieces

import (
	"encoding/json"
	"testing"
	"time"
)

// TestStructure includes all nullable types
type TestStructure struct {
	Text    Optional[string]  `json:"text,omitempty"`
	Number  Optional[int]     `json:"number,omitempty"`
	Decimal Optional[float64] `json:"decimal,omitempty"`
	Flag    Optional[bool]    `json:"flag,omitempty"`
	Generic Optional[[]int]   `json:"generic,omitempty"`
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

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && (got.Present != tt.expected.Present || got.Value != tt.expected.Value) {
				t.Errorf("got %+v, want %+v", got, tt.expected)
			}
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

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && (got.Present != tt.expected.Present || got.Value != tt.expected.Value) {
				t.Errorf("got %+v, want %+v", got, tt.expected)
			}
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

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && (got.Present != tt.expected.Present || got.Value != tt.expected.Value) {
				t.Errorf("got %+v, want %+v", got, tt.expected)
			}
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

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && (got.Present != tt.expected.Present || got.Value != tt.expected.Value) {
				t.Errorf("got %+v, want %+v", got, tt.expected)
			}
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

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && (got.Present != tt.expected.Present || !got.Value.Equal(tt.expected.Value)) {
				t.Errorf("got %+v, want %+v", got, tt.expected)
			}
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

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if got.Text.Present != tt.expected.Text.Present || got.Text.Value != tt.expected.Text.Value {
					t.Errorf("Text: got %+v, want %+v", got.Text, tt.expected.Text)
				}
				if got.Number.Present != tt.expected.Number.Present || got.Number.Value != tt.expected.Number.Value {
					t.Errorf("Number: got %+v, want %+v", got.Number, tt.expected.Number)
				}
				if got.Decimal.Present != tt.expected.Decimal.Present || got.Decimal.Value != tt.expected.Decimal.Value {
					t.Errorf("Decimal: got %+v, want %+v", got.Decimal, tt.expected.Decimal)
				}
				if got.Flag.Present != tt.expected.Flag.Present || got.Flag.Value != tt.expected.Flag.Value {
					t.Errorf("Flag: got %+v, want %+v", got.Flag, tt.expected.Flag)
				}
			}
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
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			if string(got) != tt.expected {
				t.Errorf("got %v, want %v", string(got), tt.expected)
			}

			// Verify that it can be deserialized back
			var roundTrip TestStructure
			err = json.Unmarshal(got, &roundTrip)
			if err != nil {
				t.Errorf("Unmarshal() error in roundtrip = %v", err)
			}
		})
	}
}
