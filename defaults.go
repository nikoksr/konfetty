package konfetty

import (
	"reflect"
)

// DefaultProvider is an interface for types that can provide their own default values.
// Structs implementing this interface can define a Defaults method to return default values.
type DefaultProvider interface {
	Defaults() any
}

// fillDefaults recursively fills in default values for structs that implement DefaultProvider.
// It traverses the struct hierarchy and applies defaults to fields and nested structs.
// FillDefaults recursively fills default values in the given value
func fillDefaults(v interface{}) {
	fillDefaultsRecursive(reflect.ValueOf(v), make(map[uintptr]bool))
}

func fillDefaultsRecursive(v reflect.Value, visited map[uintptr]bool) {
	if !v.IsValid() {
		return
	}

	// Handle pointers
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			if v.Type().Implements(reflect.TypeOf((*DefaultProvider)(nil)).Elem()) {
				v.Set(reflect.New(v.Type().Elem()))
			} else {
				return
			}
		}
		v = v.Elem()
	}

	if !v.CanSet() {
		return
	}

	// Handle structs
	if v.Kind() == reflect.Struct {
		// Check for circular references
		if visited[v.UnsafeAddr()] {
			return
		}
		visited[v.UnsafeAddr()] = true

		// Apply defaults from lowest to highest priority
		applyDefaultsRecursive(v)

		// Recurse on all fields
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fillDefaultsRecursive(field, visited)
		}

		return
	}

	// Handle slices and arrays
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			fillDefaultsRecursive(v.Index(i), visited)
		}
		return
	}

	// Handle maps
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			fillDefaultsRecursive(v.MapIndex(key), visited)
		}
		return
	}
}

func applyDefaultsRecursive(v reflect.Value) {
	// First, apply defaults to embedded fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Struct {
			applyDefaultsRecursive(field)
		}
	}

	// Then apply struct-specific defaults
	if v.Addr().Type().Implements(reflect.TypeOf((*DefaultProvider)(nil)).Elem()) {
		defaults := v.Addr().Interface().(DefaultProvider).Defaults()
		applyDefaults(v, reflect.ValueOf(defaults))
	}
}

func applyDefaults(target, defaults reflect.Value) {
	if target.Kind() != reflect.Struct || defaults.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < defaults.NumField(); i++ {
		defaultField := defaults.Field(i)
		targetField := target.Field(i)

		if targetField.CanSet() {
			if isZeroValue(targetField) {
				// If the target field is zero, always apply the default
				if defaultField.Type() == targetField.Type() {
					targetField.Set(defaultField)
				}
			} else if targetField.Kind() == reflect.Ptr && defaultField.Kind() == reflect.Ptr {
				// For non-zero pointers, recurse into the struct
				if !targetField.IsNil() && !defaultField.IsNil() {
					applyDefaults(targetField.Elem(), defaultField.Elem())
				}
			}
		}
	}
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == complex(0, 0)
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZeroValue(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	case reflect.String:
		return v.Len() == 0
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
