package points

import "testing"

func TestToInternal(t *testing.T) {
	cases := []struct {
		in   float64
		want int64
	}{
		{0, 0},
		{1, 100},
		{1.23, 123},
		{1.235, 124}, // округление
		{1000.50, 100050},
		{-1.23, -123},
	}
	for _, c := range cases {
		got := ToInternal(c.in)
		if got != c.want {
			t.Errorf("ToInternal(%v) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestToAPI(t *testing.T) {
	cases := []struct {
		in   int64
		want float64
	}{
		{0, 0},
		{100, 1},
		{123, 1.23},
		{124, 1.24},
		{100050, 1000.5},
		{-123, -1.23},
	}
	for _, c := range cases {
		got := ToAPI(c.in)
		if got != c.want {
			t.Errorf("ToAPI(%d) = %v, want %v", c.in, got, c.want)
		}
	}
}
