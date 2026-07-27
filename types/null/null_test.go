package null

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestJSONMarshal(t *testing.T) {
	tests := []struct {
		Input        any
		ExpectedJSON string
	}{
		{
			Input:        NewNullBool(true),
			ExpectedJSON: `true`,
		},
		{
			Input:        NewNullBool(false),
			ExpectedJSON: `false`,
		},
		{
			Input:        NewNullFloat64(0.994),
			ExpectedJSON: `0.994`,
		},
		{
			Input:        NewNullFloat64(0),
			ExpectedJSON: `0`,
		},
		{
			Input:        NewNullInt64(42),
			ExpectedJSON: `42`,
		},
		{
			Input:        NewNullInt64(0),
			ExpectedJSON: `0`,
		},
		{
			Input:        NewNullString("test"),
			ExpectedJSON: `"test"`,
		},
		{
			Input:        NewNullString(""),
			ExpectedJSON: `""`,
		},
	}

	for _, test := range tests {
		str, err := json.Marshal(test.Input)
		require.Nil(t, err)
		assert.Equal(t, test.ExpectedJSON, string(str))
	}
}

func TestJSONMarshalInvalid(t *testing.T) {
	t.Run("invalid NullInt64 marshals to null", func(t *testing.T) {
		val := NullInt64{}
		val.Int64 = 42
		val.Valid = false
		b, err := json.Marshal(val)
		require.NoError(t, err)
		assert.Equal(t, "null", string(b))
	})

	t.Run("invalid NullFloat64 marshals to null", func(t *testing.T) {
		val := NullFloat64{}
		val.Float64 = 3.14
		val.Valid = false
		b, err := json.Marshal(val)
		require.NoError(t, err)
		assert.Equal(t, "null", string(b))
	})

	t.Run("invalid NullBool marshals to null", func(t *testing.T) {
		val := NullBool{}
		val.Bool = true
		val.Valid = false
		b, err := json.Marshal(val)
		require.NoError(t, err)
		assert.Equal(t, "null", string(b))
	})

	t.Run("invalid NullString marshals to null", func(t *testing.T) {
		val := NullString{}
		val.String = "test"
		val.Valid = false
		b, err := json.Marshal(val)
		require.NoError(t, err)
		assert.Equal(t, "null", string(b))
	})
}

func TestJSONUnmarshal(t *testing.T) {
	t.Run("unmarshal NullInt64", func(t *testing.T) {
		var val NullInt64
		err := json.Unmarshal([]byte("42"), &val)
		require.NoError(t, err)
		assert.Equal(t, int64(42), val.Int64)
		assert.True(t, val.Valid)
	})

	t.Run("unmarshal NullFloat64", func(t *testing.T) {
		var val NullFloat64
		err := json.Unmarshal([]byte("3.14"), &val)
		require.NoError(t, err)
		assert.Equal(t, 3.14, val.Float64)
		assert.True(t, val.Valid)
	})

	t.Run("unmarshal NullBool", func(t *testing.T) {
		var val NullBool
		err := json.Unmarshal([]byte("true"), &val)
		require.NoError(t, err)
		assert.True(t, val.Bool)
		assert.True(t, val.Valid)
	})

	t.Run("unmarshal NullString", func(t *testing.T) {
		var val NullString
		err := json.Unmarshal([]byte(`"hello"`), &val)
		require.NoError(t, err)
		assert.Equal(t, "hello", val.String)
		assert.True(t, val.Valid)
	})

	t.Run("unmarshal invalid data sets Valid false", func(t *testing.T) {
		var val NullInt64
		err := json.Unmarshal([]byte(`"not a number"`), &val)
		assert.Error(t, err)
		assert.False(t, val.Valid)
	})
}

func TestYAMLMarshal(t *testing.T) {
	t.Run("valid NullInt64", func(t *testing.T) {
		val := NewNullInt64(42)
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("invalid NullInt64", func(t *testing.T) {
		val := NullInt64{}
		val.Int64 = 42
		val.Valid = false
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("valid NullFloat64", func(t *testing.T) {
		val := NewNullFloat64(3.14)
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("invalid NullFloat64", func(t *testing.T) {
		val := NullFloat64{}
		val.Float64 = 3.14
		val.Valid = false
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		assert.Equal(t, 0.0, result)
	})

	t.Run("valid NullBool", func(t *testing.T) {
		val := NewNullBool(true)
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("invalid NullBool", func(t *testing.T) {
		val := NullBool{}
		val.Bool = true
		val.Valid = false
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("valid NullString", func(t *testing.T) {
		val := NewNullString("test")
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("invalid NullString", func(t *testing.T) {
		val := NullString{}
		val.String = "test"
		val.Valid = false
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func TestYAMLUnmarshal(t *testing.T) {
	t.Run("unmarshal NullInt64", func(t *testing.T) {
		var val NullInt64
		err := yaml.Unmarshal([]byte("42"), &val)
		require.NoError(t, err)
		assert.Equal(t, int64(42), val.Int64)
		assert.True(t, val.Valid)
	})

	t.Run("unmarshal NullFloat64", func(t *testing.T) {
		var val NullFloat64
		err := yaml.Unmarshal([]byte("3.14"), &val)
		require.NoError(t, err)
		assert.Equal(t, 3.14, val.Float64)
		assert.True(t, val.Valid)
	})

	t.Run("unmarshal NullBool", func(t *testing.T) {
		var val NullBool
		err := yaml.Unmarshal([]byte("true"), &val)
		require.NoError(t, err)
		assert.True(t, val.Bool)
		assert.True(t, val.Valid)
	})

	t.Run("unmarshal NullString", func(t *testing.T) {
		var val NullString
		err := yaml.Unmarshal([]byte("hello"), &val)
		require.NoError(t, err)
		assert.Equal(t, "hello", val.String)
		assert.True(t, val.Valid)
	})
}

func TestNullStringValue(t *testing.T) {
	val := NewNullString("test value")
	driverVal, err := val.Value()
	require.NoError(t, err)
	assert.Equal(t, "test value", driverVal)
}

func TestNewNullBool(t *testing.T) {
	val := NewNullBool(true)
	assert.True(t, val.Bool)
	assert.True(t, val.Valid)

	val = NewNullBool(false)
	assert.False(t, val.Bool)
	assert.True(t, val.Valid)
}

func TestNewNullInt64(t *testing.T) {
	val := NewNullInt64(29)
	assert.Equal(t, int64(29), val.Int64)
	assert.True(t, val.Valid)
}

func TestNewNullString(t *testing.T) {
	val := NewNullString("statping.com")
	assert.Equal(t, "statping.com", val.String)
	assert.True(t, val.Valid)
}

func TestNewNullFloat64(t *testing.T) {
	val := NewNullFloat64(42.222)
	assert.Equal(t, float64(42.222), val.Float64)
	assert.True(t, val.Valid)
}

// TestYAMLUnmarshalErrors tests error paths in UnmarshalYAML methods
func TestYAMLUnmarshalErrors(t *testing.T) {
	t.Run("NullInt64 type mismatch", func(t *testing.T) {
		var val NullInt64
		err := yaml.Unmarshal([]byte("not_a_number"), &val)
		assert.Error(t, err)
	})

	t.Run("NullFloat64 type mismatch", func(t *testing.T) {
		var val NullFloat64
		err := yaml.Unmarshal([]byte("not_a_float"), &val)
		assert.Error(t, err)
	})

	t.Run("NullBool type mismatch", func(t *testing.T) {
		var val NullBool
		err := yaml.Unmarshal([]byte("not_a_bool"), &val)
		assert.Error(t, err)
	})

	t.Run("NullString with invalid YAML", func(t *testing.T) {
		var val NullString
		// YAML with unquoted special characters that can't parse as string
		err := yaml.Unmarshal([]byte(":\n  - invalid"), &val)
		assert.Error(t, err)
	})
}

// TestJSONUnmarshalNull tests unmarshaling JSON null values
// The current implementation unmarshals to the underlying type, which keeps previous values
// when null is encountered (no error, value unchanged)
func TestJSONUnmarshalNull(t *testing.T) {
	t.Run("NullInt64 from null preserves value", func(t *testing.T) {
		var val NullInt64
		val.Int64 = 123
		val.Valid = true
		err := json.Unmarshal([]byte("null"), &val)
		require.NoError(t, err)
		// null doesn't change the underlying int64 value
		assert.Equal(t, int64(123), val.Int64)
		assert.True(t, val.Valid)
	})

	t.Run("NullFloat64 from null preserves value", func(t *testing.T) {
		var val NullFloat64
		val.Float64 = 3.14
		val.Valid = true
		err := json.Unmarshal([]byte("null"), &val)
		require.NoError(t, err)
		assert.Equal(t, 3.14, val.Float64)
		assert.True(t, val.Valid)
	})

	t.Run("NullBool from null preserves value", func(t *testing.T) {
		var val NullBool
		val.Bool = true
		val.Valid = true
		err := json.Unmarshal([]byte("null"), &val)
		require.NoError(t, err)
		assert.True(t, val.Bool)
		assert.True(t, val.Valid)
	})

	t.Run("NullString from null preserves value", func(t *testing.T) {
		var val NullString
		val.String = "test"
		val.Valid = true
		err := json.Unmarshal([]byte("null"), &val)
		require.NoError(t, err)
		assert.Equal(t, "test", val.String)
		assert.True(t, val.Valid)
	})

	t.Run("NullInt64 zero-value from null", func(t *testing.T) {
		var val NullInt64
		err := json.Unmarshal([]byte("null"), &val)
		require.NoError(t, err)
		assert.Equal(t, int64(0), val.Int64)
		assert.True(t, val.Valid)
	})
}

// TestJSONUnmarshalInvalidJSON tests unmarshaling malformed JSON
func TestJSONUnmarshalInvalidJSON(t *testing.T) {
	t.Run("NullInt64 invalid JSON", func(t *testing.T) {
		var val NullInt64
		err := json.Unmarshal([]byte("{invalid}"), &val)
		assert.Error(t, err)
		assert.False(t, val.Valid)
	})

	t.Run("NullFloat64 invalid JSON", func(t *testing.T) {
		var val NullFloat64
		err := json.Unmarshal([]byte("{invalid}"), &val)
		assert.Error(t, err)
		assert.False(t, val.Valid)
	})

	t.Run("NullBool invalid JSON", func(t *testing.T) {
		var val NullBool
		err := json.Unmarshal([]byte("{invalid}"), &val)
		assert.Error(t, err)
		assert.False(t, val.Valid)
	})

	t.Run("NullString invalid JSON", func(t *testing.T) {
		var val NullString
		err := json.Unmarshal([]byte("{invalid}"), &val)
		assert.Error(t, err)
		assert.False(t, val.Valid)
	})
}

// TestJSONRoundTrip tests marshal then unmarshal produces same value
func TestJSONRoundTrip(t *testing.T) {
	t.Run("NullInt64 round trip", func(t *testing.T) {
		original := NewNullInt64(12345)
		b, err := json.Marshal(original)
		require.NoError(t, err)

		var restored NullInt64
		err = json.Unmarshal(b, &restored)
		require.NoError(t, err)
		assert.Equal(t, original.Int64, restored.Int64)
		assert.True(t, restored.Valid)
	})

	t.Run("NullFloat64 round trip", func(t *testing.T) {
		original := NewNullFloat64(123.456)
		b, err := json.Marshal(original)
		require.NoError(t, err)

		var restored NullFloat64
		err = json.Unmarshal(b, &restored)
		require.NoError(t, err)
		assert.Equal(t, original.Float64, restored.Float64)
		assert.True(t, restored.Valid)
	})

	t.Run("NullBool round trip true", func(t *testing.T) {
		original := NewNullBool(true)
		b, err := json.Marshal(original)
		require.NoError(t, err)

		var restored NullBool
		err = json.Unmarshal(b, &restored)
		require.NoError(t, err)
		assert.Equal(t, original.Bool, restored.Bool)
		assert.True(t, restored.Valid)
	})

	t.Run("NullBool round trip false", func(t *testing.T) {
		original := NewNullBool(false)
		b, err := json.Marshal(original)
		require.NoError(t, err)

		var restored NullBool
		err = json.Unmarshal(b, &restored)
		require.NoError(t, err)
		assert.Equal(t, original.Bool, restored.Bool)
		assert.True(t, restored.Valid)
	})

	t.Run("NullString round trip", func(t *testing.T) {
		original := NewNullString("test string value")
		b, err := json.Marshal(original)
		require.NoError(t, err)

		var restored NullString
		err = json.Unmarshal(b, &restored)
		require.NoError(t, err)
		assert.Equal(t, original.String, restored.String)
		assert.True(t, restored.Valid)
	})

	t.Run("NullString round trip empty string", func(t *testing.T) {
		original := NewNullString("")
		b, err := json.Marshal(original)
		require.NoError(t, err)

		var restored NullString
		err = json.Unmarshal(b, &restored)
		require.NoError(t, err)
		assert.Equal(t, original.String, restored.String)
		assert.True(t, restored.Valid)
	})
}

// TestYAMLRoundTrip tests YAML marshal then unmarshal
// Note: MarshalYAML returns yaml.Marshal(value) which is []byte, not the raw value
// This tests direct unmarshal which works with raw values
func TestYAMLRoundTrip(t *testing.T) {
	t.Run("NullInt64 direct unmarshal", func(t *testing.T) {
		var val NullInt64
		err := yaml.Unmarshal([]byte("99"), &val)
		require.NoError(t, err)
		assert.Equal(t, int64(99), val.Int64)
		assert.True(t, val.Valid)
	})

	t.Run("NullFloat64 direct unmarshal", func(t *testing.T) {
		var val NullFloat64
		err := yaml.Unmarshal([]byte("99.99"), &val)
		require.NoError(t, err)
		assert.Equal(t, 99.99, val.Float64)
		assert.True(t, val.Valid)
	})

	t.Run("NullBool direct unmarshal", func(t *testing.T) {
		var val NullBool
		err := yaml.Unmarshal([]byte("true"), &val)
		require.NoError(t, err)
		assert.True(t, val.Bool)
		assert.True(t, val.Valid)
	})

	t.Run("NullString direct unmarshal", func(t *testing.T) {
		var val NullString
		err := yaml.Unmarshal([]byte("hello world"), &val)
		require.NoError(t, err)
		assert.Equal(t, "hello world", val.String)
		assert.True(t, val.Valid)
	})

	t.Run("NullInt64 marshal valid", func(t *testing.T) {
		val := NewNullInt64(42)
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		// MarshalYAML returns yaml.Marshal(42) = []byte("42\n")
		assert.Equal(t, []byte("42\n"), result)
	})

	t.Run("NullFloat64 marshal valid", func(t *testing.T) {
		val := NewNullFloat64(3.14)
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		assert.Equal(t, []byte("3.14\n"), result)
	})

	t.Run("NullBool marshal valid", func(t *testing.T) {
		val := NewNullBool(true)
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		assert.Equal(t, []byte("true\n"), result)
	})

	t.Run("NullString marshal valid", func(t *testing.T) {
		val := NewNullString("test")
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		assert.Equal(t, []byte("test\n"), result)
	})
}

// TestYAMLUnmarshalZeroValues tests unmarshaling zero/empty YAML values
func TestYAMLUnmarshalZeroValues(t *testing.T) {
	t.Run("NullInt64 zero", func(t *testing.T) {
		var val NullInt64
		err := yaml.Unmarshal([]byte("0"), &val)
		require.NoError(t, err)
		assert.Equal(t, int64(0), val.Int64)
		assert.True(t, val.Valid)
	})

	t.Run("NullFloat64 zero", func(t *testing.T) {
		var val NullFloat64
		err := yaml.Unmarshal([]byte("0.0"), &val)
		require.NoError(t, err)
		assert.Equal(t, 0.0, val.Float64)
		assert.True(t, val.Valid)
	})

	t.Run("NullBool false", func(t *testing.T) {
		var val NullBool
		err := yaml.Unmarshal([]byte("false"), &val)
		require.NoError(t, err)
		assert.False(t, val.Bool)
		assert.True(t, val.Valid)
	})

	t.Run("NullString empty via quotes", func(t *testing.T) {
		var val NullString
		err := yaml.Unmarshal([]byte(`""`), &val)
		require.NoError(t, err)
		assert.Equal(t, "", val.String)
		assert.True(t, val.Valid)
	})
}

// TestNullStringValueEmpty tests Value() for empty string
func TestNullStringValueEmpty(t *testing.T) {
	val := NewNullString("")
	driverVal, err := val.Value()
	require.NoError(t, err)
	assert.Equal(t, "", driverVal)
}

// TestNullStringValueInvalid tests Value() when Valid is false
func TestNullStringValueInvalid(t *testing.T) {
	val := NullString{}
	val.String = "should be ignored"
	val.Valid = false
	driverVal, err := val.Value()
	require.NoError(t, err)
	// Value() returns the string regardless of Valid flag per current implementation
	assert.Equal(t, "should be ignored", driverVal)
}

// TestJSONUnmarshalEmptyString tests unmarshaling empty JSON string
func TestJSONUnmarshalEmptyString(t *testing.T) {
	t.Run("NullString empty JSON string", func(t *testing.T) {
		var val NullString
		err := json.Unmarshal([]byte(`""`), &val)
		require.NoError(t, err)
		assert.Equal(t, "", val.String)
		assert.True(t, val.Valid)
	})
}

// TestJSONMarshalZeroValues tests marshaling zero values that are valid
func TestJSONMarshalZeroValues(t *testing.T) {
	t.Run("NullInt64 zero valid", func(t *testing.T) {
		val := NewNullInt64(0)
		b, err := json.Marshal(val)
		require.NoError(t, err)
		assert.Equal(t, "0", string(b))
	})

	t.Run("NullFloat64 zero valid", func(t *testing.T) {
		val := NewNullFloat64(0.0)
		b, err := json.Marshal(val)
		require.NoError(t, err)
		assert.Equal(t, "0", string(b))
	})

	t.Run("NullBool false valid", func(t *testing.T) {
		val := NewNullBool(false)
		b, err := json.Marshal(val)
		require.NoError(t, err)
		assert.Equal(t, "false", string(b))
	})
}

// TestYAMLMarshalZeroValues tests YAML marshaling of zero values
func TestYAMLMarshalZeroValues(t *testing.T) {
	t.Run("NullInt64 zero valid", func(t *testing.T) {
		val := NewNullInt64(0)
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		// MarshalYAML returns yaml.Marshal(0) which is []byte("0\n")
		assert.NotNil(t, result)
	})

	t.Run("NullFloat64 zero valid", func(t *testing.T) {
		val := NewNullFloat64(0.0)
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("NullString empty valid", func(t *testing.T) {
		val := NewNullString("")
		result, err := val.MarshalYAML()
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// TestJSONUnmarshalTypeMismatch tests type coercion/mismatch scenarios
func TestJSONUnmarshalTypeMismatch(t *testing.T) {
	t.Run("NullInt64 from float JSON", func(t *testing.T) {
		var val NullInt64
		err := json.Unmarshal([]byte("3.14"), &val)
		// JSON floats cannot unmarshal to int64
		assert.Error(t, err)
		assert.False(t, val.Valid)
	})

	t.Run("NullFloat64 from int JSON", func(t *testing.T) {
		var val NullFloat64
		err := json.Unmarshal([]byte("42"), &val)
		// int can be unmarshaled to float64
		require.NoError(t, err)
		assert.Equal(t, 42.0, val.Float64)
		assert.True(t, val.Valid)
	})

	t.Run("NullBool from string JSON", func(t *testing.T) {
		var val NullBool
		err := json.Unmarshal([]byte(`"true"`), &val)
		assert.Error(t, err)
		assert.False(t, val.Valid)
	})

	t.Run("NullString from number JSON", func(t *testing.T) {
		var val NullString
		err := json.Unmarshal([]byte("42"), &val)
		assert.Error(t, err)
		assert.False(t, val.Valid)
	})
}

// TestYAMLUnmarshalNegativeValues tests negative number handling
func TestYAMLUnmarshalNegativeValues(t *testing.T) {
	t.Run("NullInt64 negative", func(t *testing.T) {
		var val NullInt64
		err := yaml.Unmarshal([]byte("-999"), &val)
		require.NoError(t, err)
		assert.Equal(t, int64(-999), val.Int64)
		assert.True(t, val.Valid)
	})

	t.Run("NullFloat64 negative", func(t *testing.T) {
		var val NullFloat64
		err := yaml.Unmarshal([]byte("-3.14159"), &val)
		require.NoError(t, err)
		assert.Equal(t, -3.14159, val.Float64)
		assert.True(t, val.Valid)
	})
}

// TestJSONUnmarshalNegativeValues tests negative number handling in JSON
func TestJSONUnmarshalNegativeValues(t *testing.T) {
	t.Run("NullInt64 negative", func(t *testing.T) {
		var val NullInt64
		err := json.Unmarshal([]byte("-12345"), &val)
		require.NoError(t, err)
		assert.Equal(t, int64(-12345), val.Int64)
		assert.True(t, val.Valid)
	})

	t.Run("NullFloat64 negative", func(t *testing.T) {
		var val NullFloat64
		err := json.Unmarshal([]byte("-99.99"), &val)
		require.NoError(t, err)
		assert.Equal(t, -99.99, val.Float64)
		assert.True(t, val.Valid)
	})
}

// TestJSONMarshalNegativeValues tests marshaling negative numbers
func TestJSONMarshalNegativeValues(t *testing.T) {
	t.Run("NullInt64 negative", func(t *testing.T) {
		val := NewNullInt64(-42)
		b, err := json.Marshal(val)
		require.NoError(t, err)
		assert.Equal(t, "-42", string(b))
	})

	t.Run("NullFloat64 negative", func(t *testing.T) {
		val := NewNullFloat64(-3.14)
		b, err := json.Marshal(val)
		require.NoError(t, err)
		assert.Equal(t, "-3.14", string(b))
	})
}

// TestJSONMarshalLargeValues tests marshaling edge case values
func TestJSONMarshalLargeValues(t *testing.T) {
	t.Run("NullInt64 max value", func(t *testing.T) {
		val := NewNullInt64(9223372036854775807)
		b, err := json.Marshal(val)
		require.NoError(t, err)
		assert.Equal(t, "9223372036854775807", string(b))
	})

	t.Run("NullInt64 min value", func(t *testing.T) {
		val := NewNullInt64(-9223372036854775808)
		b, err := json.Marshal(val)
		require.NoError(t, err)
		assert.Equal(t, "-9223372036854775808", string(b))
	})
}

// TestJSONUnmarshalSpecialFloats tests special float values
func TestJSONUnmarshalSpecialFloats(t *testing.T) {
	t.Run("NullFloat64 very small", func(t *testing.T) {
		var val NullFloat64
		err := json.Unmarshal([]byte("0.000001"), &val)
		require.NoError(t, err)
		assert.InDelta(t, 0.000001, val.Float64, 1e-10)
		assert.True(t, val.Valid)
	})

	t.Run("NullFloat64 scientific notation", func(t *testing.T) {
		var val NullFloat64
		err := json.Unmarshal([]byte("1.5e10"), &val)
		require.NoError(t, err)
		assert.Equal(t, 1.5e10, val.Float64)
		assert.True(t, val.Valid)
	})
}

// TestNullStringSpecialChars tests strings with special characters
func TestNullStringSpecialChars(t *testing.T) {
	t.Run("JSON with unicode", func(t *testing.T) {
		original := NewNullString("Hello 世界 🌍")
		b, err := json.Marshal(original)
		require.NoError(t, err)

		var restored NullString
		err = json.Unmarshal(b, &restored)
		require.NoError(t, err)
		assert.Equal(t, original.String, restored.String)
	})

	t.Run("JSON with escape sequences", func(t *testing.T) {
		original := NewNullString("line1\nline2\ttab")
		b, err := json.Marshal(original)
		require.NoError(t, err)

		var restored NullString
		err = json.Unmarshal(b, &restored)
		require.NoError(t, err)
		assert.Equal(t, original.String, restored.String)
	})

	t.Run("JSON with quotes", func(t *testing.T) {
		original := NewNullString(`He said "hello"`)
		b, err := json.Marshal(original)
		require.NoError(t, err)

		var restored NullString
		err = json.Unmarshal(b, &restored)
		require.NoError(t, err)
		assert.Equal(t, original.String, restored.String)
	})
}
