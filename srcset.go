package imgix

import (
	"log"
	"math"
	"net/url"
	"strconv"
	"strings"
)

// isDprBased determines if we can infer from params whether we need
// to create a dpr-based srcset attribute. If a width ("w") is present
// or if both the height ("h") and the aspect ratio ("ar") are present,
// then we can infer the desired srcset is dpr-based.
func (b *URLBuilder) isDprBased(params url.Values) bool {
	const EmptyStr = ""
	hasWidth := params.Get("w")
	hasHeight := params.Get("h")
	hasAspectRatio := params.Get("ar")

	if hasWidth != EmptyStr {
		return true
	}

	if hasHeight != EmptyStr && hasAspectRatio != EmptyStr {
		return true
	}
	// Getting "w", "h", and "ar" returned empty strings so none are
	// present in the params, this is _not_ a dpr-based srcset.
	return false
}

// WidthRange contains all the information about a width-range that is
// needed to create a set of target-width values. A WidthRange defines
// a range of widths that begin at the specified "begin" value, end at
// the specified "end" value. The tol, or tolerance percentage value,
// controls the amount of tolerable image-width variation in the range.
// Effectively, this tol percentage value controls the rate at which
// width values increase between the "begin" and the "end" values.
type WidthRange struct {
	begin int
	end   int
	tol   float64
}

// rangePair is a convenience structure used during validation.
// Its purpose is create a consistent interface for our validators.
type rangePair struct {
	begin int
	end   int
}

// SrcSetOpts structures together configuration options for creating
// srcset attributes.
type SrcSetOpts struct {
	widthRange             WidthRange
	disableVariableQuality bool
}

// DefaultOpts structures default srcset options together. Where a default,
// fixed-width (dpr-based) srcset will have variable quality enabled and a
// fluid-width-based (width-paris) srcset will begin at 100, end at 8192
// and have a tolerance of 0.08 (or 8%).
var DefaultOpts = SrcSetOpts{
	widthRange:             WidthRange{begin: minWidth, end: maxWidth, tol: tolerance},
	disableVariableQuality: false,
}

// CreateSrcSet creates a srcset attribute string. Given a path, set of
// parameters, and a set of SrcSetOpts, this function infers which kind
// of srcset attribute to create.
//
// If the params contain a width parameter or both
// height and aspect ratio parameters, a fixed-width srcset attribute
// will be created. This fixed-width srcset attribute will be dpr-based
// and have variable quality turned on by default. Variable quality can
// be disabled by setting the disableVariableQuality field of the
// SrcSetOpts to true.
//
// Otherwise, this function will create a fluid-width srcset attribute
// wherein each URL (or image candidate string) is described by a width
// in the specified WidthRange.
func (b *URLBuilder) CreateSrcSet(path string, params url.Values, opts SrcSetOpts) string {
	// Check params contains a width (w) or height (h) _and_ aspect ratio (ar);
	hasWidth := params.Get("w") != ""
	hasHeight := params.Get("h") != ""
	hasAspectRatio := params.Get("ar") != ""

	// If params has either a width or _both_ height and aspect ratio,
	// build a dpr-based srcset attribute.
	if hasWidth || (hasHeight && hasAspectRatio) {
		return b.buildSrcSetDpr(path, params, opts.disableVariableQuality)
	}

	// Otherwise, get the widthRange values from the opts and build a
	// width-pairs based srcset attribute.
	begin := opts.widthRange.begin
	end := opts.widthRange.end
	tol := opts.widthRange.tol
	targets := TargetWidths(begin, end, tol)
	return b.buildSrcSetPairs(path, params, targets)
}

// CreateSrcSetFromRange creates a srcset attribute whose URLs
// are described by the widths within the specified range. begin,
// end, and tol (tolerance) define the widths-range. The range
// begins at the minimum value and ends at the maximal value; the
// tol or tolerance dictates the amount of tolerable image width
// variation between each width in the range.
func (b *URLBuilder) CreateSrcSetFromRange(path string, params url.Values, wr WidthRange) string {
	targets := TargetWidths(wr.begin, wr.end, wr.tol)
	return b.buildSrcSetPairs(path, params, targets)
}

// CreateSrcSetFromWidths takes a path, a set of params, and an array of widths
// to create a srcset attribute with width-described URLs (image candidate strings).
func (b *URLBuilder) CreateSrcSetFromWidths(path string, params url.Values, widths []int) string {
	return b.buildSrcSetPairs(path, params, widths)
}

// buildSrcSetPairs builds a srcset attribute string containing width-described
// image candidate strings.
func (b *URLBuilder) buildSrcSetPairs(path string, params url.Values, targets []int) string {
	var srcSetEntries []string

	for _, w := range targets {
		widthValue := strconv.Itoa(w)
		params.Set("w", widthValue)
		entry := b.createImageCandidateString(path, params, widthValue+"w")
		srcSetEntries = append(srcSetEntries, entry)
	}
	return strings.Join(srcSetEntries, ",\n")
}

func (b *URLBuilder) buildSrcSetDpr(path string, params url.Values, disableVariableQuality bool) string {
	var DprQualities = map[string]string{"1": "75", "2": "50", "3": "35", "4": "23", "5": "20"}
	var srcSetEntries []string

	// We could iterate over the map directly, but that doesn't yield
	// deterministic results, ie. 5x might come before 1x in the final
	// srcset attribute string. To prevent this, we iterate over the
	// map "in order."
	for i := 0; i < len(DprQualities); i++ {
		ratio := strconv.Itoa(i + 1)
		params.Set("dpr", ratio)
		dprQuality := DprQualities[ratio]

		// If variable quality is disabled, then first try to get
		// any `q` param
		if disableVariableQuality {
			qValue := params.Get("q")
			if qValue != "" {
				params.Set("q", qValue)
			}
		} else {
			params.Set("q", dprQuality)
		}

		entry := b.createImageCandidateString(path, params, ratio+"x")
		srcSetEntries = append(srcSetEntries, entry)
	}
	return strings.Join(srcSetEntries, ",\n")
}

// tolerance is the default width tolerance percentage.
const tolerance float64 = 0.08

// minWidth is the default minimum width used to begin a
// srcset width range. Widths can be below this value; this
// is just the value used internally in the TargetWidths
// function.
const minWidth int = 100

// maxWidth is the default maximum width used to end a
// srcset width range. While width values can be above
// this value, they are typically less than or equal to
// this value. This is only a value used internally in
// the TargetWidths function.
const maxWidth int = 8192

// DefaultWidths is an array of image widths generated by
// calling TargetWidths(100, 8192, 0.08). These defaults are quite
// good, cover a wide range of widths, and are easy to start with.
var DefaultWidths = []int{
	100, 116, 135, 156,
	181, 210, 244, 283,
	328, 380, 441, 512,
	594, 689, 799, 927,
	1075, 1247, 1446, 1678,
	1946, 2257, 2619, 3038,
	3524, 4087, 4741, 5500,
	6380, 7401, 8192}

// TargetWidths creates an array of integer image widths.
// The image widths begin at the minimum value and end at
// the maximum width value with a tol amount of tolerable
// image width variance between them.
func TargetWidths(begin int, end int, tol float64) []int {
	validRange, err := validateRangeWithTolerance(begin, end, tol)
	if err != nil {
		log.Fatalln(err)
	}
	begin = validRange.begin
	end = validRange.end
	tol = validRange.tol

	if isNotCustom(begin, end, tol) {
		return DefaultWidths
	}

	if begin == end {
		return []int{begin}
	}
	var resolutions []int
	var start = float64(begin)

	for int(start) < end && int(start) < maxWidth {
		resolutions = append(resolutions, int(math.Round(start)))
		start = start * (1.0 + tol*2.0)
	}
	lengthOfResolutions := len(resolutions)

	// If we make it here, the lengthOfResolutions is greater
	// than or equal to 2, so accessing the last element of
	// the slice should not panic.
	if resolutions != nil && resolutions[lengthOfResolutions-1] < end {
		resolutions = append(resolutions, end)
	}
	return resolutions
}

// isNotCustom takes a "begin" value and an "end" value along with a
// tol, or tolerance value, and compares each to its respective
// default value. If every value is equal to its default value
// then this range isNotCustom, return true.
func isNotCustom(begin int, end int, tol float64) bool {
	defaultBegin := begin == minWidth
	defaultEnd := end == maxWidth
	defaultTol := tol == tolerance
	return defaultBegin && defaultEnd && defaultTol
}

// createImageCandidateString joins a URL with a space and a suffix in order
// to create an image candidate string. For more information see:
// https://html.spec.whatwg.org/multipage/images.html#srcset-attributes
func (b *URLBuilder) createImageCandidateString(path string, params url.Values, suffix string) string {
	return strings.Join([]string{b.CreateURL(path, params), " ", suffix}, "")
}
