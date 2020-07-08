package imgix

import (
	"errors"
	"fmt"
	"regexp"
)

// Todo: Review this regex, and the one that follows.
// Matches http:// and https://
var RegexpHTTPAndS = regexp.MustCompile("https?://")

// Regexp for all characters we should escape in a URI passed in.
var RegexUrlCharactersToEscape = regexp.MustCompile("([^ a-zA-Z0-9_.-])")

const zero = 0

// validateMinWidth checks if the minimum, or begin value, is valid.
// A width value is valid if it is greater than, or equal to, zero.
// If the value is less than zero, an error is returned.
func validateMinWidth(value int) (int, error) {
	msg := "`begin` width must be greater than, or equal to, zero"
	if value < zero {
		return -1, errors.New(msg)
	}
	return value, nil
}

// validateMaxWidth checks if the maximum, or end value, is valid.
// A width value is valid if it is greater than, or equal to, zero.
// If the value is less than zero, an error is returned.
func validateMaxWidth(value int) (int, error) {
	msg := "`end` width must be greater than, or equal to, zero"
	if value < zero {
		return -1, errors.New(msg)
	}
	return value, nil
}

// validateWidthTolerance checks if the tol, or width tolerance value,
// is valid. A width tolerance value is valid if it is greater than,
// or equal to, one percent (0.01). If the value is less than one
// percent, an error is returned.
func validateWidthTolerance(value float64) (float64, error) {
	const OnePercent = 0.01
	msg := "`tol`erance must be greater than, or equal to, one percent (0.01)"
	if value < OnePercent {
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
