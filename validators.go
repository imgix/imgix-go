package imgix

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// validateDomain uses Go's url.Parse and url.Hostname functions to
// validate the domain. Elsewhere we use a regex to filter invalid
// domains. However, the same regex won't work in this case as Go
// does not support positive look-a-heads (i.e. `(?=)`).
// TODO: Discuss. Go doesn't support positive look-a-heads so our
// domain regex won't work here. I will explore regex alternatives.
func validateDomain(domain string) (string, error) {
	if strings.HasPrefix(domain, "http") {
		u, err := url.Parse(domain)
		if err != nil {
			return "", err
		}
		return u.Hostname(), nil
	}

	// Otherwise, apply a "dummy" prefix so that the domain (hostname)
	// is parsed correctly.
	u, err := url.Parse("https://" + domain)
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil

}

// validateMinWidth checks if the minimum, or begin value, is valid.
// A width value is valid if it is greater than, or equal to, zero.
// If the value is less than zero, an error is returned.
func validateMinWidth(value int) (int, error) {
	msg := "`begin` width must be greater than, or equal to, zero"
	if value < 0 {
		return -1, errors.New(msg)
	}
	return value, nil
}

// validateMaxWidth checks if the maximum, or end value, is valid.
// A width value is valid if it is greater than, or equal to, zero.
// If the value is less than zero, an error is returned.
func validateMaxWidth(value int) (int, error) {
	msg := "`end` width must be greater than, or equal to, zero"
	if value < 0 {
		return -1, errors.New(msg)
	}
	return value, nil
}

// validateWidthTolerance checks if the tol, or width tolerance value,
// is valid. A width tolerance value is valid if it is greater than,
// or equal to, one percent (0.01). If the value is less than one
// percent, an error is returned.
func validateWidthTolerance(value float64) (float64, error) {
	const onePercent = 0.01
	msg := "`tol`erance must be greater than, or equal to, one percent (0.01)"
	if value < onePercent {
		return -1, errors.New(msg)
	}
	return value, nil
}

// validateRange checks that the range defined by begin and end is
// valid. The range defined by begin and end is valid if the begin
// value is less than or equal to the end value. If the end value
// is less than the begin value, an error is returned.
func validateRange(begin int, end int) (rangePair, error) {
	// This invalidRangePair is used as a return value on error.
	invalidRangePair := rangePair{-1, -1}

	validBegin, beginErr := validateMinWidth(begin)
	if beginErr != nil {
		return invalidRangePair, beginErr
	}

	validEnd, endErr := validateMaxWidth(end)
	if endErr != nil {
		return invalidRangePair, endErr
	}

	if validEnd < validBegin {
		// If the "begin width" is greater than the "end width"
		// for the range, error!
		msg := "`begin` width must be less than or equal to the `end` width"
		return rangePair{-1, -1}, errors.New(msg)
	}
	return rangePair{validBegin, validEnd}, nil
}

// validateRangeWithTolerance checks that the range defined by begin,
// end, and tol is valid. First, begin and end values are validated
// using validateRange. Next, the tolerance is validated using
// validateWidthTolerance. If we get a valida range pair and tolerance,
// we can return a valid WidthRange; otherwise, an invalid width range
// is returned along with the error that occurred.
func validateRangeWithTolerance(begin int, end int, tol float64) (WidthRange, error) {
	rp, rangeErr := validateRange(begin, end)
	if rangeErr != nil {
		return WidthRange{-1, -1, -1.0}, rangeErr
	}

	validTol, tolErr := validateWidthTolerance(tol)
	if tolErr != nil {
		return WidthRange{-1, -1, -1.0}, tolErr
	}
	return WidthRange{begin: rp.begin, end: rp.end, tol: validTol}, nil
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
