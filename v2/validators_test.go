package imgix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidators_validateNegativeWidths(t *testing.T) {
	widths := []int{100, 200, 300, -400, -500}
	validWidths, err := validateWidths(widths)

	// Ensure an error occurred, and the `err` is `NotEqual` to `nil`.
	assert.NotEqual(t, nil, err)
	assert.Equal(t, []int{}, validWidths)
}

func TestValidators_validatePositiveWidths(t *testing.T) {
	expected := []int{101, 202, 303, 404, 505}
	validWidths, err := validateWidths(expected)

	// Check the `err` is nil.
	assert.Equal(t, nil, err)
	// Check the expected widths are valid widths.
	assert.Equal(t, expected, validWidths)
}

func TestValidators_validateMinWidthValid(t *testing.T) {
	const OneHundred = 100
	validValue, err := validateMinWidth(OneHundred)
	assert.Equal(t, OneHundred, validValue)
	assert.Equal(t, nil, err)
}

func TestValidators_validateMinWidthInvalid(t *testing.T) {
	const LessThanZero = -1
	invalidValue, err := validateMinWidth(LessThanZero)
	assert.Equal(t, -1, invalidValue)
	assert.NotEqual(t, nil, err)
}

func TestValidators_validateMaxWidthValid(t *testing.T) {
	const OneHundred = 100
	validValue, err := validateMaxWidth(OneHundred)
	assert.Equal(t, OneHundred, validValue)
	assert.Equal(t, nil, err)
}

func TestValidators_validateMaxWidthInvalid(t *testing.T) {
	const LessThanZero = -1
	invalidValue, err := validateMaxWidth(LessThanZero)
	assert.Equal(t, -1, invalidValue)
	assert.NotEqual(t, nil, err)
}

func TestValidators_validateRangeInvalid(t *testing.T) {
	begin := 740
	end := 320

	_, err := validateRange(begin, end)
	assert.NotEqual(t, nil, err)
}

func TestValidators_validateRangeValid(t *testing.T) {
	rp := rangePair{minWidth: 100, maxWidth: 8192}
	validRangePair, err := validateRange(rp.minWidth, rp.maxWidth)
	assert.Equal(t, rp, validRangePair)
	assert.Equal(t, nil, err)
}

func TestValidators_validateRangeWithToleranceInvalid(t *testing.T) {
	invalidTolerance := 0.001
	_, err := validateRangeWithTolerance(100, 200, invalidTolerance)
	assert.NotEqual(t, nil, err)
}

func TestValidators_validateRangeWithToleranceValid(t *testing.T) {
	invalidTolerance := 1.25
	_, err := validateRangeWithTolerance(100, 200, invalidTolerance)
	assert.Equal(t, nil, err)
}
