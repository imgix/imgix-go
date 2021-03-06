package imgix

import (
	"log"
	"net/url"
	"strings"
)

const ixLibVersion = "go-v2.0.2"

// URLBuilder facilitates the building of imgix URLs.
type URLBuilder struct {
	domain      string // A source's domain, e.g. example.imgix.net
	token       string // A source's secure token used to sign/secure URLs.
	useHTTPS    bool   // Denotes whether or not to use HTTPS.
	useLibParam bool   // Denotes whether or not to apply the ixLibVersion.
}

// BuilderOption provides a convenient interface for supplying URLBuilder
// options to the NewURLBuilder constructor. See WithToken, WithHTTPS, etc.
// for more details.
type BuilderOption func(b *URLBuilder)

// NewURLBuilder creates a new URLBuilder with the given domain, with HTTPS enabled.
func NewURLBuilder(domain string, options ...BuilderOption) URLBuilder {
	validDomain, err := validateDomain(domain)
	if err != nil {
		log.Fatal(err)
	}

	urlBuilder := URLBuilder{domain: validDomain, useHTTPS: true, useLibParam: true}

	for _, fn := range options {
		fn(&urlBuilder)
	}
	return urlBuilder
}

// WithToken returns a BuilderOption that NewURLBuilder consumes.
// The constructor uses this closure to set the URLBuilder's token
// attribute.
func WithToken(token string) BuilderOption {
	return func(b *URLBuilder) {
		b.token = token
	}
}

// WithHTTPS returns a BuilderOption that NewURLBuilder consumes.
// The constructor uses this closure to set the URLBuilder's useHTTPS
// attribute.
func WithHTTPS(useHTTPS bool) BuilderOption {
	return func(b *URLBuilder) {
		b.useHTTPS = useHTTPS
	}
}

// WithLibParam returns a BuilderOption that NewURLBuilder consumes.
// The constructor uses this closure to set the URLBuilder's useLibParam
// attribute.
func WithLibParam(useLibParam bool) BuilderOption {
	return func(b *URLBuilder) {
		b.useLibParam = useLibParam
	}
}

// UseHTTPS returns whether HTTPS or HTTP should be used.
func (b *URLBuilder) UseHTTPS() bool {
	return b.useHTTPS
}

// SetUseLibParam toggles the library param on and off. If useLibParam is set to
// true, the ixlib param will be toggled on. Otherwise, if useLibParam is set to
// false, the ixlib param will be toggled off and will not appear in the final URL.
func (b *URLBuilder) SetUseLibParam(useLibParam bool) {
	b.useLibParam = useLibParam
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

// Domain gets the builder's domain string.
func (b *URLBuilder) Domain() string {
	return b.domain
}

// SetToken sets the token for this builder. This value will be used to sign
// URLs created through the builder.
func (b *URLBuilder) SetToken(token string) {
	b.token = token
}

// IxParam seeks to improve the ergonomics of setting url.Values.
// For instance, without IxParam,  caller's would need to write:
// url.Values{"w": []string{"480"}, "auto": []string{"format", "compress"}}
// However, by employing this functional type we can write:
// []IxParam{Param("w", "480"), Param("auto", "format", "compress")}
type IxParam func(u *url.Values)

// Param accepts a key and a variable number of values. It returns a
// closure as an IxParam that, once called, will populate the url.Values
// structure. Note that values aren't added to the query parameters
// (url.Values) until this function is applied (e.g. in CreateURL).
func Param(k string, v ...string) IxParam {
	return func(u *url.Values) {
		for _, value := range v {
			u.Add(k, value)
		}
	}
}

// CreateURL creates a URL string given a path and a set of
// params.
func (b *URLBuilder) CreateURL(path string, params ...IxParam) string {
	urlParams := url.Values{}

	for _, fn := range params {
		fn(&urlParams)
	}

	scheme := b.Scheme()
	domain := b.Domain()
	path = sanitizePath(path)
	query := b.buildQueryString(urlParams)
	signature := b.sign(path, query)

	url := scheme + "://" + domain + path

	// If the query and signature are empty, return the url.
	if query == "" && signature == "" {
		return url
	}

	// If the signature is empty, but the query is not,
	// return the url with the query appended.
	if query != "" && signature == "" {
		return url + "?" + query
	}

	// If the query is empty, but the signature is not,
	// return the url with the signature appended.
	if query == "" && signature != "" {
		return url + "?" + signature
	}

	// If neither query nor signature is empty, append the
	// query, then append the signature.
	if query != "" && signature != "" {
		url += "?" + query + "&" + signature
	}

	return url
}

// createURLFromValues functions like CreateURL except that
// it accepts url.Values.
func (b *URLBuilder) createURLFromValues(path string, params url.Values) string {
	scheme := b.Scheme()
	domain := b.Domain()
	path = sanitizePath(path)
	query := b.buildQueryString(params)
	signature := b.sign(path, query)

	url := scheme + "://" + domain + path

	// If the query and signature are empty, return the url.
	if query == "" && signature == "" {
		return url
	}

	// If the signature is empty, but the query is not,
	// return the url with the query appended.
	if query != "" && signature == "" {
		return url + "?" + query
	}

	// If the query is empty, but the signature is not,
	// return the url with the signature appended.
	if query == "" && signature != "" {
		return url + "?" + signature
	}

	// If neither query nor signature is empty, append the
	// query, then append the signature.
	if query != "" && signature != "" {
		url += "?" + query + "&" + signature
	}

	return url
}

func (b *URLBuilder) buildQueryString(params url.Values) string {
	var encodedQueryParts []string
	if b.useLibParam {
		params.Set("ixlib", ixLibVersion)
	}
	encodedQueryParts = encodeQuery(params)
	return strings.Join(encodedQueryParts, "&")
}

func (b *URLBuilder) sign(path string, query string) string {
	if b.token == "" {
		return ""
	}

	signature := createMd5Signature(b.token, path, query)
	return strings.Join([]string{"s=", signature}, "")
}

// processPath processes a path string into a form that can be
// safely used in a URL path segment.
func sanitizePath(path string) string {
	if path == "" {
		return path
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	isProxy, isEncoded := checkProxyStatus(path)

	if isProxy {
		return encodeProxy(path, isEncoded)
	}
	return encodePath(path)
}
