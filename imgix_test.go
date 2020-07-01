package imgix

import (
	"encoding/base64"
	"log"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// TODO: Idiomatic testing.
func testClient() URLBuilder {
	return NewURLBuilder("test.imgix.net")
}

func testClientWithToken() URLBuilder {
	return NewURLBuilderWithToken("my-social-network.imgix.net", "FOO123bar")
}

func TestBasicClientPath(t *testing.T) {
	c := testClient()
	assert.Equal(t, "https://test.imgix.net/1/users.jpg", c.CreateURLFromPath("/1/users.jpg"))
}

func TestClientPathWithParams(t *testing.T) {
	c := testClient()
	params := url.Values{"w": []string{"200"}, "h": []string{"400"}}
	assert.Equal(t, "https://test.imgix.net/1/users.jpg?h=400&w=200", c.CreateURL("/1/users.jpg", params))
}

func TestClientScheme(t *testing.T) {
	c := testClient()
	c.useHTTPS = false
	assert.Equal(t, "http", c.Scheme())
	c.useHTTPS = true
	assert.Equal(t, "https", c.Scheme())
}

func TestClientPath(t *testing.T) {
	c := testClient()
	u := c.CreateURLFromPath("/jax.jpg")
	assert.Equal(t, "https://test.imgix.net/jax.jpg", u)
}

func TestClientPathWithOneParam(t *testing.T) {
	c := testClient()
	params := url.Values{"w": []string{"400"}}
	u := c.CreateURL("/jax.jpg", params)
	assert.Equal(t, "https://test.imgix.net/jax.jpg?w=400", u)
}

func TestClientPathWithTwoParams(t *testing.T) {
	c := testClient()
	params := url.Values{"w": []string{"400"}, "h": []string{"300"}}
	u := c.CreateURL("/jax.jpg", params)
	assert.Equal(t, "https://test.imgix.net/jax.jpg?h=300&w=400", u)
}

func TestClientPathWithParamsEncodesParamKeys(t *testing.T) {
	c := testClient()
	params := url.Values{"hello world": []string{"interesting"}}
	u := c.CreateURL("/demo.png", params)
	assert.Equal(t, "https://test.imgix.net/demo.png?hello%20world=interesting", u)
}

func TestClientPathWithParamsEncodesParamValues(t *testing.T) {
	c := testClient()
	params := url.Values{"hello_world": []string{"/foo\"> <script>alert(\"hacked\")</script><"}}
	u := c.CreateURL("/demo.png", params)
	assert.Equal(t, "https://test.imgix.net/demo.png?hello_world=%2Ffoo%22%3E%20%3Cscript%3Ealert%28%22hacked%22%29%3C%2Fscript%3E%3C", u)
}

func TestClientPathWithParamsEncodesBase64ParamVariants(t *testing.T) {
	c := testClient()
	params := url.Values{"txt64": []string{"I cannøt belîév∑ it wors! 😱"}}
	u := c.CreateURL("~text", params)
	assert.Equal(t, "https://test.imgix.net/~text?txt64=SSBjYW5uw7h0IGJlbMOuw6l24oiRIGl0IHdvcu-jv3MhIPCfmLE", u)
}

func TestClientPathWithSignature(t *testing.T) {
	c := testClientWithToken()
	u := c.CreateSignedURLFromPath("/users/1.png")
	assert.Equal(t, "https://my-social-network.imgix.net/users/1.png?s=6797c24146142d5b40bde3141fd3600c", u)
}

func TestClientPathWithSignatureAndParams(t *testing.T) {
	c := testClientWithToken()
	params := url.Values{"w": []string{"400"}, "h": []string{"300"}}
	assert.Equal(t, "https://my-social-network.imgix.net/users/1.png?h=300&w=400&s=9de08f728192f92b7132176a0a17ef08", c.CreateSignedURL("/users/1.png", params))
}

func TestClientPathWithSignatureAndEmptyParams(t *testing.T) {
	c := testClientWithToken()
	params := url.Values{}
	assert.Equal(t, "https://my-social-network.imgix.net/users/1.png?s=6797c24146142d5b40bde3141fd3600c", c.CreateSignedURL("/users/1.png", params))
}

func TestClientFullyQualifiedUrlPath(t *testing.T) {
	c := testClientWithToken()
	assert.Equal(t, "https://my-social-network.imgix.net/http%3A%2F%2Favatars.com%2Fjohn-smith.png?s=493a52f008c91416351f8b33d4883135", c.CreateSignedURLFromPath("http://avatars.com/john-smith.png"))
}

func TestClientFullyQualifiedUrlPathWithParams(t *testing.T) {
	c := testClientWithToken()
	params := url.Values{"w": []string{"400"}, "h": []string{"300"}}
	assert.Equal(t, "https://my-social-network.imgix.net/http%3A%2F%2Favatars.com%2Fjohn-smith.png?h=300&w=400&s=a58d87a5dfa5f7478e06715571e96f78", c.CreateSignedURL("http://avatars.com/john-smith.png", params))
}

func TestClientHostsCountValidation(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(error)
			assert.True(t, ok)
			assert.EqualError(t, e, "hosts must be provided")
		}
	}()

	c := testClient()
	c.domain = string("")
	c.Domain()
}

func TestURLBuilder_CreateSrcSetFromWidths(t *testing.T) {
	c := testClient()
	actual := c.CreateSrcSetFromWidths("image.jpg", url.Values{}, []int{100, 200, 300, 400})
	expected := "https://test.imgix.net/image.jpg?w=100 100w,\n" +
		"https://test.imgix.net/image.jpg?w=200 200w,\n" +
		"https://test.imgix.net/image.jpg?w=300 300w,\n" +
		"https://test.imgix.net/image.jpg?w=400 400w"
	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcSetFromRange(t *testing.T) {
	c := testClient()
	// For demonstration, the below is a longer version of the actual call:
	// c.CreateSrcSetFromRange("image.png", url.Values{}, WidthRange{begin: 100, end: 380, tol: 0.08})
	actual := c.CreateSrcSetFromRange("image.png", url.Values{}, WidthRange{100, 380, 0.08})
	expected := "https://test.imgix.net/image.png?w=100 100w,\n" +
		"https://test.imgix.net/image.png?w=116 116w,\n" +
		"https://test.imgix.net/image.png?w=135 135w,\n" +
		"https://test.imgix.net/image.png?w=156 156w,\n" +
		"https://test.imgix.net/image.png?w=181 181w,\n" +
		"https://test.imgix.net/image.png?w=210 210w,\n" +
		"https://test.imgix.net/image.png?w=244 244w,\n" +
		"https://test.imgix.net/image.png?w=283 283w,\n" +
		"https://test.imgix.net/image.png?w=328 328w,\n" +
		"https://test.imgix.net/image.png?w=380 380w"
	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcSetFixedW(t *testing.T) {
	c := testClient()
	params := url.Values{"w": []string{"320"}}
	options := SrcSetOpts{disableVariableQuality: false}
	expected := "https://test.imgix.net/image.png?dpr=1&q=75&w=320 1x,\n" +
		"https://test.imgix.net/image.png?dpr=2&q=50&w=320 2x,\n" +
		"https://test.imgix.net/image.png?dpr=3&q=35&w=320 3x,\n" +
		"https://test.imgix.net/image.png?dpr=4&q=23&w=320 4x,\n" +
		"https://test.imgix.net/image.png?dpr=5&q=20&w=320 5x"
	actual := c.CreateSrcSet("image.png", params, options)
	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcSetFixedHandAR(t *testing.T) {
	c := testClient()
	params := url.Values{"h": []string{"320"}, "ar": []string{"4:3"}}
	options := SrcSetOpts{disableVariableQuality: false}
	// TODO: it would appear that go's `url.Values` does aggressively
	// alphabetize query parameters...
	expected := "https://test.imgix.net/image.png?ar=4%3A3&dpr=1&h=320&q=75 1x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=2&h=320&q=50 2x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=3&h=320&q=35 3x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=4&h=320&q=23 4x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=5&h=320&q=20 5x"
	actual := c.CreateSrcSet("image.png", params, options)
	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcSetFluidHighTol(t *testing.T) {
	c := testClient()
	wr := WidthRange{100, 8192, 1000.0}
	options := SrcSetOpts{widthRange: wr}

	expected := "https://test.imgix.net/image.png?w=100 100w,\n" +
		"https://test.imgix.net/image.png?w=8192 8192w"

	actual := c.CreateSrcSet("image.png", url.Values{}, options)
	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcSetFluidWidth100to108at2percent(t *testing.T) {
	c := testClient()
	wr := WidthRange{100, 108, 0.02}
	config := SrcSetOpts{widthRange: wr}

	expected := "https://test.imgix.net/image.png?w=100 100w,\n" +
		"https://test.imgix.net/image.png?w=104 104w,\n" +
		"https://test.imgix.net/image.png?w=108 108w"

	actual := c.CreateSrcSet("image.png", url.Values{}, config)
	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcSetQoverridesDisableVarQuality(t *testing.T) {
	c := testClient()
	params := url.Values{"h": []string{"800"}, "ar": []string{"4:3"}, "q": []string{"99"}}
	options := SrcSetOpts{disableVariableQuality: true}
	expected := "https://test.imgix.net/image.png?ar=4%3A3&dpr=1&h=800&q=99 1x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=2&h=800&q=99 2x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=3&h=800&q=99 3x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=4&h=800&q=99 4x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=5&h=800&q=99 5x"
	actual := c.CreateSrcSet("image.png", params, options)
	assert.Equal(t, expected, actual)
}

func TestValidators_validateNegativeWidths(t *testing.T) {
	widths := []int{100, 200, 300, -400, -500}
	validWidths, err := validateWidths(widths)

	// Ensure an error occurred, and the `err` is `NotEqual` to `nil`.
	assert.NotEqual(t, nil, err)
	assert.Equal(t, []int{}, validWidths)
}

func TestValidators_validatePositiveWidths(t *testing.T) {
	expected := []int{101, 202, 303, 404, 505}
	validWidths, err := validateWidths(expected)

	// Check the `err` is nil.
	assert.Equal(t, nil, err)
	// Check the expected widths are valid widths.
	assert.Equal(t, expected, validWidths)
}

func TestValidators_validateMinWidthValid(t *testing.T) {
	const OneHundred = 100
	validValue, err := validateMinWidth(OneHundred)
	assert.Equal(t, OneHundred, validValue)
	assert.Equal(t, nil, err)
}

func TestValidators_validateMinWidthInvalid(t *testing.T) {
	const LessThanZero = -1
	invalidValue, err := validateMinWidth(LessThanZero)
	assert.Equal(t, -1, invalidValue)
	assert.NotEqual(t, nil, err)
}

func TestValidators_validateMaxWidthValid(t *testing.T) {
	const OneHundred = 100
	validValue, err := validateMaxWidth(OneHundred)
	assert.Equal(t, OneHundred, validValue)
	assert.Equal(t, nil, err)
}

func TestValidators_validateMaxWidthInvalid(t *testing.T) {
	const LessThanZero = -1
	invalidValue, err := validateMaxWidth(LessThanZero)
	assert.Equal(t, -1, invalidValue)
	assert.NotEqual(t, nil, err)
}

func TestValidators_validateRangeInvalid(t *testing.T) {
	begin := 740
	end := 320

	_, err := validateRange(begin, end)
	assert.NotEqual(t, nil, err)
}

func TestValidators_validateRangeValid(t *testing.T) {
	rp := rangePair{begin: 100, end: 8192}
	validRangePair, err := validateRange(rp.begin, rp.end)
	assert.Equal(t, rp, validRangePair)
	assert.Equal(t, nil, err)
}

func TestValidators_validateRangeWithToleranceInvalid(t *testing.T) {
	invalidTolerance := 0.001
	_, err := validateRangeWithTolerance(100, 200, invalidTolerance)
	assert.NotEqual(t, nil, err)
}

func TestValidators_validateRangeWithToleranceValid(t *testing.T) {
	invalidTolerance := 1.25
	_, err := validateRangeWithTolerance(100, 200, invalidTolerance)
	assert.Equal(t, nil, err)
}

func TestReadMe_basicURLUsage(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net")
	expected := "https://demo.imgix.net/path/to/image.jpg"
	actual := ub.CreateURL("path/to/image.jpg", url.Values{})
	assert.Equal(t, expected, actual)
}

func TestReadMe_basicURLUsageHandW100(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net")
	expected := "https://demo.imgix.net/path/to/image.jpg?h=100&w=100"
	actual := ub.CreateURL("path/to/image.jpg", url.Values{"h": []string{"100"}, "w": []string{"100"}})
	assert.Equal(t, expected, actual)
}

func TestReadMe_basicURLUsageUsingHttp(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net")
	// Set the UseHttps field to false to begin using HTTP.
	ub.SetUseHTTPS(false)

	expected := "http://demo.imgix.net/path/to/image.jpg"
	actual := ub.CreateURL("path/to/image.jpg", url.Values{})
	assert.Equal(t, expected, actual)
}

func TestReadMe_basicURLUsageSigningWithToken(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ixToken := os.Getenv("IX_TOKEN")
	ub := NewURLBuilderWithToken("demo.imgix.net", ixToken)

	expected := "https://demo.imgix.net/path/to/image.jpg?s=c8bd1807209f7f1d96dd7123f92febb4"
	actual := ub.CreateSignedURL("path/to/image.jpg", url.Values{})
	assert.Equal(t, expected, actual)
}

// TODO: Think harder about how signing occurs.
func TestReadMe_SignedSrcSetCreation(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ixToken := os.Getenv("IX_TOKEN")
	ub := NewURLBuilderWithToken("demos.imgix.net", ixToken)
	srcset := ub.CreateSrcSet("image.png", url.Values{}, DefaultOpts)

	expectedLength := 31
	splitSrcSet := strings.Split(srcset, ",\n")

	for _, u := range splitSrcSet {
		assert.Contains(t, u, "s=")
	}

	actualLength := len(splitSrcSet)
	assert.Equal(t, expectedLength, actualLength)
}

func TestReadMe_FixedWidthSrcSetDefault(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net")
	params := url.Values{"h": []string{"800"}, "ar": []string{"4:3"}}
	options := SrcSetOpts{disableVariableQuality: false}
	expected := "https://demo.imgix.net/image.png?ar=4%3A3&dpr=1&h=800&q=75 1x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=2&h=800&q=50 2x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=3&h=800&q=35 3x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=4&h=800&q=23 4x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=5&h=800&q=20 5x"
	actual := ub.CreateSrcSet("image.png", params, options)
	assert.Equal(t, expected, actual)
}

func TestReadMe_FixedWidthSrcSetVariableQualityDisabled(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net")
	params := url.Values{"h": []string{"800"}, "ar": []string{"4:3"}}
	options := SrcSetOpts{disableVariableQuality: true}
	expected := "https://demo.imgix.net/image.png?ar=4%3A3&dpr=1&h=800 1x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=2&h=800 2x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=3&h=800 3x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=4&h=800 4x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=5&h=800 5x"
	actual := ub.CreateSrcSet("image.png", params, options)
	assert.Equal(t, expected, actual)
}

func TestReadMe_FixedWidthSrcSetNoOpts(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net")
	params := url.Values{"h": []string{"800"}, "ar": []string{"4:3"}}
	expected := "https://demo.imgix.net/image.png?ar=4%3A3&dpr=1&h=800&q=75 1x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=2&h=800&q=50 2x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=3&h=800&q=35 3x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=4&h=800&q=23 4x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=5&h=800&q=20 5x"
	actual := ub.CreateSrcSet("image.png", params, SrcSetOpts{})
	assert.Equal(t, expected, actual)
}

func TestReadMe_FluidWidthSrcSetFromWidths(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net")
	actual := ub.CreateSrcSetFromWidths("image.jpg", url.Values{}, []int{100, 200, 300, 400})
	expected := "https://demo.imgix.net/image.jpg?w=100 100w,\n" +
		"https://demo.imgix.net/image.jpg?w=200 200w,\n" +
		"https://demo.imgix.net/image.jpg?w=300 300w,\n" +
		"https://demo.imgix.net/image.jpg?w=400 400w"
	assert.Equal(t, expected, actual)
}

func TestReadMe_FluidWidthSrcSetFromSrcSetOpts(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net")
	options := SrcSetOpts{widthRange: WidthRange{begin: 100, end: 380, tol: 0.08}}
	actual := ub.CreateSrcSet("image.png", url.Values{}, options)
	expected := "https://demo.imgix.net/image.png?w=100 100w,\n" +
		"https://demo.imgix.net/image.png?w=116 116w,\n" +
		"https://demo.imgix.net/image.png?w=135 135w,\n" +
		"https://demo.imgix.net/image.png?w=156 156w,\n" +
		"https://demo.imgix.net/image.png?w=181 181w,\n" +
		"https://demo.imgix.net/image.png?w=210 210w,\n" +
		"https://demo.imgix.net/image.png?w=244 244w,\n" +
		"https://demo.imgix.net/image.png?w=283 283w,\n" +
		"https://demo.imgix.net/image.png?w=328 328w,\n" +
		"https://demo.imgix.net/image.png?w=380 380w"
	assert.Equal(t, expected, actual)
}

func TestReadMe_FluidWidthSrcSetFromSrcSetOptsTol20(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net")
	options := SrcSetOpts{widthRange: WidthRange{begin: 100, end: 384, tol: 0.20}}
	actual := ub.CreateSrcSet("image.png", url.Values{}, options)
	expected := "https://demo.imgix.net/image.png?w=100 100w,\n" +
		"https://demo.imgix.net/image.png?w=140 140w,\n" +
		"https://demo.imgix.net/image.png?w=196 196w,\n" +
		"https://demo.imgix.net/image.png?w=274 274w,\n" +
		"https://demo.imgix.net/image.png?w=384 384w"
	assert.Equal(t, expected, actual)
}

func TestReadMe_TargetWidths(t *testing.T) {
	expected := []int{300, 378, 476, 600, 756, 953, 1200, 1513, 1906, 2401, 3000}
	actual := TargetWidths(300, 3000, 0.13)
	assert.Equal(t, expected, actual)

	sm := expected[:3]
	expectedSm := []int{300, 378, 476}
	assert.Equal(t, expectedSm, sm)

	md := expected[3:7]
	expectedMd := []int{600, 756, 953, 1200}
	assert.Equal(t, expectedMd, md)

	lg := expected[7:]
	expectedLg := []int{1513, 1906, 2401, 3000}
	assert.Equal(t, expectedLg, lg)

	ub := NewURLBuilder("demos.imgix.net")
	srcset := ub.CreateSrcSetFromWidths("image.png", url.Values{}, sm)
	actualSrcset := "https://demos.imgix.net/image.png?w=300 300w,\n" +
		"https://demos.imgix.net/image.png?w=378 378w,\n" +
		"https://demos.imgix.net/image.png?w=476 476w"
	assert.Equal(t, actualSrcset, srcset)
}

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
