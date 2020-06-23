package imgix

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: Idiomatic testing.
func testClient() Builder {
	return NewBuilder("test.imgix.net")
}

func testClientWithToken() Builder {
	return NewBuilderWithToken("my-social-network.imgix.net", "FOO123bar")
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
	assert.Equal(t, "https://test.imgix.net/demo.png?hello%%20world=interesting", u)
}

func TestClientPathWithParamsEncodesParamValues(t *testing.T) {
	c := testClient()
	params := url.Values{"hello_world": []string{"/foo\"> <script>alert(\"hacked\")</script><"}}
	u := c.CreateURL("/demo.png", params)
	assert.Equal(t, "https://test.imgix.net/demo.png?hello_world=%2Ffoo%22%3E%%20%3Cscript%3Ealert%28%22hacked%22%29%3C%2Fscript%3E%3C", u)
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

func TestBuilder_CreateSrcSetFromWidths(t *testing.T) {
	c := testClient()
	actual := c.CreateSrcSetFromWidths("image.jpg", url.Values{}, []int{100, 200, 300, 400})
	expected := "https://test.imgix.net/image.jpg?w=100 100w,\n" +
		"https://test.imgix.net/image.jpg?w=200 200w,\n" +
		"https://test.imgix.net/image.jpg?w=300 300w,\n" +
		"https://test.imgix.net/image.jpg?w=400 400w"
	assert.Equal(t, expected, actual)
}

func TestBuilder_CreateSrcSetFromRange(t *testing.T) {
	c := testClient()
	actual := c.CreateSrcSetFromRange("image.png", url.Values{}, 100, 380, 0.08)
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

func TestValidators_validateNegativeWidths(t *testing.T) {
	widths := []int{100, 200, 300, -400, -500}
	validWidths, err := validateWidths(widths)

	// Check that an error occurred and that `err` is `NotEqual` to nil.
	assert.NotEqual(t, nil, err)
	assert.Equal(t, []int{}, validWidths)
}

func TestValidators_validatePositiveWidths(t *testing.T) {
	expected := []int{101, 202, 303, 404, 505}
	validWidths, err := validateWidths(expected)

	// Check that the `err` is nil.
	assert.Equal(t, nil, err)
	// Check that the expected widths are valid widths.
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
