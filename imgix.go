package imgix

import (
	"net/url"
	"strings"
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
	hasToken := b.token != ""
	return b.createAndMaybeSignURL(path, params, hasToken)
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

// maybePrependSlash prepends if the path does not begin with one:
// "users/1.png" -> "/users/1.png"
func maybePrependSlash(path string) string {
	const Slash = "/"

	if strings.Index(path, Slash) != 0 {
		path = strings.Join([]string{Slash, path}, "")
	}
	return path
}
