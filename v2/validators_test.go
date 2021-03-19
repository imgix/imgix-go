package imgix

import (
	"testing"
)

func TestValidators_validateNegativeWidths(t *testing.T) {
	widths := []int{100, 200, 300, -400, -500}
	got, err := validateWidths(widths)

	// Assert an error occurred, i.e. that the `err` is NOT `nil`.
	// If the err is nil, fail.
	if err == nil {
		t.Errorf("got: err == nil; want: err != nil")
	}

	want := []int{}

	// Assert the lengths of the wanted widths and gotten widths
	// are the same (i.e. that both are empty arrays).
	if len(got) != len(want) {
		t.Errorf("got: %v; want: %v", got, want)
	}
}

func TestValidators_validatePositiveWidths(t *testing.T) {
	want := []int{101, 202, 303, 404, 505}
	got, err := validateWidths(want)

	// Assert NO error occurred, i.e. err SHOULD BE nil.
	// If not, fail.
	if err != nil {
		t.Errorf("got: err != nil; want: err == nil")
	}

	// We want the lengths to be equal, fail if they aren't.
	if len(got) != len(want) {
		t.Errorf("got: len(got) != len(want); want: len(got) == len(want)")
	}

	// Check that we got the widths we wanted. If not, fail and detail
	// the indices at which the arrays differ.
	for idx, v := range want {
		if got[idx] != v {
			t.Errorf("got: %v; want: %v; arrays differ at index %d", got[idx], v, idx)
		}
	}
}

func TestValidators_validateMinWidthValid(t *testing.T) {
	const want = 100
	got, err := validateMinWidth(want)

	// Assert NO error occurred, i.e. err SHOULD BE nil.
	// If not, fail.
	if err != nil {
		t.Errorf("got: err != nil; want: err == nil")
	}

	// We want 100, but if we got something else, FAIL!
	if got != want {
		t.Errorf("got: %v; want: %v", got, want)
	}
}

func TestValidators_validateMinWidthInvalid(t *testing.T) {
	const LessThanZero = -1
	_, err := validateMinWidth(LessThanZero)

	// Assert an error occurred, i.e. that the `err` is NOT `nil`.
	// If the err is nil, fail.
	if err == nil {
		t.Errorf("got: err == nil; want: err != nil")
	}
}

func TestValidators_validateMaxWidthValid(t *testing.T) {
	const want = 100
	got, err := validateMaxWidth(want)

	// Assert NO error occurred, i.e. err SHOULD BE nil.
	// If not, fail.
	if err != nil {
		t.Errorf("got: err != nil; want: err == nil")
	}

	if got != want {
		t.Errorf("got: %v; want: %v", got, want)
	}
}

func TestValidators_validateMaxWidthInvalid(t *testing.T) {
	const LessThanZero = -1
	_, err := validateMaxWidth(LessThanZero)

	// Assert an error occurred, i.e. that the `err` is NOT `nil`.
	// If the err is nil, fail.
	if err == nil {
		t.Errorf("got: err == nil; want: err != nil")
	}
}

func TestValidators_validateRangeInvalid(t *testing.T) {
	begin := 740
	end := 320

	_, err := validateRange(begin, end)

	// Assert an error occurred, i.e. that the `err` is NOT `nil`.
	// If the err is nil, fail.
	if err == nil {
		t.Errorf("got: err == nil; want: err != nil")
	}
}

func TestValidators_validateRangeValid(t *testing.T) {
	want := rangePair{minWidth: 100, maxWidth: 8192}
	got, err := validateRange(want.minWidth, want.maxWidth)

	// Assert NO error occurred, i.e. err SHOULD BE nil.
	// If not, fail.
	if err != nil {
		t.Errorf("got: err != nil; want: err == nil")
	}

	if got != want {
		t.Errorf("got: %v; want: %v", got, want)
	}
}

func TestValidators_validateRangeWithToleranceInvalid(t *testing.T) {
	invalidTolerance := 0.001
	_, err := validateRangeWithTolerance(100, 200, invalidTolerance)

	// Assert an error occurred, i.e. that the `err` is NOT `nil`.
	// If the err is nil, fail.
	if err == nil {
		t.Errorf("got: err == nil; want: err != nil")
	}
}

func TestValidators_validateRangeWithToleranceValid(t *testing.T) {
	const want = 1.25
	got, err := validateRangeWithTolerance(100, 200, want)

	// Assert NO error occurred, i.e. err SHOULD BE nil.
	// If not, fail.
	if err != nil {
		t.Errorf("got: err != nil; want: err == nil")
	}

	if got.tolerance != want {
		t.Errorf("got: %v; want: %v", got.tolerance, want)
	}
}
