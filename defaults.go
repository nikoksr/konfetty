package konfetty

import (
	"errors"
	"reflect"
)

// maxDepth defines the maximum recursion depth for filling defaults.
// This prevents infinite recursion in case of circular dependencies.
const maxDepth = 100 // Adjust this value as needed

// DefaultProvider is an interface for types that can provide their own default values.
// Structs implementing this interface can define a Defaults method to return default values.
type DefaultProvider interface {
	Defaults() any
}

// fillDefaults recursively fills in default values for structs that implement DefaultProvider.
// It traverses the struct hierarchy and applies defaults to fields and nested structs.
func fillDefaults(v any, maxDepth int) error {
	return fillDefaultsRecursive(reflect.ValueOf(v), 0, maxDepth)
}

// fillDefaultsRecursive is the core function that recursively fills default values.
// It handles various types including pointers, structs, slices, and maps.
func fillDefaultsRecursive(v reflect.Value, depth, maxDepth int) error {
	// Check if maximum recursion depth is reached
	if depth > maxDepth {
		return errors.New("maximum recursion depth exceeded, possible circular dependency")
	}

	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			// Check if we can set the value before creating a new instance
			if !v.CanSet() {
				return nil // Skip if we can't set the value
			}
			// If the pointer is nil, create a new instance of the pointed-to type
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	// Only process struct types
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()

	// Iterate through all fields of the struct
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Handle embedded fields
		if fieldType.Anonymous {
			if field.Kind() == reflect.Ptr && field.IsNil() {
				// Check if we can set the field before creating a new instance
				if !field.CanSet() {
					continue
				}
				// If the embedded field is a nil pointer, create a new instance
				field.Set(reflect.New(field.Type().Elem()))
			}
			if err := fillDefaultsRecursive(field, depth+1, maxDepth); err != nil {
				return err
			}

			// Apply defaults for the embedded field if it implements DefaultProvider
			if field.CanAddr() && field.Addr().Type().Implements(reflect.TypeOf((*DefaultProvider)(nil)).Elem()) {
				defaulter := field.Addr().Interface().(DefaultProvider)
				defaults := reflect.ValueOf(defaulter.Defaults())
				if defaults.Kind() == reflect.Ptr {
					defaults = defaults.Elem()
				}
				fillFromDefaults(field, defaults)
			}

			continue
		}

		// Handle different field types
		switch field.Kind() {
		case reflect.Ptr:
			if field.IsNil() {
				// Check if we can set the field before creating a new instance
				if !field.CanSet() {
					continue
				}
				// If the field is a nil pointer, create a new instance
				field.Set(reflect.New(field.Type().Elem()))
			}
			if err := fillDefaultsRecursive(field.Elem(), depth+1, maxDepth); err != nil {
				return err
			}
		case reflect.Struct:
			if err := fillDefaultsRecursive(field, depth+1, maxDepth); err != nil {
				return err
			}
		case reflect.Slice:
			for j := 0; j < field.Len(); j++ {
				if err := fillDefaultsRecursive(field.Index(j), depth+1, maxDepth); err != nil {
					return err
				}
			}
		case reflect.Map:
			// Handle maps
			for _, key := range field.MapKeys() {
				value := field.MapIndex(key)
				if value.CanAddr() {
					if err := fillDefaultsRecursive(value, depth+1, maxDepth); err != nil {
						return err
					}
				}
			}
		}
	}

	// Apply defaults to the current struct if it implements DefaultProvider
	if v.CanAddr() && v.Addr().Type().Implements(reflect.TypeOf((*DefaultProvider)(nil)).Elem()) {
		defaulter := v.Addr().Interface().(DefaultProvider)
		defaults := reflect.ValueOf(defaulter.Defaults())
		if defaults.Kind() == reflect.Ptr {
			defaults = defaults.Elem()
		}
		fillFromDefaults(v, defaults)
	}

	return nil
}

// fillFromDefaults copies default values from the source to the destination struct.
// It only sets values for fields that are zero-valued in the destination.
func fillFromDefaults(dst, src reflect.Value) {
	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		srcFieldName := src.Type().Field(i).Name

		// Check if the destination has this field
		dstField := dst.FieldByName(srcFieldName)
		if !dstField.IsValid() || !dstField.CanSet() {
			continue // Skip fields that don't exist in the destination or can't be set
		}

		// Check if the types are compatible
		if !srcField.Type().AssignableTo(dstField.Type()) {
			continue // Skip if types are not compatible
		}

		// Only set the default value if the destination field is zero-valued
		if isZeroValue(dstField) {
			dstField.Set(srcField)
		}
	}
}

// isZeroValue checks if the given reflect.Value is the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	zero := reflect.Zero(v.Type()).Interface()
	return reflect.DeepEqual(v.Interface(), zero)
}
