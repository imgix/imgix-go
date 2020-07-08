package imgix

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// URLBuilder facilitates the building of URLs.
type URLBuilder struct {
	domain   string
	token    string
	useHTTPS bool
}

// NewURLBuilder creates a new URLBuilder with the given domain, with HTTPS enabled.
func NewURLBuilder(domain string) URLBuilder {
	return URLBuilder{domain: domain, useHTTPS: true}
}

// NewURLBuilderWithToken creates a new URLBuilder with the given domain and token
// with HTTPS enabled.
func NewURLBuilderWithToken(domain string, token string) URLBuilder {
	return URLBuilder{domain: domain, useHTTPS: true, token: token}
}

// UseHTTPS returns whether HTTPS or HTTP should be used.
func (b *URLBuilder) UseHTTPS() bool {
	return b.useHTTPS
}

// SetUseHTTPS sets a builder's useHTTPS field to true or false. Setting
// useHTTPS to false forces the builder to use HTTP.
func (b *URLBuilder) SetUseHTTPS(useHTTPS bool) {
	b.useHTTPS = useHTTPS
}

// Scheme gets the URL scheme to use, either "http" or "https"
// (the scheme uses HTTPS by default).
func (b *URLBuilder) Scheme() string {
	if b.UseHTTPS() {
		return "https"
	}
	return "http"
}

// TODO: Review this regex-replace-all code.
// Domain gets the builder's domain string.
func (b *URLBuilder) Domain() string {
	return RegexpHTTPAndS.ReplaceAllString(b.domain, "") // Strips out the scheme if exists
}

// SetToken sets the token for this builder. This value will be used to sign
// URLs created through the builder.
func (b *URLBuilder) SetToken(token string) {
	b.token = token
}

// CreateURL creates a URL string given a path and a set of
// params.
func (b *URLBuilder) CreateURL(path string, params url.Values) string {
	if b.token != "" {
		return b.createAndMaybeSignURL(path, params, true)
	}
	return b.createAndMaybeSignURL(path, params, false)
}

// CreateSrcSet creates a srcset attribute string. Given a path, set of
// parameters, and a Config, this function infers which kind of srcset
// attribute to create.
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

	// Otherwise, get the widthRange values from the config and build a
	// width-pairs based srcset attribute.
	begin := opts.widthRange.begin
	end := opts.widthRange.end
	tol := opts.widthRange.tol
	targets := TargetWidths(begin, end, tol)
	return b.buildSrcSetPairs(path, params, targets)
}

// CreateSignedURL is like CreateURL except that it creates a signed URL.
func (b *URLBuilder) CreateSignedURL(path string, params url.Values) string {
	return b.createAndMaybeSignURL(path, params, true)
}

// CreateURLFromPath creates a URL string given a path.
func (b *URLBuilder) CreateURLFromPath(path string) string {
	return b.createAndMaybeSignURL(path, url.Values{}, false)
}

// CreateSignedURLFromPath is like CreateURLFromPath except that it creates
// a full URL to the image that has been signed using the builder's token.
func (b *URLBuilder) CreateSignedURLFromPath(path string) string {
	return b.createAndMaybeSignURL(path, url.Values{}, true)
}

// createURLFromPathAndParams will manually build a URL from a given path string and
// parameters passed in. Because of the differences in how net/url escapes
// path components, we need to manually build a URL as best we can.
func (b *URLBuilder) createAndMaybeSignURL(path string, params url.Values, shouldSign bool) string {
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
func (b *URLBuilder) signPathAndParams(path string, params url.Values) string {
	queryParams := strings.Join(encodeQueryParamsFromValues(params), "&")
	signature := createMd5Signature(b.token, path, queryParams)
	return strings.Join([]string{"s=", signature}, "")
}

// createMd5Signature creates the signature by joining the token, path, and params
// strings into a signatureBase. Next, create a hashedSig and write the
// signatureBase into it. Finally, return the encoded, signed string.
func createMd5Signature(token string, path string, params string) string {
	var delim string

	if params == "" {
		delim = ""
	} else {
		delim = "?"
	}

	signatureBase := strings.Join([]string{token, path, delim, params}, "")
	hashedSig := md5.New()
	hashedSig.Write([]byte(signatureBase))
	return hex.EncodeToString(hashedSig.Sum(nil))
}

func createParameterString(params url.Values, signature string) string {
	encodedParameters := encodeQueryParamsFromValues(params)
	parameterString := strings.Join(encodedParameters, "&")

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

// createImageCandidateString joins a URL with a space and a suffix in order
// to create an image candidate string. For more information see:
// https://html.spec.whatwg.org/multipage/images.html#srcset-attributes
func (b *URLBuilder) createImageCandidateString(path string, params url.Values, suffix string) string {
	return strings.Join([]string{b.CreateURL(path, params), " ", suffix}, "")
}

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

// Todo: Review this regex, and the one that follows.
// Matches http:// and https://
var RegexpHTTPAndS = regexp.MustCompile("https?://")

// Regexp for all characters we should escape in a URI passed in.
var RegexUrlCharactersToEscape = regexp.MustCompile("([^ a-zA-Z0-9_.-])")

// This code is less than ideal, but it's the only way we've found out how to do it
// give Go's URL capabilities and escaping behavior.
//
// See:
// https://github.com/parkr/imgix-go/pull/1#issuecomment-109014369 and
// https://github.com/imgix/imgix-blueprint#securing-urls
func cgiEscape(s string) string {
	return RegexUrlCharactersToEscape.ReplaceAllStringFunc(s, func(s string) string {
		runeValue, _ := utf8.DecodeLastRuneInString(s)
		return "%" + strings.ToUpper(fmt.Sprintf("%x", runeValue))
	})
}

func encodePathOrProxy(p string) string {
	if isProxy, isEncoded := checkProxyStatus(p); isProxy {
		return encodeProxy(p, isEncoded)
	}
	return encodePath(p)
}

func checkProxyStatus(p string) (isProxy bool, isEncoded bool) {
	// TODO:
	// If we don't use a slash here, then we could do a prefix check
	// in the calling code and pass a slice to this function (if
	// the original sequence is prefixed with a slash).
	const asciiHTTP = "http://"
	const asciiHTTPS = "https://"
	if strings.HasPrefix(p, asciiHTTP) || strings.HasPrefix(p, asciiHTTPS) {
		return true, false
	}
	const encodedHTTP = "http%3A%2F%2F"
	const encodedHTTPS = "https%3A%2F%2F"

	if strings.HasPrefix(p, encodedHTTP) || strings.HasPrefix(p, encodedHTTPS) {
		return true, true
	}
	return false, false
}

// encodeProxy will encode the given path string if it hasn't been
// encoded. If the path string isEncoded, then the path string is
// returned unchanged. Otherwise, the path is passed to PathEscape.
// The proxy-path is nearly escaped for our use-case after the call
// to PathEscape.
//
// Per net/url, PathEscape enters the shouldEscape function with the
// mode set to encodePathSegment. This means that COLON (:) will be
// considered unreserved and make it into the escaped path component.
// It also means that '/', ';', ',', and '?' will be escaped.
//
// See:
// https://golang.org/src/net/url/url.go?s=7851:7884#L137
func encodeProxy(proxyPath string, isEncoded bool) (escapedProxyPath string) {
	if isEncoded {
		return proxyPath
	}
	nearlyEscaped := url.PathEscape(proxyPath)
	escapedProxyPath = strings.Replace(nearlyEscaped, ":", "%3A", -1)
	return escapedProxyPath
}

func encodePath(p string) string {
	return url.PathEscape(p)
}

func encodeQueryParamsFromValues(params url.Values) (encodedParams []string) {

	keys := make([]string, 0, len(params))

	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		encodedKey, encodedValue := encodeQueryParam(k, params.Get(k))
		encodedPairStr := joinQueryPair(encodedKey, encodedValue)
		encodedParams = append(encodedParams, encodedPairStr)
	}
	return encodedParams
}

func joinQueryPair(key string, value string) string {
	return strings.Join([]string{key, value}, "=")
}

func encodeQueryParam(key string, value string) (eK string, eV string) {
	eK = encodeQueryParamValue(key)

	if isBase64(key) {
		eV = base64EncodeQueryParamValue(value)
		return eK, eV
	}

	eV = encodeQueryParamValue(value)
	return eK, eV
}

// encodeQueryParamValue encodes a query parameter value by first
// replacing all PLUS (+) characters with their escaped form, '%2B'.
// The value with escaped PLUS signs is passed to QueryEscape
// which escapes "everything." This function aggressively escapes PLUS
// (+) as SPACE and substitutes '%20' (for PLUS) and this is why we
// attempt to preserve PLUS (+) characters first, then escape the
// query, and then go back through and replace all the PLUS (+) signs
// which Go's net/url module prefers.
//
// See:
// https://golang.org/src/net/url/url.go?s=7851:7884#L149
func encodeQueryParamValue(queryValue string) string {
	escapedPlus := strings.Replace(queryValue, "+", "%2B", -1)
	queryEscaped := url.QueryEscape(escapedPlus)
	fullyEscaped := strings.Replace(queryEscaped, "+", "%20", -1)
	return fullyEscaped
}

// isBase64 checks if the paramKey is suffixed by "64," indicating
// that the value is intended to be base64-URL-encoded.
func isBase64(paramKey string) bool {
	return strings.HasSuffix(paramKey, "64")
}

// base64EncodeQueryParamValue base64 encodes the queryValue string. It
// does so in accordance with RFC 4648, which obsoletes RFC 3548. The
// important points are that the diff isn't significant for anything
// we care about.
//
// See:
// https://tools.ietf.org/rfcdiff?url2=rfc4648
func base64EncodeQueryParamValue(queryValue string) string {
	maybePaddedValue := base64.URLEncoding.EncodeToString([]byte(queryValue))
	return unPad(maybePaddedValue)
}

func unPad(s string) string {
	if strings.HasSuffix(s, "=") {
		return strings.Replace(s, "=", "", -1)
	}
	return s
}
