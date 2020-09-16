package imgix

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURL_DefaultBuilder(t *testing.T) {
	const domain = "test.imgix.net"
	u := NewURLBuilder(domain, map[string]string{})

	// Assert the builder uses HTTPS by default.
	assert.Equal(t, true, u.useHTTPS)

	// Assert the builder scheme is HTTPS by default.
	assert.Equal(t, "https", u.Scheme())

	// Assert the builder uses the lib param by default.
	assert.Equal(t, true, u.useLibParam)
}

func testBuilder() URLBuilder {
	u := NewURLBuilder("test.imgix.net", Opts{"useLibParam": "false"})
	return u
}

func TestURL_BasicPathNoParams(t *testing.T) {
	u := testBuilder()
	actual := u.CreateURL("image.png", url.Values{})
	expected := "https://test.imgix.net/image.png"
	assert.Equal(t, expected, actual)
}

func TestURL_BasicPathWithParams(t *testing.T) {
	u := testBuilder()
	actual := u.CreateURL("image.png", url.Values{"w": []string{"100"}})
	expected := "https://test.imgix.net/image.png?w=100"
	assert.Equal(t, expected, actual)
}

func TestURL_paramValuesAreEscaped(t *testing.T) {
	key := "hello_world"
	value := "/foo\"> <script>alert(\"hacked\")</script><"
	params := url.Values{key: []string{value}}

	u := testBuilder()
	actual := u.CreateURL("image.png", params)
	expected := "https://test.imgix.net/image.png?hello_world=%2Ffoo%22%3E+%3Cscript%3Ealert%28%22hacked%22%29%3C%2Fscript%3E%3C"
	assert.Equal(t, expected, actual)
}

func TestURL_PathsArePlusSafe(t *testing.T) {
	// https://github.com/imgix/imgix-core-js/issues/88
	u := testBuilder()
	expected := "https://test.imgix.net/E%2BP-003_D.jpeg"
	actual := u.CreateURL("E+P-003_D.jpeg", url.Values{})
	assert.Equal(t, expected, actual)
}

func TestURL_Base64WithUnicodeParam(t *testing.T) {
	u := testBuilder()
	params := url.Values{"txt64": []string{"I cannøt belîév∑ it wors! 😱"}}
	actual := u.CreateURL("~text", params)
	expected := "https://test.imgix.net/~text?txt64=SSBjYW5uw7h0IGJlbMOuw6l24oiRIGl0IHdvcu-jv3MhIPCfmLE"
	assert.Equal(t, expected, actual)
}

func TestURL_WithRepeatedParamValues(t *testing.T) {
	u := testBuilder()
	params := url.Values{"auto": []string{"format", "compress"}}
	expected := "https://test.imgix.net?auto=format%2Ccompress"
	actual := u.CreateURL("", params)
	assert.Equal(t, expected, actual)
}

func TestURL_BluePrintSigning(t *testing.T) {
	u := NewSecureURLBuilder(
		"my-social-network.imgix.net",
		"FOO123bar",
		Opts{})
	u.SetUseLibParam(false)
	expected := "https://my-social-network.imgix.net/http%3A%2F%2Favatars.com%2Fjohn-smith.png?s=493a52f008c91416351f8b33d4883135"
	actual := u.CreateURL("/http%3A%2F%2Favatars.com%2Fjohn-smith.png", url.Values{})
	assert.Equal(t, expected, actual)
}

func TestURL_BluePrintSigningWithParams(t *testing.T) {
	u := NewSecureURLBuilder(
		"my-social-network.imgix.net",
		"FOO123bar",
		Opts{"useLibParam": "false"})
	expected := "https://my-social-network.imgix.net/users/1.png?h=300&w=400&s=1a4e48641614d1109c6a7af51be23d18"
	params := url.Values{"h": []string{"300"}, "w": []string{"400"}}

	actualPathPrefixed := u.CreateURL("/users/1.png", params)
	assert.Equal(t, expected, actualPathPrefixed)

	// The only difference between this and the above is that
	// the below is not prefixed with a slash.
	actual := u.CreateURL("users/1.png", params)
	assert.Equal(t, expected, actual)
}

func TestURL_BluePrintSigningWithProblematicParams(t *testing.T) {
	// https://github.com/imgix/imgix-blueprint#base64url-encode-problematic-parameters
	u := testBuilder()
	expected := "https://test.imgix.net/image.png?mark64=aHR0cHM6Ly9hc3NldHMuaW1naXgubmV0L2xvZ28ucG5n"

	params := url.Values{"mark64": []string{"https://assets.imgix.net/logo.png"}}
	actual := u.CreateURL("image.png", params)
	assert.Equal(t, expected, actual)
}

func TestURL_SigningFullyQualifiedWithParams(t *testing.T) {
	u := NewSecureURLBuilder(
		"my-social-network.imgix.net",
		"FOO123bar",
		Opts{"useLibParam": "false"})
	expected := "https://my-social-network.imgix.net/http%3A%2F%2Favatars.com%2Fjohn-smith.png?h=300&w=400&s=a201fe1a3caef4944dcb40f6ce99e746"

	params := url.Values{"w": []string{"400"}, "h": []string{"300"}}
	actual := u.CreateURL("/http%3A%2F%2Favatars.com%2Fjohn-smith.png", params)
	assert.Equal(t, expected, actual)
}
