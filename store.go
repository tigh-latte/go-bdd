package bdd

// Store for storing.
type Store map[string]interface{}

// Store an item.
func (s Store) Store(key string, v interface{}) {
	switch a := v.(type) {
	case float32:
		if a == float32(int32(a)) {
			s.Store(key, int32(a))
		}
	case float64:
		if a == float64(int64(a)) {
			s.Store(key, int64(a))
		}
	default:
		s[key] = v
	}
}

// Get retrieve an item.
func (s Store) Get(key string) interface{} {
	return s[key]
}

// Has check an item exists.
func (s Store) Has(key string) bool {
	_, ok := s[key]
	return ok
}

// Int retrieve an int.
func (s Store) Int(key string) int {
	return s[key].(int)
}

// Int8 retrieve an int8.
func (s Store) Int8(key string) int8 {
	return s[key].(int8)
}

// Int16 retrieve an int16.
func (s Store) Int16(key string) int16 {
	return s[key].(int16)
}

// Int32 retrieve an int32.
func (s Store) Int32(key string) int32 {
	return s[key].(int32)
}

// Int64 retrieve an int64.
func (s Store) Int64(key string) int64 {
	return s[key].(int64)
}

// Uint retrieve an uint.
func (s Store) Uint(key string) uint {
	return s[key].(uint)
}

// Uint8 retrieve an uint8.
func (s Store) Uint8(key string) uint8 {
	return s[key].(uint8)
}

// Uint16 retrieve an uint16.
func (s Store) Uint16(key string) uint16 {
	return s[key].(uint16)
}

// Uint32 retrieve an uint32.
func (s Store) Uint32(key string) uint32 {
	return s[key].(uint32)
}

// Uint64 retrieve an uint64.
func (s Store) Uint64(key string) uint64 {
	return s[key].(uint64)
}

// String retrieve a string.
func (s Store) String(key string) string {
	return s[key].(string)
}

// Float32 retrieve a float32.
func (s Store) Float32(key string) float32 {
	return s[key].(float32)
}

// Float64 retrieve a float64.
func (s Store) Float64(key string) float64 {
	return s[key].(float64)
}

// Rune retrieve a rune.
func (s Store) Rune(key string) rune {
	return s[key].(rune)
}

// Byte retrieve a byte.
func (s Store) Byte(key string) byte {
	return s[key].(byte)
}

// ByteSlice retrieve a byte slice.
func (s Store) ByteSlice(key string) []byte {
	return s[key].([]byte)
}
