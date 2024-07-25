package konfetty

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BaseCheck represents a basic check configuration
type BaseCheck struct {
	Name     string
	Interval time.Duration
	Timeout  time.Duration
}

// Defaults provides default values for BaseCheck
func (b BaseCheck) Defaults() any {
	return BaseCheck{
		Interval: 30 * time.Second,
		Timeout:  5 * time.Second,
	}
}

// PingCheck represents a ping check configuration
type PingCheck struct {
	*BaseCheck
	Host string
}

// HTTPCheck represents an HTTP check configuration
type HTTPCheck struct {
	*BaseCheck
	URL    string
	Method string
}

// Defaults provides default values for HTTPCheck
func (h HTTPCheck) Defaults() any {
	return HTTPCheck{
		BaseCheck: &BaseCheck{},
		Method:    "GET",
	}
}

// Checks represents a collection of various checks
type Checks struct {
	Ping []PingCheck
	HTTP []HTTPCheck
}

// Profile represents the main configuration profile
type Profile struct {
	Name   string
	Checks Checks
}

type NestedConfig struct {
	Value int
}

func (n NestedConfig) Defaults() any {
	return NestedConfig{Value: 42}
}

type OuterConfig struct {
	Nested NestedConfig
}

type Embedded struct {
	Value string
}

func (e Embedded) Defaults() any {
	return Embedded{Value: "default"}
}

type TestStruct struct {
	Embedded
}

type ErrorStruct struct{}

func (e ErrorStruct) Defaults() any {
	panic("error in Defaults")
}

func TestFillDefaults(t *testing.T) {
	t.Run("Fill Profile with default values", func(t *testing.T) {
		// Arrange
		profile := &Profile{
			Checks: Checks{
				Ping: []PingCheck{
					{
						BaseCheck: &BaseCheck{Name: "Custom Ping"},
						Host:      "example.com",
					},
					{
						Host: "google.com",
					},
				},
				HTTP: []HTTPCheck{
					{
						BaseCheck: &BaseCheck{Name: "Custom HTTP"},
						URL:       "https://api.example.com",
					},
					{
						URL: "https://api.google.com",
					},
				},
			},
		}

		// Act
		err := fillDefaults(profile, maxDepth)

		// Assert
		require.NoError(t, err)

		// Check Ping defaults
		require.Len(t, profile.Checks.Ping, 2)
		assert.Equal(t, "Custom Ping", profile.Checks.Ping[0].Name)
		assert.Equal(t, 30*time.Second, profile.Checks.Ping[0].Interval)
		assert.Equal(t, 5*time.Second, profile.Checks.Ping[0].Timeout)
		assert.Equal(t, "example.com", profile.Checks.Ping[0].Host)

		assert.Equal(t, "", profile.Checks.Ping[1].Name)
		assert.Equal(t, 30*time.Second, profile.Checks.Ping[1].Interval)
		assert.Equal(t, 5*time.Second, profile.Checks.Ping[1].Timeout)
		assert.Equal(t, "google.com", profile.Checks.Ping[1].Host)

		// Check HTTP defaults
		require.Len(t, profile.Checks.HTTP, 2)
		assert.Equal(t, "Custom HTTP", profile.Checks.HTTP[0].Name)
		assert.Equal(t, 30*time.Second, profile.Checks.HTTP[0].Interval)
		assert.Equal(t, 5*time.Second, profile.Checks.HTTP[0].Timeout)
		assert.Equal(t, "https://api.example.com", profile.Checks.HTTP[0].URL)
		assert.Equal(t, "GET", profile.Checks.HTTP[0].Method)

		assert.Equal(t, "", profile.Checks.HTTP[1].Name)
		assert.Equal(t, 30*time.Second, profile.Checks.HTTP[1].Interval)
		assert.Equal(t, 5*time.Second, profile.Checks.HTTP[1].Timeout)
		assert.Equal(t, "https://api.google.com", profile.Checks.HTTP[1].URL)
		assert.Equal(t, "GET", profile.Checks.HTTP[1].Method)
	})

	t.Run("Fill nested structs with default values", func(t *testing.T) {
		// Arrange
		outer := &OuterConfig{}

		// Act
		err := fillDefaults(outer, maxDepth)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 42, outer.Nested.Value)
	})

	t.Run("Handle nil pointers", func(t *testing.T) {
		// Arrange
		profile := &Profile{
			Checks: Checks{
				Ping: []PingCheck{
					{
						Host: "example.com",
					},
				},
				HTTP: []HTTPCheck{
					{
						URL: "https://api.example.com",
					},
				},
			},
		}

		// Act
		err := fillDefaults(profile, maxDepth)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, profile.Checks.Ping[0].BaseCheck)
		require.NotNil(t, profile.Checks.HTTP[0].BaseCheck)
	})

	t.Run("Max depth exceeded", func(t *testing.T) {
		// Arrange
		type RecursiveStruct struct {
			Next *RecursiveStruct
		}

		cfg := &RecursiveStruct{}
		current := cfg
		for i := 0; i < maxDepth+1; i++ {
			current.Next = &RecursiveStruct{}
			current = current.Next
		}

		// Act
		err := fillDefaults(cfg, maxDepth)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "maximum recursion depth exceeded")
	})

	t.Run("Non-struct type", func(t *testing.T) {
		// Arrange
		value := 42

		// Act
		err := fillDefaults(&value, maxDepth)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 42, value)
	})

	t.Run("Empty struct", func(t *testing.T) {
		// Arrange
		emptyStruct := struct{}{}

		// Act
		err := fillDefaults(&emptyStruct, maxDepth)

		// Assert
		require.NoError(t, err)
		// Empty struct should remain unchanged
		assert.Equal(t, struct{}{}, emptyStruct)
	})

	t.Run("Struct with unexported fields", func(t *testing.T) {
		// Arrange
		type unexportedStruct struct {
			name string
			age  int
		}
		s := unexportedStruct{name: "original", age: 25}

		// Act
		err := fillDefaults(&s, maxDepth)

		// Assert
		require.NoError(t, err)
		// Unexported fields should remain unchanged
		assert.Equal(t, unexportedStruct{name: "original", age: 25}, s)
	})

	t.Run("Struct with interface field", func(t *testing.T) {
		// Arrange
		type interfaceStruct struct {
			Data interface{}
		}
		s := interfaceStruct{}

		// Act
		err := fillDefaults(&s, maxDepth)

		// Assert
		require.NoError(t, err)
		assert.Nil(t, s.Data)
	})

	t.Run("Handle unexported fields", func(t *testing.T) {
		type unexportedField struct {
			exported   string
			unexported string
		}
		value := unexportedField{exported: "test"}
		err := fillDefaults(&value, maxDepth)
		require.NoError(t, err)
		assert.Equal(t, "test", value.exported)
		assert.Empty(t, value.unexported)
	})

	t.Run("Handle embedded fields with nil pointers", func(t *testing.T) {
		type Embedded struct {
			Value string
		}
		type TestStruct struct {
			*Embedded
		}
		value := TestStruct{}
		err := fillDefaults(&value, maxDepth)
		require.NoError(t, err)
		require.NotNil(t, value.Embedded)
		assert.Empty(t, value.Embedded.Value)
	})

	t.Run("Handle nil pointer", func(t *testing.T) {
		type TestStruct struct {
			Value string
		}
		var nilPtr *TestStruct
		v := reflect.ValueOf(&nilPtr).Elem()

		err := fillDefaultsRecursive(v, 0, maxDepth)
		require.NoError(t, err)

		assert.NotNil(t, nilPtr)
		assert.Equal(t, "", nilPtr.Value)
	})

	t.Run("Handle unexported nil pointer", func(t *testing.T) {
		type TestStruct struct {
			value *string
		}
		test := TestStruct{}
		v := reflect.ValueOf(&test).Elem().Field(0)

		err := fillDefaultsRecursive(v, 0, maxDepth)
		require.NoError(t, err)

		assert.Nil(t, test.value)
	})

	t.Run("Handle embedded fields implementing DefaultProvider", func(t *testing.T) {
		value := TestStruct{}
		err := fillDefaults(&value, maxDepth)
		require.NoError(t, err)
		assert.Equal(t, "default", value.Value)
	})

	t.Run("Handle nil pointers in non-embedded fields", func(t *testing.T) {
		type Inner struct {
			Value string
		}
		type TestStruct struct {
			Ptr *Inner
		}
		value := TestStruct{}
		err := fillDefaults(&value, maxDepth)
		require.NoError(t, err)
		require.NotNil(t, value.Ptr)
		assert.Empty(t, value.Ptr.Value)
	})

	t.Run("Handle errors in nested structs", func(t *testing.T) {
		type DeepNested struct {
			Value *int
		}
		type Nested struct {
			Deep DeepNested
		}
		type TestStruct struct {
			Nested Nested
		}
		value := TestStruct{}
		err := fillDefaults(&value, 1) // Set max depth to 1 to force an error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "maximum recursion depth exceeded")
	})

	t.Run("Handle slice of structs", func(t *testing.T) {
		type Item struct {
			Value string
		}
		type TestStruct struct {
			Items []Item
		}
		value := TestStruct{Items: []Item{{}, {}}}
		err := fillDefaults(&value, maxDepth)
		require.NoError(t, err)
		assert.Len(t, value.Items, 2)
	})

	t.Run("Handle map of structs", func(t *testing.T) {
		type Item struct {
			Value string
		}
		type TestStruct struct {
			Items map[string]Item
		}
		value := TestStruct{Items: map[string]Item{"a": {}, "b": {}}}
		err := fillDefaults(&value, maxDepth)
		require.NoError(t, err)
		assert.Len(t, value.Items, 2)
	})

	t.Run("Panic in user-defined Defaults method", func(t *testing.T) {
		value := ErrorStruct{}
		assert.Panics(t, func() {
			_ = fillDefaults(&value, maxDepth)
		})
	})
}

func TestFillFromDefaults(t *testing.T) {
	t.Run("Fill struct with defaults", func(t *testing.T) {
		// Arrange
		type TestStruct struct {
			Name string
			Age  int
		}
		dst := TestStruct{}
		src := TestStruct{Name: "Default", Age: 30}

		// Act
		fillFromDefaults(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src))

		// Assert
		assert.Equal(t, "Default", dst.Name)
		assert.Equal(t, 30, dst.Age)
	})

	t.Run("Do not overwrite non-zero values", func(t *testing.T) {
		// Arrange
		type TestStruct struct {
			Name string
			Age  int
		}
		dst := TestStruct{Name: "Original"}
		src := TestStruct{Name: "Default", Age: 30}

		// Act
		fillFromDefaults(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src))

		// Assert
		assert.Equal(t, "Original", dst.Name)
		assert.Equal(t, 30, dst.Age)
	})

	t.Run("Handle different field types", func(t *testing.T) {
		// Arrange
		type SrcStruct struct {
			Value int
		}
		type DstStruct struct {
			Value string
		}
		dst := DstStruct{}
		src := SrcStruct{Value: 42}

		// Act
		fillFromDefaults(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src))

		// Assert
		assert.Equal(t, "", dst.Value) // Should not change due to incompatible types
	})
}

func TestIsZeroValue(t *testing.T) {
	testCases := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"Integer zero", 0, true},
		{"Integer non-zero", 42, false},
		{"Empty string", "", true},
		{"Non-empty string", "hello", false},
		{"False boolean", false, true},
		{"True boolean", true, false},
		{"Zero time", time.Time{}, true},
		{"Non-zero time", time.Now(), false},
		{"Nil slice", []int(nil), true},
		{"Empty slice", []int{}, false},
		{"Nil map", map[string]int(nil), true},
		{"Empty map", map[string]int{}, false},
		{"Nil pointer", (*int)(nil), true},
		{"Non-nil pointer", new(int), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isZeroValue(reflect.ValueOf(tc.value))
			assert.Equal(t, tc.expected, result)
		})
	}
}
