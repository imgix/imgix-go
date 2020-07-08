package imgix

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"unicode/utf8"
)

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
	// Rather than adding or removing a '/', check if `p` is prefixed
	// with a '/'. If so, use another slice with the leading '/'
	// removed.
	path := p
	if strings.HasPrefix(p, "/") {
		path = p[1:]
	}

	const asciiHTTP = "http://"
	const asciiHTTPS = "https://"
	if strings.HasPrefix(path, asciiHTTP) || strings.HasPrefix(path, asciiHTTPS) {
		return true, false
	}

	const encodedHTTP = "http%3A%2F%2F"
	const encodedHTTPS = "https%3A%2F%2F"
	if strings.HasPrefix(path, encodedHTTP) || strings.HasPrefix(path, encodedHTTPS) {
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
		encodedPairStr := strings.Join([]string{encodedKey, encodedValue}, "=")
		encodedParams = append(encodedParams, encodedPairStr)
	}
	return encodedParams
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

// unPad removes the extra '=' (equal signs) from strings. In base64,
// '=' are added to the end of the encoding as padding. This padding
// is significant if concatenating multiple base64-encoded strings.
// In our case, '&' acts as the primary delimeter and base64-encoded
// strings (query string values, usually) are dealt with individually
// (meaning that the length of the base64 encoded string is always
// known; this is important when decoding base64).
func unPad(s string) string {
	if strings.HasSuffix(s, "=") {
		return strings.Replace(s, "=", "", -1)
	}
	return s
}
