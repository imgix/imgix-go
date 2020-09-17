package imgix

import (
	"log"
	"net/url"
	"strings"
)

const ixLibVersion = "go-v2.0.0"

// URLBuilder facilitates the building of URLs.
type URLBuilder struct {
	domain      string
	token       string
	useHTTPS    bool
	useLibParam bool
}

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

func WithToken(token string) BuilderOption {
	return func(b *URLBuilder) {
		b.token = token
	}
}

func WithHTTPS(useHTTPS bool) BuilderOption {
	return func(b *URLBuilder) {
		b.useHTTPS = useHTTPS
	}
}

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

// CreateURL creates a URL string given a path and a set of
// params.
func (b *URLBuilder) CreateURL(path string, params url.Values) string {
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
		params.Add("ixlib", ixLibVersion)
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
