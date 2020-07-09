package imgix

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testClient() URLBuilder {
	return NewURLBuilder("test.imgix.net")
}

func testClientWithToken() URLBuilder {
	return NewSecureURLBuilder("my-social-network.imgix.net", "FOO123bar")
}

func TestURLBuilder_CreateSrcSetFromWidths(t *testing.T) {
	c := testClient()
	c.SetUseLibParam(false)
	actual := c.CreateSrcSetFromWidths("image.jpg", url.Values{}, []int{100, 200, 300, 400})
	expected := "https://test.imgix.net/image.jpg?w=100 100w,\n" +
		"https://test.imgix.net/image.jpg?w=200 200w,\n" +
		"https://test.imgix.net/image.jpg?w=300 300w,\n" +
		"https://test.imgix.net/image.jpg?w=400 400w"
	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcSetFromRange(t *testing.T) {
	c := testClient()
	c.SetUseLibParam(false)
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
	c.SetUseLibParam(false)
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
	c.SetUseLibParam(false)
	params := url.Values{"h": []string{"320"}, "ar": []string{"4:3"}}
	options := SrcSetOpts{disableVariableQuality: false}
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
	c.SetUseLibParam(false)
	wr := WidthRange{100, 8192, 1000.0}
	options := SrcSetOpts{widthRange: wr}

	expected := "https://test.imgix.net/image.png?w=100 100w,\n" +
		"https://test.imgix.net/image.png?w=8192 8192w"

	actual := c.CreateSrcSet("image.png", url.Values{}, options)
	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcSetFluidWidth100to108at2percent(t *testing.T) {
	c := testClient()
	c.SetUseLibParam(false)
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
	c.SetUseLibParam(false)
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
