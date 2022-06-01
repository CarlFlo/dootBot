package test

import "testing"

func Validate(t *testing.T, got interface{}, expected interface{}, msg string) {
	if got != expected {
		t.Error("Expected", expected, ", got", got, ":", msg)
	}
}
