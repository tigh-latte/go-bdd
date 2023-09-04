package bdd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

// toInt converts v to int
func toInt(v any) int {
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

func toHostname(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return s, fmt.Errorf("failed to parse url '%s': %w", s, err)
	}
	return u.Hostname(), nil
}

func assertFuture(s string) (string, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s, fmt.Errorf("value '%s' is not a timestamp", err)
	}
	if time.Now().UTC().After(t) {
		return s, fmt.Errorf("time %s not in future", t.String())
	}
	return s, nil
}

func assertNotEmpty(s any) (any, error) {
	if s == nil {
		return s, errors.New("value is nil")
	}

	val := reflect.ValueOf(s)

	for val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}

	if val.IsZero() {
		return s, errors.New("value is empty")
	}

	return s, nil
}

func assertJsonString(s string) (string, error) {
	if s == "" {
		return s, fmt.Errorf("empty: '%s'", s)
	}

	var a any

	// Unmarshal to test validity.
	if err := json.Unmarshal([]byte(s), &a); err != nil {
		return s, fmt.Errorf("'%s' not a json string: %w", s, err)
	}

	// Remarshal to re-escape.
	bb, err := json.Marshal(s)
	if err != nil {
		return s, fmt.Errorf("'%s' could not be escaped: %w", s, err)
	}
	return string(bb[1 : len(bb)-1]), nil
}
