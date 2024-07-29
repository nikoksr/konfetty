package konfetty

import (
	"errors"
	"reflect"
)

// applyDefaults is the entry point for applying default values to the loaded config. It ensures the config is a non-nil
// pointer to a struct before initiating the recursive process.
func applyDefaults(config any, defaults map[reflect.Type][]any) error {
	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Ptr {
		return errors.New("config must be a pointer to a struct")
	}
	if v.IsNil() {
		return errors.New("config cannot be nil")
	}

	return applyDefaultsRecursive(v.Elem(), defaults)
}

// applyDefaultsRecursive contains the core logic for applying default values to the config. It recursively traverses
// the config structure, applying defaults where appropriate.
func applyDefaultsRecursive(v reflect.Value, defaults map[reflect.Type][]any) error {
	t := v.Type()

	// Apply defaults for this specific type, if any exist. We iterate in reverse order to respect the priority of later
	// defaults.
	if typeDefaults, ok := defaults[t]; ok {
		for i := len(typeDefaults) - 1; i >= 0; i-- {
			mergeDefault(v, reflect.ValueOf(typeDefaults[i]))
		}
	}

	// Recursive case: struct fields
	// We dive into each field of a struct, applying defaults recursively.
	if t.Kind() == reflect.Struct {
		for i := range v.NumField() {
			if err := applyDefaultsRecursive(v.Field(i), defaults); err != nil {
				return err
			}
		}
	}

	// Recursive case: slice elements
	// We apply defaults to each element of a slice.
	if t.Kind() == reflect.Slice {
		for i := range v.Len() {
			if err := applyDefaultsRecursive(v.Index(i), defaults); err != nil {
				return err
			}
		}
	}

	// Recursive case: pointer
	// If we encounter a non-nil pointer, we dereference and continue.
	if t.Kind() == reflect.Ptr && !v.IsNil() {
		return applyDefaultsRecursive(v.Elem(), defaults)
	}

	return nil
}

// mergeDefault applies default values from src to dst, but only for zero-value fields in dst. This function respects
// the existing values in the config while filling in missing ones.
func mergeDefault(dst, src reflect.Value) {
	for i := range src.NumField() {
		srcField := src.Field(i)
		dstField := dst.Field(i)

		// We only apply the default if the destination field is zero-value. This preserves any explicitly set values in
		// the config.
		if dstField.IsZero() {
			dstField.Set(srcField)
		}

		// Recursive case: struct fields
		// We dive deeper into struct fields to apply nested defaults.
		if srcField.Kind() == reflect.Struct {
			mergeDefault(dstField, srcField)
		}

		// Recursive case: pointer to struct
		// We handle the case where default values include pointers to structs.
		if srcField.Kind() == reflect.Ptr && !srcField.IsNil() && srcField.Elem().Kind() == reflect.Struct {
			if dstField.IsNil() {
				// If the destination field is nil, we create a new struct to hold the defaults.
				dstField.Set(reflect.New(srcField.Elem().Type()))
			}
			mergeDefault(dstField.Elem(), srcField.Elem())
		}
	}
}
