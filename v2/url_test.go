package imgix

import (
	"testing"
)

func TestURL_DefaultBuilder(t *testing.T) {
	const domain = "test.imgix.net"
	u := NewURLBuilder(domain)

	// Assert the builder uses HTTPS by default.
	gotBool := u.useHTTPS
	if gotBool != true {
		t.Errorf("useHTTPS\ngot: %t; want: true", gotBool)
	}

	// Assert the builder scheme is HTTPS by default.
	gotStr := u.Scheme()
	if gotStr != "https" {
		t.Errorf("Scheme\ngot: %s; want: https", gotStr)
	}

	// Assert the builder uses the lib param by default.
	gotLibBool := u.useLibParam
	if gotLibBool != true {
		t.Errorf("useLibParam\ngot: %t; want: true", gotBool)
	}
}

func testBuilder() URLBuilder {
	u := NewURLBuilder("test.imgix.net", WithLibParam(false))
	return u
}

func TestURL_BasicPathNoParams(t *testing.T) {
	u := testBuilder()
	got := u.CreateURL("image.png")
	want := "https://test.imgix.net/image.png"
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestURL_BasicPathWithParams(t *testing.T) {
	u := testBuilder()

	got := u.CreateURL("image.png", Param("w", "100"))
	want := "https://test.imgix.net/image.png?w=100"
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestURL_paramValuesAreEscaped(t *testing.T) {
	key := "hello_world"
	value := "/foo\"> <script>alert(\"hacked\")</script><"
	u := testBuilder()
	got := u.CreateURL("image.png", Param(key, value))
	want := "https://test.imgix.net/image.png?hello_world=%2Ffoo%22%3E+%3Cscript%3Ealert%28%22hacked%22%29%3C%2Fscript%3E%3C"
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestURL_PathsArePlusSafe(t *testing.T) {
	// https://github.com/imgix/imgix-core-js/issues/88
	u := testBuilder()
	got := "https://test.imgix.net/E%2BP-003_D.jpeg"
	want := u.CreateURL("E+P-003_D.jpeg")
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestURL_Base64WithUnicodeParam(t *testing.T) {
	u := testBuilder()
	got := u.CreateURL("~text", Param("txt64", "I cannøt belîév∑ it wors! 😱"))
	want := "https://test.imgix.net/~text?txt64=SSBjYW5uw7h0IGJlbMOuw6l24oiRIGl0IHdvcu-jv3MhIPCfmLE"
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestURL_WithRepeatedParamValues(t *testing.T) {
	u := testBuilder()
	want := "https://test.imgix.net?auto=format%2Ccompress"
	got := u.CreateURL("", Param("auto", "format", "compress"))
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestURL_BluePrintSigning(t *testing.T) {
	u := NewURLBuilder("my-social-network.imgix.net", WithToken("FOO123bar"))
	u.SetUseLibParam(false)
	want := "https://my-social-network.imgix.net/http%3A%2F%2Favatars.com%2Fjohn-smith.png?s=493a52f008c91416351f8b33d4883135"
	got := u.CreateURL("/http%3A%2F%2Favatars.com%2Fjohn-smith.png")
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestURL_BluePrintSigningWithParams(t *testing.T) {
	u := NewURLBuilder(
		"my-social-network.imgix.net",
		WithToken("FOO123bar"),
		WithLibParam(false))

	want := "https://my-social-network.imgix.net/users/1.png?h=300&w=400&s=1a4e48641614d1109c6a7af51be23d18"
	params := []IxParam{Param("h", "300"), Param("w", "400")}
	gotPathPrefixed := u.CreateURL("/users/1.png", params...)
	if gotPathPrefixed != want {
		t.Errorf("\ngot:  %s\nwant: %s", gotPathPrefixed, want)
	}

	// The only difference between this and the above is that
	// the below is not prefixed with a slash.
	got := u.CreateURL("users/1.png", params...)
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestURL_BluePrintSigningWithProblematicParams(t *testing.T) {
	// https://github.com/imgix/imgix-blueprint#base64url-encode-problematic-parameters
	u := testBuilder()
	want := "https://test.imgix.net/image.png?mark64=aHR0cHM6Ly9hc3NldHMuaW1naXgubmV0L2xvZ28ucG5n"

	params := []IxParam{Param("mark64", "https://assets.imgix.net/logo.png")}
	got := u.CreateURL("image.png", params...)
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestURL_SigningFullyQualifiedWithParams(t *testing.T) {
	u := NewURLBuilder(
		"my-social-network.imgix.net",
		WithToken("FOO123bar"),
		WithLibParam(false))
	want := "https://my-social-network.imgix.net/http%3A%2F%2Favatars.com%2Fjohn-smith.png?h=300&w=400&s=a201fe1a3caef4944dcb40f6ce99e746"

	params := []IxParam{Param("w", "400"), Param("h", "300")}
	got := u.CreateURL("/http%3A%2F%2Favatars.com%2Fjohn-smith.png", params...)
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}
