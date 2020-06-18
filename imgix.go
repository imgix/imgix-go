package imgix

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Builder facilitates the building of URLs.
type Builder struct {
	domain   string
	token    string
	useHTTPS bool
}

// NewBuilder creates a new Builder with the given domain, with HTTPS enabled.
func NewBuilder(domain string) Builder {
	return Builder{domain: domain, useHTTPS: true}
}

// NewBuilderWithToken creates a new Builder with the given domain and token
// with HTTPS enabled.
func NewBuilderWithToken(domain string, token string) Builder {
	return Builder{domain: domain, useHTTPS: true, token: token}
}

// UseHTTPS returns whether HTTPS or HTTP should be used.
func (b *Builder) UseHTTPS() bool {
	return b.useHTTPS
}

// SetUseHTTPS sets a builder's useHTTPS field to true or false. Setting
// useHTTPS to false forces the builder to use HTTP.
func (b *Builder) SetUseHTTPS(useHTTPS bool) {
	b.useHTTPS = useHTTPS
}

// Scheme gets the URL scheme to use, either "http" or "https"
// (the scheme uses HTTPS by default).
func (b *Builder) Scheme() string {
	if b.UseHTTPS() {
		return "https"
	} else {
		return "http"
	}
}

// TODO: Review this regex-replace-all code.
// Domain gets the builder's domain string.
func (b *Builder) Domain() string {
	return RegexpHTTPAndS.ReplaceAllString(b.domain, "") // Strips out the scheme if exists
}

// SetToken sets the token for this builder. This value will be used to sign
// URLs created through the builder.
func (b *Builder) SetToken(token string) {
	b.token = token
}

// CreateURL creates a URL string given a path and a set of
// params.
func (b *Builder) CreateURL(path string, params url.Values) string {
	return b.createAndMaybeSignURL(path, params, false)
}

// CreateSignedURL is like CreateURL except that it creates a signed URL.
func (b *Builder) CreateSignedURL(path string, params url.Values) string {
	return b.createAndMaybeSignURL(path, params, true)
}

// CreateURLFromPath creates a URL string given a path.
func (b *Builder) CreateURLFromPath(path string) string {
	return b.createAndMaybeSignURL(path, url.Values{}, false)
}

// CreateSignedURLFromPath is like CreateURLFromPath except that it creates
// a full URL to the image that has been signed using the builder's token.
func (b *Builder) CreateSignedURLFromPath(path string) string {
	return b.createAndMaybeSignURL(path, url.Values{}, true)
}

// createURLFromPathAndParams will manually build a URL from a given path string and
// parameters passed in. Because of the differences in how net/url escapes
// path components, we need to manually build a URL as best we can.
func (b *Builder) createAndMaybeSignURL(path string, params url.Values, shouldSign bool) string {
	u := url.URL{
		Scheme: b.Scheme(),
		Host:   b.Domain(),
	}
	urlString := u.String()

	// TODO: Review this cgiEscape code.
	// If we are given a fully-qualified URL, escape it per the note located
	// near the `cgiEscape` function definition.
	if RegexpHTTPAndS.MatchString(path) {
		path = cgiEscape(path)
	}
	// TODO: This portion is still a little busy...
	path = maybePrependSlash(path)
	urlWithPath := strings.Join([]string{urlString, path}, "")
	maybeBase64EncodeParameters(&params)

	var parameterString string
	if shouldSign {
		signature := b.signPathAndParams(path, params)
		parameterString = createParameterString(params, signature)
	} else {
		parameterString = createParameterString(params, "")
	}

	if parameterString != "" {
		return strings.Join([]string{urlWithPath, parameterString}, "?")
	}
	return urlWithPath
}

// signPathAndParams takes a builder's token value, a path, and a set of
// params and creates a signature from these values.
func (b *Builder) signPathAndParams(path string, params url.Values) string {
	hasQueryParams := false

	if len(params) > 0 {
		hasQueryParams = true
	}
	queryParams := params.Encode()
	signature := createMd5Signature(b.token, path, queryParams)

	if hasQueryParams {
		return strings.Join([]string{"s=", signature}, "")
	}
	return strings.Join([]string{"s=", signature}, "")
}

// createMd5Signature creates the signature by joining the token, path, and params
// strings into a signatureBase. Next, create a hashedSig and write the
// signatureBase into it. Finally, return the encoded, signed string.
func createMd5Signature(token string, path string, params string) string {
	signatureBase := strings.Join([]string{token, path, params}, "")
	hashedSig := md5.New()
	hashedSig.Write([]byte(signatureBase))
	return hex.EncodeToString(hashedSig.Sum(nil))
}

// maybeBase64EncodeParameters base64-encodes a parameter
// if the parameter has the "64" suffix.
func maybeBase64EncodeParameters(params *url.Values) {
	for key, val := range *params {
		if strings.HasSuffix(key, "64") {
			encodedParam := base64EncodeParameter(val[0])
			params.Set(key, encodedParam)
		}
	}
}

// TODO: Revisit this encoding and replacement code.
func createParameterString(params url.Values, signature string) string {
	parameterString := params.Encode()
	parameterString = strings.Replace(parameterString, "+", "%%20", -1)

	if signature != "" && len(params) > 0 {
		parameterString += "&" + signature
	} else if signature != "" && len(params) == 0 {
		parameterString = signature
	}
	return parameterString
}

// maybePrependSlash prepends if the path does not begin with one:
// "users/1.png" -> "/users/1.png"
func maybePrependSlash(path string) string {
	const Slash = "/"

	if strings.Index(path, Slash) != 0 {
		path = strings.Join([]string{Slash, path}, "")
	}
	return path
}

// Base64-encodes a parameter according to imgix's Base64 variant requirements.
// https://docs.imgix.com/apis/url#base64-variants
func base64EncodeParameter(param string) string {
	paramData := []byte(param)
	base64EncodedParam := base64.URLEncoding.EncodeToString(paramData)
	base64EncodedParam = strings.Replace(base64EncodedParam, "=", "", -1)

	return base64EncodedParam
}

// Todo: Review this regex and the one that follows.
// Matches http:// and https://
var RegexpHTTPAndS = regexp.MustCompile("https?://")

// Regexp for all characters we should escape in a URI passed in.
var RegexUrlCharactersToEscape = regexp.MustCompile("([^ a-zA-Z0-9_.-])")

// This code is less than ideal, but it's the only way we've found out how to do it
// give Go's URL capabilities and escaping behavior.
//
// See: https://github.com/parkr/imgix-go/pull/1#issuecomment-109014369 and
// https://github.com/imgix/imgix-blueprint#securing-urls
func cgiEscape(s string) string {
	return RegexUrlCharactersToEscape.ReplaceAllStringFunc(s, func(s string) string {
		runeValue, _ := utf8.DecodeLastRuneInString(s)
		return "%" + strings.ToUpper(fmt.Sprintf("%x", runeValue))
	})
}

// Tolerance is the default width tolerance percentage.
const Tolerance float64 = 0.08

// MinWidth is the default minimum width used to begin a
// srcset width range. Widths can be below this value; this
// is just the value used internally in the TargetWidths
// function.
const MinWidth int = 100

// MaxWidth is the default maximum width used to end a
// srcset width range. While width values can be above
// this value, they are typically less than or equal to
// this value. This is only a value used internally in
// the TargetWidths function.
const MaxWidth int = 8192

// DefaultTargetWidths is an array of image widths generated by
// calling TargetWidths(100, 8192, 0.08). These defaults are quite
// good, cover a wide range of widths, and are easy to start with.
var DefaultTargetWidths = []int{
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
	if isNotCustom(begin, end, tol) {
		return DefaultTargetWidths
	}

	if begin == end {
		return []int{begin}
	}
	var resolutions []int
	var start = float64(begin)

	for int(start) < end && int(start) < MaxWidth {
		resolutions = append(resolutions, int(math.Round(start)))
		start = start * (1.0 + tol*2.0)
	}
	lengthOfResolutions := len(resolutions)

	if lengthOfResolutions < 2 {
		return resolutions
	}

	// If we make it here, the lengthOfResolutions is greater
	// than or equal to 2, so accessing the last element of
	// the slice should not panic.
	if resolutions[lengthOfResolutions-1] < end {
		resolutions = append(resolutions, end)
	}
	return resolutions
}

// isNotCustom takes a "begin" value and an "end" value along with a
// tol, or tolerance value, and compares each to its respective
// default value. If every value is equal to its default value
// then this range isNotCustom, return true.
func isNotCustom(begin int, end int, tol float64) bool {
	defaultBegin := begin == MinWidth
	defaultEnd := end == MaxWidth
	defaultTol := tol == Tolerance
	return defaultBegin && defaultEnd && defaultTol
}

// CreateSrcSetFromRange creates a srcset attribute whose URLs
// are described by the widths within the specified range. The
// range is defined by begin, end, and tol (tolerance). The range
// begins at the minimum value and ends at the maximal value; the
// tol or tolerance dictates the amount of tolerable image width
// variation between each width in the range.
func (b *Builder) CreateSrcSetFromRange(path string, params url.Values, begin int, end int, tol float64) string {
	targets := TargetWidths(begin, end, tol)
	return b.buildSrcSetPairs(path, params, targets)
}

// CreateSrcSetFromWidths takes a path, a set of params, and an array of widths
// to create a srcset attribute with width-described URLs (image candidate strings).
func (b *Builder) CreateSrcSetFromWidths(path string, params url.Values, widths []int) string {
	return b.buildSrcSetPairs(path, params, widths)
}

// buildSrcSetPairs builds a srcset attribute string containing width-described
// image candidate strings.
func (b *Builder) buildSrcSetPairs(path string, params url.Values, targets []int) string {
	var srcSetEntries []string

	for _, w := range targets {
		widthValue := strconv.Itoa(w)
		params.Set("w", widthValue)
		entry := b.createImageCandidateString(path, params, widthValue+"w")
		srcSetEntries = append(srcSetEntries, entry)
	}
	return strings.Join(srcSetEntries, ",\n")
}

func (b *Builder) buildSrcSetDpr(path string, params url.Values, disableVariableQuality bool) string {
	var DprRatios = map[int]int{1: 75, 2: 50, 3: 35, 4: 23, 5: 20}
	var srcSetEntries []string

	for dprRatio, dprQuality := range DprRatios {
		ratio := strconv.Itoa(dprRatio)
		params.Set("dpr", ratio)

		// If variable quality has not been disabled,
		// attempt to get the "q" param. If the "q"
		// param is not found in the params, then an
		// empty string will be returned. In this case,
		// set the "q" params' value to be dprQuality
		if !disableVariableQuality {
			qParam := params.Get("q")
			if qParam == "" {
				params.Set("q", strconv.Itoa(dprQuality))
			}
		}
		entry := b.createImageCandidateString(path, params, ratio+"x")
		srcSetEntries = append(srcSetEntries, entry)
	}
	return strings.Join(srcSetEntries, ",\n")
}

// createImageCandidateString joins a URL with a space and a suffix in order
// to create an image candidate string. For more information see:
// https://html.spec.whatwg.org/multipage/images.html#srcset-attributes
func (b *Builder) createImageCandidateString(path string, params url.Values, suffix string) string {
	return strings.Join([]string{b.CreateURL(path, params), " ", suffix}, "")
}

// isDprBased determines if we can infer from params whether we need
// to create a dpr-based srcset attribute. If a width ("w") is present
// or if both the height ("h") and the aspect ratio ("ar") are present,
// then we can infer the desired srcset is dpr-based.
func (b *Builder) isDprBased(params url.Values) bool {
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
