package imgix

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncoding_isBase64(t *testing.T) {
	assert.True(t, isBase64("64"))
	assert.True(t, isBase64("   64"))
	assert.True(t, isBase64("646464"))
	assert.True(t, isBase64("fit64"))
	assert.True(t, isBase64("markalign64"))
}

func TestEncoding_isNotBase64(t *testing.T) {
	assert.False(t, isBase64("6  4"))
	assert.False(t, isBase64("646464 "))
	assert.False(t, isBase64("\x40"))
}

func TestEncoding_base64EncodeQueryParamValue(t *testing.T) {
	const expectedWarmUp = "SGVsbG8sIOS4lueVjA"
	const data = "Hello, 世界"
	actualWarmUp := base64EncodeQueryParamValue(data)
	assert.Equal(t, expectedWarmUp, actualWarmUp)

	const preEncoded = "Avenir Next Demi,Bold"
	const expectedAve = "QXZlbmlyIE5leHQgRGVtaSxCb2xk"
	actualAve := base64EncodeQueryParamValue("Avenir Next Demi,Bold")
	assert.Equal(t, expectedAve, actualAve)

	decodedAve, _ := base64.StdEncoding.DecodeString(actualAve)
	assert.Equal(t, preEncoded, string(decodedAve))
}

func TestEncoding_base64UTF8(t *testing.T) {
	s := `I cannøt belîév∑ it wors! 😱`
	actual := base64EncodeQueryParamValue(s)
	expected := "SSBjYW5uw7h0IGJlbMOuw6l24oiRIGl0IHdvcu-jv3MhIPCfmLE"
	assert.Equal(t, expected, actual)
}

func TestEncoding_BlueprintBase64(t *testing.T) {
	s := `Hello,+World!`
	actual := base64EncodeQueryParamValue(s)
	expected := "SGVsbG8sK1dvcmxkIQ"
	assert.Equal(t, expected, actual)
}

func TestEncoding_checkProxyStatusEmpty(t *testing.T) {
	isProxy, isEncoded := checkProxyStatus("")
	assert.Equal(t, false, isProxy)
	assert.Equal(t, false, isEncoded)
}

func TestEncoding_checkProxyStatusEncoded(t *testing.T) {
	const encodedProxy = "http%3A%2F%2Fwww.this.com%2Fpic.jpg"
	isProxy, isEncoded := checkProxyStatus(encodedProxy)
	assert.Equal(t, true, isProxy)
	assert.Equal(t, true, isEncoded)
}

func TestEncoding_checkProxyStatusAscii(t *testing.T) {

	const proxyHTTP = "http://www.this.com/pic.jpg"
	isProxyHTTP, isEncodedHTTP := checkProxyStatus(proxyHTTP)
	assert.Equal(t, true, isProxyHTTP)
	assert.Equal(t, false, isEncodedHTTP)

	const proxyHTTPS = "https://www.this.com/pic.jpg"
	isProxyHTTPS, isEncodedHTTPS := checkProxyStatus(proxyHTTPS)
	assert.Equal(t, true, isProxyHTTPS)
	assert.Equal(t, false, isEncodedHTTPS)
}

func TestEncoding_encodePathProxyEncoded(t *testing.T) {
	const expected = "/http%3A%2F%2Fwww.this.com%2Fpic.jpg"
	actual := processPath(expected)
	assert.Equal(t, expected, actual)
}

func TestEncoding_encodePathProxyRaw(t *testing.T) {
	const proxyPath = "http://www.this.com/pic.jpg"
	const expected = "/http%3A%2F%2Fwww.this.com%2Fpic.jpg"
	actual := processPath(proxyPath)

	assert.Equal(t, expected, actual)
}
