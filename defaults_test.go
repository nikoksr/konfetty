//nolint:testpackage // Access to unexported functions.
package konfetty

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyDefaults(t *testing.T) {
	t.Parallel()

	type SimpleStruct struct {
		Name string
		Age  int
	}

	type NestedStruct struct {
		Simple SimpleStruct
		Value  float64
	}

	type EmbeddedStruct struct {
		SimpleStruct
		Extra string
	}

	type SliceStruct struct {
		Items []SimpleStruct
	}

	type PointerStruct struct {
		Ptr *SimpleStruct
	}

	tests := []struct {
		name     string
		config   interface{}
		defaults map[reflect.Type][]interface{}
		expected interface{}
	}{
		{
			name: "Simple struct with basic types",
			config: &SimpleStruct{
				Name: "",
				Age:  0,
			},
			defaults: map[reflect.Type][]interface{}{
				reflect.TypeOf(SimpleStruct{}): {
					SimpleStruct{Name: "Default", Age: 30},
				},
			},
			expected: &SimpleStruct{
				Name: "Default",
				Age:  30,
			},
		},
		{
			name: "Nested struct",
			config: &NestedStruct{
				Simple: SimpleStruct{},
				Value:  0,
			},
			defaults: map[reflect.Type][]interface{}{
				reflect.TypeOf(NestedStruct{}): {
					NestedStruct{Value: 3.14},
				},
				reflect.TypeOf(SimpleStruct{}): {
					SimpleStruct{Name: "Default", Age: 30},
				},
			},
			expected: &NestedStruct{
				Simple: SimpleStruct{Name: "Default", Age: 30},
				Value:  3.14,
			},
		},
		{
			name: "Embedded struct",
			config: &EmbeddedStruct{
				SimpleStruct: SimpleStruct{},
				Extra:        "",
			},
			defaults: map[reflect.Type][]interface{}{
				reflect.TypeOf(EmbeddedStruct{}): {
					EmbeddedStruct{Extra: "ExtraDefault"},
				},
				reflect.TypeOf(SimpleStruct{}): {
					SimpleStruct{Name: "Default", Age: 30},
				},
			},
			expected: &EmbeddedStruct{
				SimpleStruct: SimpleStruct{Name: "Default", Age: 30},
				Extra:        "ExtraDefault",
			},
		},
		{
			name: "Slice of structs",
			config: &SliceStruct{
				Items: []SimpleStruct{
					{Name: "Item1", Age: 0},
					{Name: "", Age: 25},
				},
			},
			defaults: map[reflect.Type][]interface{}{
				reflect.TypeOf(SimpleStruct{}): {
					SimpleStruct{Name: "Default", Age: 30},
				},
			},
			expected: &SliceStruct{
				Items: []SimpleStruct{
					{Name: "Item1", Age: 30},
					{Name: "Default", Age: 25},
				},
			},
		},
		{
			name: "Pointer to struct",
			config: &PointerStruct{
				Ptr: &SimpleStruct{},
			},
			defaults: map[reflect.Type][]interface{}{
				reflect.TypeOf(SimpleStruct{}): {
					SimpleStruct{Name: "Default", Age: 30},
				},
			},
			expected: &PointerStruct{
				Ptr: &SimpleStruct{Name: "Default", Age: 30},
			},
		},
		{
			name:   "Nil pointer",
			config: &PointerStruct{},
			defaults: map[reflect.Type][]interface{}{
				reflect.TypeOf(SimpleStruct{}): {
					SimpleStruct{Name: "Default", Age: 30},
				},
			},
			expected: &PointerStruct{},
		},
		{
			name: "Multiple defaults",
			config: &SimpleStruct{
				Name: "",
				Age:  0,
			},
			defaults: map[reflect.Type][]interface{}{
				reflect.TypeOf(SimpleStruct{}): {
					SimpleStruct{Name: "First", Age: 10},
					SimpleStruct{Name: "Second", Age: 20},
					SimpleStruct{Name: "Third", Age: 30},
				},
			},
			expected: &SimpleStruct{
				Name: "Third",
				Age:  30,
			},
		},
		{
			name: "Complex nested struct with time.Duration",
			config: &struct {
				Name     string
				Timeout  time.Duration
				Nested   NestedStruct
				Slices   []int
				MapField map[string]int
			}{
				Name:    "",
				Timeout: 0,
				Nested: NestedStruct{
					Simple: SimpleStruct{},
					Value:  0,
				},
				Slices:   nil,
				MapField: nil,
			},
			defaults: map[reflect.Type][]interface{}{
				reflect.TypeOf(struct {
					Name     string
					Timeout  time.Duration
					Nested   NestedStruct
					Slices   []int
					MapField map[string]int
				}{}): {
					struct {
						Name     string
						Timeout  time.Duration
						Nested   NestedStruct
						Slices   []int
						MapField map[string]int
					}{
						Name:    "DefaultName",
						Timeout: 5 * time.Second,
						Slices:  []int{1, 2, 3},
						MapField: map[string]int{
							"key": 42,
						},
					},
				},
				reflect.TypeOf(NestedStruct{}): {
					NestedStruct{
						Simple: SimpleStruct{Name: "NestedDefault", Age: 25},
						Value:  3.14,
					},
				},
			},
			expected: &struct {
				Name     string
				Timeout  time.Duration
				Nested   NestedStruct
				Slices   []int
				MapField map[string]int
			}{
				Name:    "DefaultName",
				Timeout: 5 * time.Second,
				Nested: NestedStruct{
					Simple: SimpleStruct{Name: "NestedDefault", Age: 25},
					Value:  3.14,
				},
				Slices: []int{1, 2, 3},
				MapField: map[string]int{
					"key": 42,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := applyDefaults(tt.config, tt.defaults)
			require.NoError(t, err, "applyDefaults should not return an error")
			assert.Equal(t, tt.expected, tt.config, "applyDefaults result should match expected")
		})
	}
}

func TestApplyDefaultsErrors(t *testing.T) {
	t.Parallel()

	t.Run("Non-pointer input", func(t *testing.T) {
		t.Parallel()

		config := struct{ Name string }{}
		defaults := map[reflect.Type][]interface{}{}

		err := applyDefaults(config, defaults)
		require.Error(t, err, "applyDefaults should return an error for non-pointer input")
		assert.Contains(
			t,
			err.Error(),
			"must be a pointer",
			"Error message should indicate that input must be a pointer",
		)
	})

	t.Run("Nil input", func(t *testing.T) {
		t.Parallel()

		var config *struct{ Name string }
		defaults := map[reflect.Type][]interface{}{}

		err := applyDefaults(config, defaults)
		require.Error(t, err, "applyDefaults should return an error for nil input")
		assert.Contains(t, err.Error(), "cannot be nil", "Error message should indicate that input cannot be nil")
	})
}

func TestApplyDefaultsWithInterfaceSlice(t *testing.T) {
	t.Parallel()

	type BaseDevice struct {
		Enabled bool
	}

	type LightDevice struct {
		BaseDevice
		Brightness int
	}

	type ThermostatDevice struct {
		BaseDevice
		Temperature float64
	}

	type RoomConfig struct {
		Devices []interface{}
	}

	config := &RoomConfig{
		Devices: []interface{}{
			&LightDevice{BaseDevice: BaseDevice{Enabled: true}},
			&LightDevice{Brightness: 75},
			&ThermostatDevice{},
		},
	}

	defaults := map[reflect.Type][]interface{}{
		reflect.TypeOf(BaseDevice{}):  {BaseDevice{Enabled: false}},
		reflect.TypeOf(LightDevice{}): {LightDevice{Brightness: 50}},
		reflect.TypeOf(ThermostatDevice{}): {ThermostatDevice{
			BaseDevice:  BaseDevice{Enabled: true},
			Temperature: 20.0,
		}},
	}

	err := applyDefaults(config, defaults)
	require.NoError(t, err)

	devices := config.Devices
	require.Len(t, devices, 3)

	light1, ok := devices[0].(*LightDevice)
	require.True(t, ok)
	assert.True(t, light1.Enabled)
	assert.Equal(t, 50, light1.Brightness)

	light2, ok := devices[1].(*LightDevice)
	require.True(t, ok)
	assert.False(t, light2.Enabled)
	assert.Equal(t, 75, light2.Brightness)

	thermostat, ok := devices[2].(*ThermostatDevice)
	require.True(t, ok)
	assert.True(t, thermostat.Enabled)
	assert.InEpsilon(t, 20.0, thermostat.Temperature, 0.0)
}
