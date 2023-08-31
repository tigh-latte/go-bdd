package bdd

import (
	"reflect"
	"strconv"
)

// intify converts v to int
func intify(v any) int {
	if v == nil {
		return 0
	}

	val := reflect.ValueOf(v)
	// Keep de-referencing until base value is reached
	for val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}

	switch val.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		if val.Len() > 0 {
			return 1
		}
		return 0
	case reflect.String:
		if val.Len() == 0 {
			return 0
		}

		i, err := strconv.Atoi(val.String())
		if err != nil {
			return 0
		}

		return i
	case reflect.Bool:
		if val.Bool() {
			return 1
		}

		return 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int(val.Uint())
	case reflect.Float32, reflect.Float64:
		return int(val.Float())
	default:
		return 0
	}
}
