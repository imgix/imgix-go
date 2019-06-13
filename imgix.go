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

// Matches http:// and https://
var RegexpHTTPAndS = regexp.MustCompile("https?://")

// Regexp for all characters we should escape in a URI passed in.
var RegexUrlCharactersToEscape = regexp.MustCompile("([^ a-zA-Z0-9_.-])")

// Create a new Client with the given domain, with HTTPS enabled.
func NewClient(domain string) Client {
	return Client{domain: domain, secure: true}
}

// Create a new Client with the given domain and token. HTTPS enabled.
func NewClientWithToken(domain string, token string) Client {
	return Client{domain: domain, secure: true, token: token}
}

// The Client is used to build URLs.
type Client struct {
	domain string
	token  string
	secure bool
}

// Returns whether HTTPS should be used.
func (c *Client) Secure() bool {
	return c.secure
}

// Returns the URL scheme to use. One of 'http' or 'https'.
func (c *Client) Scheme() string {
	if c.Secure() {
		return "https"
	} else {
		return "http"
	}
}

// Returns the domain for the given path.
func (c *Client) Domain() string {
	return RegexpHTTPAndS.ReplaceAllString(c.domain, "") // Strips out the scheme if exists
}

// Creates and returns the URL signature in the form of "s=SIGNATURE" with
// no values.
func (c *Client) SignatureForPath(path string) string {
	return c.SignatureForPathAndParams(path, url.Values{})
}

// Creates and returns the URL signature in the form of "s=SIGNATURE" for
// the given parameters. Requires that the client have a token.
func (c *Client) SignatureForPathAndParams(path string, params url.Values) string {
	if c.token == "" {
		return ""
	}

	hasher := md5.New()
	hasher.Write([]byte(c.token + path))

	// Do not mix in the parameters into the signature hash if no parameters
	// have been given
	if len(params) != 0 {
		hasher.Write([]byte("?" + params.Encode()))
	}

	return "s=" + hex.EncodeToString(hasher.Sum(nil))
}

// Builds the full URL to the image (including the domain) with no params.
func (c *Client) Path(imgPath string) string {
	return c.PathWithParams(imgPath, url.Values{})
}

// `PathWithParams` will manually build a URL from a given path string and
// parameters passed in. Because of the differences in how net/url escapes
// path components, we need to manually build a URL as best we can.
//
// The behavior of this function is highly dependent upon its test suite.
func (c *Client) PathWithParams(imgPath string, params url.Values) string {
	u := url.URL{
		Scheme: c.Scheme(),
		Host:   c.Domain(),
	}

	urlString := u.String()

	// If we are given a fully-qualified URL, escape it per the note located
	// near the `cgiEscape` function definition
	if RegexpHTTPAndS.MatchString(imgPath) {
		imgPath = cgiEscape(imgPath)
	}

	// Add a leading slash if one does not exist:
	//     "users/1.png" -> "/users/1.png"
	if strings.Index(imgPath, "/") != 0 {
		imgPath = "/" + imgPath
	}

	urlString += imgPath

	for key, val := range params {
		if strings.HasSuffix(key, "64") {
			encodedParam := base64EncodeParameter(val[0])
			params.Set(key, encodedParam)
		}
	}

	// The signature in an imgix URL must always be the **last** parameter in a URL,
	// hence some of the gross string concatenation here. net/url will aggressively
	// alphabetize the URL parameters.
	signature := c.SignatureForPathAndParams(imgPath, params)
	parameterString := params.Encode()
	parameterString = strings.Replace(parameterString, "+", "%%20", -1)

	if signature != "" && len(params) > 0 {
		parameterString += "&" + signature
	} else if signature != "" && len(params) == 0 {
		parameterString = signature
	}

	// Only append the parameter string if it is not blank.
	if parameterString != "" {
		urlString += "?" + parameterString
	}

	return urlString
}

// Base64-encodes a parameter according to imgix's Base64 variant requirements.
// https://docs.imgix.com/apis/url#base64-variants
func base64EncodeParameter(param string) string {
	paramData := []byte(param)
	base64EncodedParam := base64.URLEncoding.EncodeToString(paramData)
	base64EncodedParam = strings.Replace(base64EncodedParam, "=", "", -1)

	return base64EncodedParam
}

// This code is less than ideal, but it's the only way we've found out how to do it
// give Go's URL capabilities and escaping behavior.
//
// This method replicates the beavhior of Ruby's CGI::escape in Go.
//
// Here is that method:
//
//     def CGI::escape(string)
//       string.gsub(/([^ a-zA-Z0-9_.-]+)/) do
//         '%' + $1.unpack('H2' * $1.bytesize).join('%').upcase
//       end.tr(' ', '+')
//      end
//
// It replaces
//
// See:
//  - https://github.com/parkr/imgix-go/pull/1#issuecomment-109014369
//  - https://github.com/imgix/imgix-blueprint#securing-urls
func cgiEscape(s string) string {
	return RegexUrlCharactersToEscape.ReplaceAllStringFunc(s, func(s string) string {
		rune, _ := utf8.DecodeLastRuneInString(s)
		return "%" + strings.ToUpper(fmt.Sprintf("%x", rune))
	})
}
