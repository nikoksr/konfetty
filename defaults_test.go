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

func (b BaseCheck) Defaults() any {
	return BaseCheck{
		// This should take the lowest priority, filling the missing fields
		Interval: 30 * time.Second,
		Timeout:  5 * time.Second,
	}
}

type PingCheck struct {
	*BaseCheck
	Host string
}

func (b PingCheck) Defaults() any {
	return PingCheck{
		BaseCheck: &BaseCheck{
			Interval: 1 * time.Second, // This should take the second highest priority, filling the missing fields
		},
	}
}

type HTTPCheck struct {
	*BaseCheck
	URL    string
	Method string
}

func (h HTTPCheck) Defaults() any {
	return HTTPCheck{
		Method: "GET",
	}
}

type Checks struct {
	Ping []PingCheck
	HTTP []HTTPCheck
}

type Profile struct {
	Name   string
	Checks Checks
}

func TestProfileWithChecks(t *testing.T) {
	t.Run("Fill Profile with default values", func(t *testing.T) {
		profile := &Profile{
			Checks: Checks{
				Ping: []PingCheck{
					{
						BaseCheck: &BaseCheck{Name: "Custom Ping"}, // This should take the highest priority, but, the BaseCheck has not all fields filled
						Host:      "example.com",
					},
					{
						Host: "google.com",
					},
				},
			},
		}

		fillDefaults(profile)

		require.Len(t, profile.Checks.Ping, 2)

		assert.Equal(t, "Custom Ping", profile.Checks.Ping[0].Name)
		assert.Equal(t, 1*time.Second, profile.Checks.Ping[0].Interval)
		assert.Equal(t, 5*time.Second, profile.Checks.Ping[0].Timeout)
		assert.Equal(t, "example.com", profile.Checks.Ping[0].Host)

		assert.Equal(t, "", profile.Checks.Ping[1].Name)
		assert.Equal(t, 1*time.Second, profile.Checks.Ping[1].Interval) // This is expected to be 1 second, due to the order of priority. If, say, PingCheck.Defaults would not set Interval, it would be 30 seconds due to BaseCheck.Defaults
		assert.Equal(t, 5*time.Second, profile.Checks.Ping[1].Timeout)
		assert.Equal(t, "google.com", profile.Checks.Ping[1].Host)
	})
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
						BaseCheck: &BaseCheck{},
						URL:       "https://api.google.com",
					},
				},
			},
		}

		// Regular tests continue...

		fillDefaults(profile)

		require.Len(t, profile.Checks.Ping, 2)
		assert.Equal(t, "Custom Ping", profile.Checks.Ping[0].Name)
		assert.Equal(t, 1*time.Second, profile.Checks.Ping[0].Interval)
		assert.Equal(t, 5*time.Second, profile.Checks.Ping[0].Timeout)
		assert.Equal(t, "example.com", profile.Checks.Ping[0].Host)

		assert.Equal(t, "", profile.Checks.Ping[1].Name)
		assert.Equal(t, 1*time.Second, profile.Checks.Ping[1].Interval)
		assert.Equal(t, 5*time.Second, profile.Checks.Ping[1].Timeout)
		assert.Equal(t, "google.com", profile.Checks.Ping[1].Host)

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
		outer := &OuterConfig{}
		fillDefaults(outer)
		assert.Equal(t, 42, outer.Nested.Value)
	})

	t.Run("Handle nil pointers", func(t *testing.T) {
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

		fillDefaults(profile)
		require.NotNil(t, profile.Checks.Ping[0].BaseCheck)
		require.NotNil(t, profile.Checks.HTTP[0].BaseCheck)
	})

	t.Run("Non-struct type", func(t *testing.T) {
		value := 42
		fillDefaults(&value)
		assert.Equal(t, 42, value)
	})

	t.Run("Empty struct", func(t *testing.T) {
		emptyStruct := struct{}{}
		fillDefaults(&emptyStruct)
		assert.Equal(t, struct{}{}, emptyStruct)
	})

	t.Run("Struct with unexported fields", func(t *testing.T) {
		type unexportedStruct struct {
			name string
			age  int
		}
		s := unexportedStruct{name: "original", age: 25}
		fillDefaults(&s)
		assert.Equal(t, unexportedStruct{name: "original", age: 25}, s)
	})

	t.Run("Struct with interface field", func(t *testing.T) {
		type interfaceStruct struct {
			Data interface{}
		}
		s := interfaceStruct{}
		fillDefaults(&s)
		assert.Nil(t, s.Data)
	})

	t.Run("Handle embedded fields with nil pointers", func(t *testing.T) {
		type Embedded struct {
			Value string
		}
		type TestStruct struct {
			*Embedded
		}
		value := TestStruct{}
		fillDefaults(&value)
		require.Nil(t, value.Embedded)
	})

	t.Run("Handle embedded fields implementing DefaultProvider", func(t *testing.T) {
		value := TestStruct{}
		fillDefaults(&value)
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
		fillDefaults(&value)
		require.Nil(t, value.Ptr)
	})

	t.Run("Handle slice of structs", func(t *testing.T) {
		type Item struct {
			Value string
		}
		type TestStruct struct {
			Items []Item
		}
		value := TestStruct{Items: []Item{{}, {}}}
		fillDefaults(&value)
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
		fillDefaults(&value)
		assert.Len(t, value.Items, 2)
	})

	t.Run("Panic in user-defined Defaults method", func(t *testing.T) {
		value := ErrorStruct{}
		assert.Panics(t, func() {
			fillDefaults(&value)
		})
	})

	t.Run("Handle nested pointer fields", func(t *testing.T) {
		type NestedPtr struct {
			Value *string
		}
		type TestStruct struct {
			Nested *NestedPtr
		}
		value := TestStruct{}
		fillDefaults(&value)
		require.Nil(t, value.Nested)
	})

	t.Run("Handle pointer to pointer", func(t *testing.T) {
		type TestStruct struct {
			Value *string
		}
		value := &TestStruct{}
		ptrToPtr := &value
		fillDefaults(&ptrToPtr)
		require.NotNil(t, *ptrToPtr)
		require.Nil(t, (*ptrToPtr).Value)
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
