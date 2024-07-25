package konfetty

import "reflect"

// DefaultProvider is an interface for types that can provide their own default values
type DefaultProvider interface {
	Defaults() any
}

// fillDefaults recursively fills in default values for structs that implement DefaultProvider
func fillDefaults(v any) error {
	return fillDefaultsRecursive(reflect.ValueOf(v))
}

func fillDefaultsRecursive(v reflect.Value) error {
	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			// If the pointer is nil, create a new instance of the pointed-to type
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()

	// Iterate through all fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Handle embedded fields
		if fieldType.Anonymous {
			if field.Kind() == reflect.Ptr && field.IsNil() {
				// If the embedded field is a nil pointer, create a new instance
				field.Set(reflect.New(field.Type().Elem()))
			}
			if err := fillDefaultsRecursive(field); err != nil {
				return err
			}

			// Apply defaults for the embedded field if it implements DefaultProvider
			if defaulter, ok := field.Addr().Interface().(DefaultProvider); ok {
				defaults := reflect.ValueOf(defaulter.Defaults())
				if defaults.Kind() == reflect.Ptr {
					defaults = defaults.Elem()
				}
				fillFromDefaults(field, defaults)
			}

			continue
		}

		switch field.Kind() {
		case reflect.Ptr:
			if field.IsNil() {
				// If the field is a nil pointer, create a new instance
				field.Set(reflect.New(field.Type().Elem()))
			}
			if err := fillDefaultsRecursive(field.Elem()); err != nil {
				return err
			}
		case reflect.Struct:
			if err := fillDefaultsRecursive(field); err != nil {
				return err
			}
		case reflect.Slice:
			for j := 0; j < field.Len(); j++ {
				if err := fillDefaultsRecursive(field.Index(j)); err != nil {
					return err
				}
			}
		}
	}

	// Apply defaults to the current struct if it implements DefaultProvider
	if defaulter, ok := v.Addr().Interface().(DefaultProvider); ok {
		defaults := reflect.ValueOf(defaulter.Defaults())
		if defaults.Kind() == reflect.Ptr {
			defaults = defaults.Elem()
		}
		fillFromDefaults(v, defaults)
	}

	return nil
}

func fillFromDefaults(dst, src reflect.Value) {
	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		srcFieldName := src.Type().Field(i).Name

		// Check if the destination has this field
		dstField := dst.FieldByName(srcFieldName)
		if !dstField.IsValid() {
			continue // Skip fields that don't exist in the destination
		}

		if dstField.CanSet() && isZeroValue(dstField) {
			// Only set the value if it's settable and currently zero
			dstField.Set(srcField)
		}
	}
}

func isZeroValue(v reflect.Value) bool {
	zero := reflect.Zero(v.Type()).Interface()
	return reflect.DeepEqual(v.Interface(), zero)
}
