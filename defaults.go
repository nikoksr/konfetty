package konfetty

import (
	"reflect"
)

// applyDefaults is the entry point for applying default values to the loaded config.
func applyDefaults(config any, defaults map[reflect.Type][]any) error {
	v := reflect.ValueOf(config)

	if v.Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	if v.IsNil() {
		return ErrNilConfig
	}

	visited := make(map[uintptr]bool)

	return applyDefaultsRecursive(v.Elem(), defaults, visited)
}

// applyDefaultsRecursive contains the core logic for applying default values to the config.
func applyDefaultsRecursive(v reflect.Value, defaults map[reflect.Type][]any, visited map[uintptr]bool) error {
	if err := checkCircularReference(v, visited); err != nil {
		return err
	}

	t := v.Type()

	if err := applyTypeDefaults(v, defaults[t]); err != nil {
		return err
	}

	//nolint:exhaustive // Only handling relevant types for config structures; other types don't need special processing
	switch t.Kind() {
	case reflect.Struct:
		return handleStruct(v, defaults, visited)
	case reflect.Slice:
		return handleSlice(v, defaults, visited)
	case reflect.Map:
		return handleMap(v, defaults, visited)
	case reflect.Ptr:
		return handlePointer(v, defaults, visited)
	case reflect.Interface:
		return handleInterface(v, defaults, visited)
	default:
		// Other kinds don't need special handling
	}

	return nil
}

func checkCircularReference(v reflect.Value, visited map[uintptr]bool) error {
	if v.Kind() == reflect.Ptr {
		ptr := v.Pointer()
		if visited[ptr] {
			return ErrCircularReference
		}
		visited[ptr] = true
	}

	return nil
}

func applyTypeDefaults(v reflect.Value, typeDefaults []any) error {
	for i := len(typeDefaults) - 1; i >= 0; i-- {
		if err := mergeDefault(v, reflect.ValueOf(typeDefaults[i])); err != nil {
			return err
		}
	}

	return nil
}

func handleStruct(v reflect.Value, defaults map[reflect.Type][]any, visited map[uintptr]bool) error {
	for i := range v.NumField() {
		if err := applyDefaultsRecursive(v.Field(i), defaults, visited); err != nil {
			return err
		}
	}

	return nil
}

func handleSlice(v reflect.Value, defaults map[reflect.Type][]any, visited map[uintptr]bool) error {
	for i := range v.Len() {
		elem := v.Index(i)
		if elem.Kind() == reflect.Interface && !elem.IsNil() {
			elem = elem.Elem()
		}

		newElem := reflect.New(elem.Type()).Elem()
		newElem.Set(elem)
		if err := applyDefaultsRecursive(newElem, defaults, visited); err != nil {
			return err
		}

		v.Index(i).Set(newElem)
	}

	return nil
}

func handleMap(v reflect.Value, defaults map[reflect.Type][]any, visited map[uintptr]bool) error {
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	for _, key := range v.MapKeys() {
		elem := v.MapIndex(key)
		if elem.Kind() == reflect.Interface && !elem.IsNil() {
			elem = elem.Elem()
		}

		newElem := reflect.New(elem.Type()).Elem()
		newElem.Set(elem)
		if err := applyDefaultsRecursive(newElem, defaults, visited); err != nil {
			return err
		}

		v.SetMapIndex(key, newElem)
	}

	return applyMapDefaults(v, defaults[v.Type()])
}

func applyMapDefaults(v reflect.Value, defaultValues []any) error {
	for _, dv := range defaultValues {
		defaultMap := reflect.ValueOf(dv)
		for _, key := range defaultMap.MapKeys() {
			if !v.MapIndex(key).IsValid() {
				v.SetMapIndex(key, defaultMap.MapIndex(key))
			}
		}
	}

	return nil
}

func handlePointer(v reflect.Value, defaults map[reflect.Type][]any, visited map[uintptr]bool) error {
	if !v.IsNil() {
		return applyDefaultsRecursive(v.Elem(), defaults, visited)
	}

	return nil
}

func handleInterface(v reflect.Value, defaults map[reflect.Type][]any, visited map[uintptr]bool) error {
	if !v.IsNil() {
		return applyDefaultsRecursive(v.Elem(), defaults, visited)
	}

	return nil
}

// mergeDefault applies default values from src to dst, but only for zero-value fields in dst.
func mergeDefault(dst, src reflect.Value) error {
	dst = dereference(dst)
	src = dereference(src)

	if src.Kind() != reflect.Struct || dst.Kind() != reflect.Struct {
		return nil
	}

	for i := range src.NumField() {
		if err := mergeField(dst.Field(i), src.Field(i), dst.Type().Field(i)); err != nil {
			return err
		}
	}

	return nil
}

func mergeField(dst, src reflect.Value, structField reflect.StructField) error {
	if !structField.IsExported() {
		return nil
	}

	if dst.IsZero() {
		return setField(dst, src)
	}

	//nolint:exhaustive // Only merging struct, ptr, and map fields; other types are handled by the default zero-value
	//                  // check
	switch src.Kind() {
	case reflect.Struct:
		return mergeDefault(dst, src)
	case reflect.Ptr:
		return mergePtrField(dst, src)
	case reflect.Map:
		return mergeMapField(dst, src)
	default:
		// Other kinds don't need special handling
	}

	return nil
}

func mergePtrField(dst, src reflect.Value) error {
	if src.IsNil() || src.Elem().Kind() != reflect.Struct {
		return nil
	}

	if dst.IsNil() {
		dst.Set(reflect.New(src.Elem().Type()))
	}

	return mergeDefault(dst.Elem(), src.Elem())
}

func mergeMapField(dst, src reflect.Value) error {
	if dst.IsNil() {
		return nil
	}

	for _, key := range src.MapKeys() {
		if !dst.MapIndex(key).IsValid() {
			dst.SetMapIndex(key, src.MapIndex(key))
		}
	}

	return nil
}

func dereference(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}

	return v
}

func setField(dst, src reflect.Value) error {
	if src.Kind() == reflect.Map && dst.IsNil() {
		dst.Set(reflect.MakeMap(src.Type()))
		for _, key := range src.MapKeys() {
			dst.SetMapIndex(key, src.MapIndex(key))
		}

		return nil
	}

	dst.Set(src)

	return nil
}
