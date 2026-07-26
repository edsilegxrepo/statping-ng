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
