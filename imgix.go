package imgix

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
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
