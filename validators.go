package imgix

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// rangePair is a convenience structure used during validation.
// Its purpose is create a consistent interface for our validators.
type rangePair struct {
	minWidth int
	maxWidth int
}

// widthRange contains all the information about a width-range that is
// needed to create a set of target-width values.
type widthRange struct {
	minWidth  int
	maxWidth  int
	tolerance float64
}

// validateDomain uses Go's url.Parse and url.Hostname functions to
// validate the domain. Elsewhere we use a regex to filter invalid
// domains. However, the same regex won't work in this case as Go
// does not support positive look-a-heads (i.e. `(?=)`).
func validateDomain(domain string) (string, error) {
	if strings.HasPrefix(domain, "http") {
		u, err := url.Parse(domain)
		if err != nil {
			return "", fmt.Errorf(
				"failed to parse URL form from domain %s due to %w", domain, err)
		}
		return u.Hostname(), nil
	}

	// Otherwise, apply a "dummy" prefix so that the domain (hostname)
	// is parsed correctly.
	u, err := url.Parse("https://" + domain)
	if err != nil {
		return "", fmt.Errorf(
			"failed to parse domain %s with scheme: https, due to: %w", domain, err)
	}
	return u.Hostname(), nil

}

// validateMinWidth checks if the value is a valid minWidth.
// A minWidth value is valid if it is greater than, or equal to, zero.
// If the value is less than zero, an error is returned.
func validateMinWidth(minWidth int) (int, error) {
	msg := "`minWidth` value must be greater than, or equal to, zero"
	if minWidth < 0 {
		return -1, errors.New(msg)
	}
	return minWidth, nil
}

// validateMaxWidth checks if the value is a valid maxWidth.
// A maxWidth value is valid if it is greater than, or equal to, zero.
// If the value is less than zero, an error is returned.
func validateMaxWidth(maxWidth int) (int, error) {
	msg := "`maxWidth` value must be greater than, or equal to, zero"
	if maxWidth < 0 {
		return -1, errors.New(msg)
	}
	return maxWidth, nil
}

// validateWidthTolerance checks if the vallue is a valid tolerance value.
// A width tolerance value is valid if it is greater than, or equal to,
// one percent (0.01). If the value is less than one percent, an error
// is returned.
func validateWidthTolerance(value float64) (float64, error) {
	const onePercent = 0.01
	msg := "`defaultTolerance`erance must be greater than, or equal to, one percent (0.01)"
	if value < onePercent {
		return -1, errors.New(msg)
	}
	return value, nil
}

// validateRange checks that the range defined by minWidth and maxWidth is
// valid. The range defined by minWidth and maxWidth is valid if both
// values pass their checks and if the range is found to be increasing.
func validateRange(minWidth int, maxWidth int) (rangePair, error) {
	// This invalidRangePair is used as a return value on error.
	invalidRangePair := rangePair{-1, -1}

	validMin, minErr := validateMinWidth(minWidth)
	if minErr != nil {
		return invalidRangePair, minErr
	}

	validMax, maxErr := validateMaxWidth(maxWidth)
	if maxErr != nil {
		return invalidRangePair, maxErr
	}

	// Check if range increases.
	if validMax < validMin {
		msg := "`minWidth` must be less than or equal to the `maxWidth`"
		return rangePair{-1, -1}, errors.New(msg)
	}
	return rangePair{validMin, validMax}, nil
}

// validateRangeWithTolerance checks that the range defined by
// minWidth, maxWidth, and tolerance is valid.
func validateRangeWithTolerance(minWidth int, maxWidth int, tolerance float64) (widthRange, error) {
	rp, rangeErr := validateRange(minWidth, maxWidth)
	if rangeErr != nil {
		return widthRange{-1, -1, -1.0}, rangeErr
	}

	validTol, tolErr := validateWidthTolerance(tolerance)
	if tolErr != nil {
		return widthRange{-1, -1, -1.0}, tolErr
	}
	return widthRange{
		minWidth:  rp.minWidth,
		maxWidth:  rp.maxWidth,
		tolerance: validTol}, nil
}

// validateWidths checks that an array is comprised of only positive
// integers. An error is when the first negative value is encountered.
func validateWidths(widthValues []int) ([]int, error) {
	idx, allPositive := allPositive(widthValues)

	if !allPositive {
		msg := fmt.Sprintf("width values must be positive, "+
			"found negative width at index `%d`", idx)
		return []int{}, errors.New(msg)
	}
	return widthValues, nil
}

// allPositive returns true if every value in values is positive, false otherwise.
func allPositive(values []int) (int, bool) {
	const zero = 0
	var idx int
	for idx, v := range values {
		if v < zero {
			return idx, false
		}
	}
	return idx, true
}
