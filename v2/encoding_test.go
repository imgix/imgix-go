package imgix

import (
	"encoding/base64"
	"testing"
)

func TestEncoding_isBase64(t *testing.T) {
	gotBase64 := isBase64("64")
	if !gotBase64 {
		t.Errorf("got:  %t; want: %t", gotBase64, true)
	}

	gotBase64 = isBase64("   64")
	if !gotBase64 {
		t.Errorf("got:  %t; want: %t", gotBase64, true)
	}

	gotBase64 = isBase64("646464")
	if !gotBase64 {
		t.Errorf("got:  %t; want: %t", gotBase64, true)
	}

	gotBase64 = isBase64("fit64")
	if !gotBase64 {
		t.Errorf("got:  %t; want: %t", gotBase64, true)
	}
	gotBase64 = isBase64("markalign64")
	if !gotBase64 {
		t.Errorf("got:  %t; want: %t", gotBase64, true)
	}
}

func TestEncoding_isNotBase64(t *testing.T) {
	// Ensure the following strings are NOT accepted as
	// valid base64 keys.
	gotBase64 := isBase64("6  4")
	if gotBase64 {
		t.Errorf("got:  %t; want: %t", gotBase64, true)
	}

	gotBase64 = isBase64("646464 ")
	if gotBase64 {
		t.Errorf("got:  %t; want: %t", gotBase64, true)
	}

	gotBase64 = isBase64("\x40")
	if gotBase64 {
		t.Errorf("got:  %t; want: %t", gotBase64, true)
	}
}

func TestEncoding_base64EncodeQueryParamValue(t *testing.T) {
	const data = "Hello, 世界"
	const wantHello64 = "SGVsbG8sIOS4lueVjA"
	gotHello64 := base64EncodeQueryParamValue(data)

	if gotHello64 != wantHello64 {
		t.Errorf("got:  %s; want: %s", gotHello64, wantHello64)
	}

	const wantAve64 = "QXZlbmlyIE5leHQgRGVtaSxCb2xk"
	const preEncodedAveStr = "Avenir Next Demi,Bold"

	gotAve64 := base64EncodeQueryParamValue(preEncodedAveStr)
	if gotAve64 != wantAve64 {
		t.Errorf("got:  %s; want: %s", gotAve64, wantAve64)
	}

	decodedAve, _ := base64.StdEncoding.DecodeString(gotAve64)

	// Check that we can reconstruct the original string from
	// the encoded string.
	gotDecoded := string(decodedAve)
	if gotDecoded != preEncodedAveStr {
		t.Errorf("got:  %s; want: %s", gotDecoded, preEncodedAveStr)
	}
}

func TestEncoding_base64UTF8(t *testing.T) {
	s := `I cannøt belîév∑ it wors! 😱`
	got := base64EncodeQueryParamValue(s)
	want := "SSBjYW5uw7h0IGJlbMOuw6l24oiRIGl0IHdvcu-jv3MhIPCfmLE"

	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestEncoding_BlueprintBase64(t *testing.T) {
	s := `Hello,+World!`
	got := base64EncodeQueryParamValue(s)
	want := "SGVsbG8sK1dvcmxkIQ"

	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestEncoding_checkProxyStatusEmpty(t *testing.T) {
	isProxy, isEncoded := checkProxyStatus("")
	const want = false

	if isProxy || isEncoded {
		t.Errorf("got:  isProxy == %t; want: isProxy == %t", isProxy, want)
		t.Errorf("got:  isEncoded == %t; want: isEncoded == %t", isEncoded, want)
	}
}

func TestEncoding_checkProxyStatusEncoded(t *testing.T) {
	const encodedProxy = "http%3A%2F%2Fwww.this.com%2Fpic.jpg"
	isProxy, isEncoded := checkProxyStatus(encodedProxy)

	const want = true
	if !(isProxy && isEncoded) {
		t.Errorf("got:  isProxy == %t; want: isProxy == %t", isProxy, want)
		t.Errorf("got:  isEncoded == %t; want: isEncoded == %t", isEncoded, want)
	}
}

func TestEncoding_checkProxyStatusAscii(t *testing.T) {
	const wantProxy = true
	const wantEncoded = false
	const proxyHTTP = "http://www.this.com/pic.jpg"
	isProxyHTTP, isEncodedHTTP := checkProxyStatus(proxyHTTP)

	if !isProxyHTTP {
		t.Errorf("got:  isProxyHTTP == %t; want: isProxyHTTP == %t", isProxyHTTP, wantProxy)
	}

	if isEncodedHTTP {
		t.Errorf("got:  isEncodedHTTP == %t; want: isEncodedHTTP == %t", isEncodedHTTP, wantEncoded)
	}

	const proxyHTTPS = "https://www.this.com/pic.jpg"
	isProxyHTTPS, isEncodedHTTPS := checkProxyStatus(proxyHTTPS)

	if !isProxyHTTPS {
		t.Errorf("got:  isProxyHTTPS == %t; want: isProxyHTTPS == %t", isProxyHTTPS, wantProxy)
	}

	if isEncodedHTTPS {
		t.Errorf("got:  isEncodedHTTPS == %t; want: isEncodedHTTPS == %t", isEncodedHTTPS, wantEncoded)
	}
}

func TestEncoding_encodePathProxyEncoded(t *testing.T) {
	const want = "/http%3A%2F%2Fwww.this.com%2Fpic.jpg"
	got := sanitizePath(want)

	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestEncoding_encodePathProxyRaw(t *testing.T) {
	const proxyPath = "http://www.this.com/pic.jpg"
	const want = "/http%3A%2F%2Fwww.this.com%2Fpic.jpg"
	got := sanitizePath(proxyPath)

	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}
