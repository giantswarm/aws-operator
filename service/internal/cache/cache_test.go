package cache

import (
	"testing"
	"time"
)

func Test_floatCache_Set(t *testing.T) {
	c := NewFloat64Cache(time.Minute * 1)

	testCases := []struct {
		key           string
		value         float64
		expectedValue float64
	}{
		{
			key:           "key",
			value:         0,
			expectedValue: 0,
		},
		{
			key:           "key",
			value:         3.1416,
			expectedValue: 3.1416,
		},
	}

	for _, tc := range testCases {
		c.Set(tc.key, tc.value)

		value, ok := c.Get(tc.key)
		if !ok {
			t.Fatalf("cache key must exist")
		}
		if value != tc.expectedValue {
			t.Fatalf("cache value == %v, want %v", value, tc.expectedValue)
		}
	}
}

func Test_floatCache_Expire(t *testing.T) {
	c := NewFloat64Cache(time.Second * 1)

	key := "pi"
	pi := 3.14
	c.Set(key, pi)

	time.Sleep(time.Second * 1)

	value, _ := c.Get(key)
	if value == pi {
		t.Fatalf("cache value must expire")
	}
}
