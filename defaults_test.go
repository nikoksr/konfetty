//nolint:testpackage // We want to thoroughly test the underlying defaulting logic.
package konfetty

import (
	"reflect"
	"testing"
	"time"

	"github.com/shoenig/test/must"
)

func TestApplyDefaults(t *testing.T) {
	t.Parallel()

	t.Run("Basic Types", func(t *testing.T) {
		t.Parallel()
		testBasicTypes(t)
	})

	t.Run("Nested and Embedded Structs", func(t *testing.T) {
		t.Parallel()
		testNestedAndEmbeddedStructs(t)
	})

	t.Run("Slices and Pointers", func(t *testing.T) {
		t.Parallel()
		testSlicesAndPointers(t)
	})

	t.Run("Multiple Defaults", func(t *testing.T) {
		t.Parallel()
		testMultipleDefaults(t)
	})

	t.Run("Complex Nested Struct", func(t *testing.T) {
		t.Parallel()
		testComplexNestedStruct(t)
	})

	t.Run("Interface Slice", func(t *testing.T) {
		t.Parallel()
		testInterfaceSlice(t)
	})

	t.Run("Circular References", func(t *testing.T) {
		t.Parallel()
		testCircularReferences(t)
	})

	t.Run("Interface Fields", func(t *testing.T) {
		t.Parallel()
		testInterfaceFields(t)
	})

	t.Run("Maps", func(t *testing.T) {
		t.Parallel()
		testMaps(t)
	})

	t.Run("Embedded Structs", func(t *testing.T) {
		t.Parallel()
		testEmbeddedStructs(t)
	})

	t.Run("Slices of Interfaces", func(t *testing.T) {
		t.Parallel()
		testSlicesOfInterfaces(t)
	})
}

func TestApplyDefaultsErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   any
		defaults map[reflect.Type][]any
		expected string
	}{
		{
			name:     "Non-pointer input",
			config:   struct{ Name string }{},
			defaults: map[reflect.Type][]any{},
			expected: "must be a pointer",
		},
		{
			name:     "Nil input",
			config:   (*struct{ Name string })(nil),
			defaults: map[reflect.Type][]any{},
			expected: "cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := applyDefaults(tt.config, tt.defaults)
			must.Error(t, err)
			must.ErrorContains(t, err, tt.expected)
		})
	}
}

func testBasicTypes(t *testing.T) {
	type SimpleStruct struct {
		Name string
		Age  int
	}

	config := &SimpleStruct{Name: "", Age: 0}
	defaults := map[reflect.Type][]any{
		reflect.TypeFor[SimpleStruct](): {
			SimpleStruct{Name: "Default", Age: 30},
		},
	}

	err := applyDefaults(config, defaults)
	must.NoError(t, err)
	must.Eq(t, &SimpleStruct{Name: "Default", Age: 30}, config)
}

func testNestedAndEmbeddedStructs(t *testing.T) {
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

	tests := []struct {
		name     string
		config   any
		defaults map[reflect.Type][]any
		expected any
	}{
		{
			name: "Nested struct",
			config: &NestedStruct{
				Simple: SimpleStruct{},
				Value:  0,
			},
			defaults: map[reflect.Type][]any{
				reflect.TypeFor[NestedStruct](): {
					NestedStruct{Value: 3.14},
				},
				reflect.TypeFor[SimpleStruct](): {
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
			defaults: map[reflect.Type][]any{
				reflect.TypeFor[EmbeddedStruct](): {
					EmbeddedStruct{Extra: "ExtraDefault"},
				},
				reflect.TypeFor[SimpleStruct](): {
					SimpleStruct{Name: "Default", Age: 30},
				},
			},
			expected: &EmbeddedStruct{
				SimpleStruct: SimpleStruct{Name: "Default", Age: 30},
				Extra:        "ExtraDefault",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := applyDefaults(tt.config, tt.defaults)
			must.NoError(t, err)
			must.Eq(t, tt.expected, tt.config)
		})
	}
}

func testSlicesAndPointers(t *testing.T) {
	type SimpleStruct struct {
		Name string
		Age  int
	}

	type SliceStruct struct {
		Items []SimpleStruct
	}

	type PointerStruct struct {
		Ptr *SimpleStruct
	}

	tests := []struct {
		name     string
		config   any
		defaults map[reflect.Type][]any
		expected any
	}{
		{
			name: "Slice of structs",
			config: &SliceStruct{
				Items: []SimpleStruct{
					{Name: "Item1", Age: 0},
					{Name: "", Age: 25},
				},
			},
			defaults: map[reflect.Type][]any{
				reflect.TypeFor[SimpleStruct](): {
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
			defaults: map[reflect.Type][]any{
				reflect.TypeFor[SimpleStruct](): {
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
			defaults: map[reflect.Type][]any{
				reflect.TypeFor[SimpleStruct](): {
					SimpleStruct{Name: "Default", Age: 30},
				},
			},
			expected: &PointerStruct{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := applyDefaults(tt.config, tt.defaults)
			must.NoError(t, err)
			must.Eq(t, tt.expected, tt.config)
		})
	}
}

func testMultipleDefaults(t *testing.T) {
	type SimpleStruct struct {
		Name string
		Age  int
	}

	config := &SimpleStruct{Name: "", Age: 0}
	defaults := map[reflect.Type][]any{
		reflect.TypeFor[SimpleStruct](): {
			SimpleStruct{Name: "First", Age: 10},
			SimpleStruct{Name: "Second", Age: 20},
			SimpleStruct{Name: "Third", Age: 30},
		},
	}

	err := applyDefaults(config, defaults)
	must.NoError(t, err)
	must.Eq(t, &SimpleStruct{Name: "Third", Age: 30}, config)
}

func testComplexNestedStruct(t *testing.T) {
	type NestedStruct struct {
		Simple struct {
			Name string
			Age  int
		}
		Value float64
	}

	type ComplexStruct struct {
		Name     string
		Timeout  time.Duration
		Nested   NestedStruct
		Slices   []int
		MapField map[string]int
	}

	config := &ComplexStruct{
		Name:    "",
		Timeout: 0,
		Nested: NestedStruct{
			Simple: struct {
				Name string
				Age  int
			}{},
			Value: 0,
		},
		Slices:   nil,
		MapField: nil,
	}

	defaults := map[reflect.Type][]any{
		reflect.TypeFor[ComplexStruct](): {
			ComplexStruct{
				Name:    "DefaultName",
				Timeout: 5 * time.Second,
				Slices:  []int{1, 2, 3},
				MapField: map[string]int{
					"key": 42,
				},
			},
		},
		reflect.TypeFor[NestedStruct](): {
			NestedStruct{
				Simple: struct {
					Name string
					Age  int
				}{Name: "NestedDefault", Age: 25},
				Value: 3.14,
			},
		},
	}

	err := applyDefaults(config, defaults)
	must.NoError(t, err)

	expected := &ComplexStruct{
		Name:    "DefaultName",
		Timeout: 5 * time.Second,
		Nested: NestedStruct{
			Simple: struct {
				Name string
				Age  int
			}{Name: "NestedDefault", Age: 25},
			Value: 3.14,
		},
		Slices: []int{1, 2, 3},
		MapField: map[string]int{
			"key": 42,
		},
	}

	must.Eq(t, expected, config)
}

func testInterfaceSlice(t *testing.T) {
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
		Devices []any
	}

	config := &RoomConfig{
		Devices: []any{
			&LightDevice{BaseDevice: BaseDevice{Enabled: true}},
			&LightDevice{Brightness: 75},
			&ThermostatDevice{},
		},
	}

	defaults := map[reflect.Type][]any{
		reflect.TypeFor[BaseDevice]():  {BaseDevice{Enabled: false}},
		reflect.TypeFor[LightDevice](): {LightDevice{Brightness: 50}},
		reflect.TypeFor[ThermostatDevice](): {ThermostatDevice{
			BaseDevice:  BaseDevice{Enabled: true},
			Temperature: 20.0,
		}},
	}

	err := applyDefaults(config, defaults)
	must.NoError(t, err)

	devices := config.Devices
	must.Len(t, 3, devices)

	light1, ok := devices[0].(*LightDevice)
	must.True(t, ok)
	must.True(t, light1.Enabled)
	must.Eq(t, 50, light1.Brightness)

	light2, ok := devices[1].(*LightDevice)
	must.True(t, ok)
	must.False(t, light2.Enabled)
	must.Eq(t, 75, light2.Brightness)

	thermostat, ok := devices[2].(*ThermostatDevice)
	must.True(t, ok)
	must.True(t, thermostat.Enabled)
	must.InDelta(t, 20.0, thermostat.Temperature, 0.001)
}

func testCircularReferences(t *testing.T) {
	type Circular struct {
		Name string
		Next *Circular
	}

	config := &Circular{Name: "Start"}
	config.Next = &Circular{Name: "Middle"}
	config.Next.Next = &Circular{Name: ""}
	config.Next.Next.Next = config

	defaults := map[reflect.Type][]any{
		reflect.TypeFor[Circular](): {
			Circular{Name: "Default"},
		},
	}

	err := applyDefaults(config, defaults)
	must.Error(t, err)
	must.Eq(t, ErrCircularReference, err)

	must.Eq(t, "Start", config.Name)
	must.Eq(t, "Middle", config.Next.Name)
	must.Eq(t, "Default", config.Next.Next.Name)
	must.Eq(t, config, config.Next.Next.Next)
}

func testInterfaceFields(t *testing.T) {
	type ConcreteType struct {
		Value string
	}

	type InterfaceContainer struct {
		Data            any
		NestedContainer any
	}

	config := &InterfaceContainer{
		Data: &ConcreteType{},
		NestedContainer: &InterfaceContainer{
			Data: &ConcreteType{Value: "nested"},
		},
	}

	defaults := map[reflect.Type][]any{
		reflect.TypeFor[ConcreteType](): {
			ConcreteType{Value: "default"},
		},
	}

	err := applyDefaults(config, defaults)
	must.NoError(t, err)

	concData, ok := config.Data.(*ConcreteType)
	must.True(t, ok)
	must.Eq(t, "default", concData.Value)

	nestedContainer, ok := config.NestedContainer.(*InterfaceContainer)
	must.True(t, ok)
	nestedData, ok := nestedContainer.Data.(*ConcreteType)
	must.True(t, ok)
	must.Eq(t, "nested", nestedData.Value)
}

func testMaps(t *testing.T) {
	type SimpleStruct struct {
		Name string
		Age  int
	}

	type Config struct {
		StringMap map[string]string
		IntMap    map[string]int
		StructMap map[string]SimpleStruct
	}

	config := &Config{
		StringMap: map[string]string{"existing": "value"},
		IntMap:    nil,
		StructMap: map[string]SimpleStruct{
			"existing": {Name: "Alice", Age: 30},
			"empty":    {},
		},
	}

	defaults := map[reflect.Type][]any{
		reflect.TypeFor[Config](): {
			Config{
				StringMap: map[string]string{"default": "value"},
				IntMap:    map[string]int{"default": 42},
				StructMap: map[string]SimpleStruct{
					"default": {Name: "Default", Age: 25},
				},
			},
		},
		reflect.TypeFor[SimpleStruct](): {
			SimpleStruct{Name: "DefaultName", Age: 20},
		},
	}

	err := applyDefaults(config, defaults)
	must.NoError(t, err)

	must.Eq(t, "value", config.StringMap["existing"])
	must.Eq(t, "value", config.StringMap["default"])
	must.Eq(t, 42, config.IntMap["default"])
	must.Eq(t, SimpleStruct{Name: "Alice", Age: 30}, config.StructMap["existing"])
	must.Eq(t, SimpleStruct{Name: "DefaultName", Age: 20}, config.StructMap["empty"])
	must.Eq(t, SimpleStruct{Name: "Default", Age: 25}, config.StructMap["default"])
}

func testEmbeddedStructs(t *testing.T) {
	type unexportedEmbedded struct {
		privateField string
	}

	type EmbeddedLevel1 struct {
		Level1Field string
	}

	type EmbeddedLevel2 struct {
		EmbeddedLevel1

		Level2Field int
	}

	type Config struct {
		unexportedEmbedded //nolint:unused // Used for testing that unexported fields don't cause panics
		EmbeddedLevel2

		TopLevelField bool
	}

	config := &Config{}

	defaults := map[reflect.Type][]any{
		reflect.TypeFor[unexportedEmbedded](): {
			unexportedEmbedded{privateField: "default private"},
		},
		reflect.TypeFor[EmbeddedLevel1](): {
			EmbeddedLevel1{Level1Field: "default level 1"},
		},
		reflect.TypeFor[EmbeddedLevel2](): {
			EmbeddedLevel2{Level2Field: 42},
		},
		reflect.TypeFor[Config](): {
			Config{TopLevelField: true},
		},
	}

	err := applyDefaults(config, defaults)
	must.NoError(t, err)

	unexportedValue := reflect.ValueOf(config).Elem().FieldByName("unexportedEmbedded")
	privateFieldValue := unexportedValue.FieldByName("privateField")
	must.Eq(t, "", privateFieldValue.String())

	must.Eq(t, "default level 1", config.Level1Field)
	must.Eq(t, 42, config.Level2Field)
	must.True(t, config.TopLevelField)
}

type Animal interface {
	Sound() string
}

type Dog struct {
	Name string
}

func (d Dog) Sound() string {
	return "Woof!"
}

type Cat struct {
	Name string
}

func (c Cat) Sound() string {
	return "Meow!"
}

type Zoo struct {
	Animals []Animal
}

func testSlicesOfInterfaces(t *testing.T) {
	config := &Zoo{
		Animals: []Animal{
			&Dog{Name: "Buddy"},
			&Cat{},
		},
	}

	defaults := map[reflect.Type][]any{
		reflect.TypeFor[*Dog](): {
			&Dog{Name: "DefaultDog"},
		},
		reflect.TypeFor[*Cat](): {
			&Cat{Name: "DefaultCat"},
		},
	}

	err := applyDefaults(config, defaults)
	must.NoError(t, err)

	must.Len(t, 2, config.Animals)

	dog, ok := config.Animals[0].(*Dog)
	must.True(t, ok)
	must.Eq(t, "Buddy", dog.Name)

	cat, ok := config.Animals[1].(*Cat)
	must.True(t, ok)
	must.Eq(t, "DefaultCat", cat.Name)
}
