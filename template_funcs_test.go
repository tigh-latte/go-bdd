package bdd

import "testing"

func Test_intify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  interface{}
		want int
	}{
		{
			name: "should convert string to int",
			arg:  "2",
			want: 2,
		},
		{
			name: "should correctly return int",
			arg:  2,
			want: 2,
		},
		{
			name: "should convert int8 to int",
			arg:  int8(5),
			want: 5,
		},
		{
			name: "should convert int16 to int",
			arg:  int16(5),
			want: 5,
		},
		{
			name: "should convert int32 to int",
			arg:  int32(5),
			want: 5,
		},
		{
			name: "should convert int64 to int",
			arg:  int64(5),
			want: 5,
		},
		{
			name: "should convert uint to int",
			arg:  uint(5),
			want: 5,
		},
		{
			name: "should convert uint8 to int",
			arg:  uint8(5),
			want: 5,
		},
		{
			name: "should convert uint16 to int",
			arg:  uint16(5),
			want: 5,
		},
		{
			name: "should convert uint32 to int",
			arg:  uint32(5),
			want: 5,
		},
		{
			name: "should convert uint64 to int",
			arg:  uint64(5),
			want: 5,
		},
		{
			name: "should convert uintptr to int",
			arg:  uintptr(5),
			want: 5,
		},
		{
			name: "should convert float32 to int",
			arg:  float32(3.11),
			want: 3,
		},
		{
			name: "should convert float64 to int",
			arg:  5.669,
			want: 5,
		},
		{
			name: "should convert map with value to 1",
			arg:  map[string]int{"hello": 3, "Goodbye": 1},
			want: 1,
		},
		{
			name: "should convert map with no values to 0",
			arg:  map[string]int{},
			want: 0,
		},
		{
			name: "should convert slice with values to 1",
			arg:  []int{1, 2, 3},
			want: 1,
		},
		{
			name: "should convert slice with no values to 0",
			arg:  []int{},
			want: 0,
		},
		{
			name: "should convert array with values to 1",
			arg:  [3]int{1, 2, 3},
			want: 1,
		},
		{
			name: "should convert array with no values to 0",
			arg:  [0]int{},
			want: 0,
		},
		{
			name: "should return 0 when string can not be converted to int",
			arg:  "this is not int",
			want: 0,
		},
		{
			name: "should return 0 when string is empty",
			arg:  "",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toInt(tt.arg); got != tt.want {
				t.Errorf("\n intify() got = %v, want %v \n", got, tt.want)
			}
		})
	}
}
