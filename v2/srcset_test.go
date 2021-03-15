package imgix

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testClient() URLBuilder {
	return NewURLBuilder("test.imgix.net", WithLibParam(false))
}

func testClientWithToken() URLBuilder {
	return NewURLBuilder("my-social-network.imgix.net", WithToken("FOO123bar"))
}

func TestURLBuilder_CreateSrcSetFromWidths(t *testing.T) {
	c := testClient()
	actual := c.CreateSrcsetFromWidths("image.jpg", []IxParam{}, []int{100, 200, 300, 400})
	expected := "https://test.imgix.net/image.jpg?w=100 100w,\n" +
		"https://test.imgix.net/image.jpg?w=200 200w,\n" +
		"https://test.imgix.net/image.jpg?w=300 300w,\n" +
		"https://test.imgix.net/image.jpg?w=400 400w"
	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcSetFromRange(t *testing.T) {
	c := testClient()
	// Example of setting the useLibParam after initial construction.
	c.SetUseLibParam(false)

	actual := c.CreateSrcset(
		"image.png",
		[]IxParam{},
		WithMinWidth(100),
		WithMaxWidth(380),
		WithTolerance(0.08))

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
	expected := "https://test.imgix.net/image.png?dpr=1&q=75&w=320 1x,\n" +
		"https://test.imgix.net/image.png?dpr=2&q=50&w=320 2x,\n" +
		"https://test.imgix.net/image.png?dpr=3&q=35&w=320 3x,\n" +
		"https://test.imgix.net/image.png?dpr=4&q=23&w=320 4x,\n" +
		"https://test.imgix.net/image.png?dpr=5&q=20&w=320 5x"
	actual := c.CreateSrcset("image.png", []IxParam{Param("w", "320")})
	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcSetFixedHandAR(t *testing.T) {
	c := testClient()
	params := []IxParam{Param("h", "320"), Param("ar", "4:3")}
	expected := "https://test.imgix.net/image.png?ar=4%3A3&dpr=1&h=320&q=75 1x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=2&h=320&q=50 2x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=3&h=320&q=35 3x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=4&h=320&q=23 4x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=5&h=320&q=20 5x"
	actual := c.CreateSrcset("image.png", params, WithVariableQuality(true))
	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcsetFixedHandARImplicitVarQuality(t *testing.T) {
	// Same as above, but omitting WithVariableQuality(true) to show that variable
	// quality is the implicit-default.
	c := testClient()
	params := []IxParam{Param("h", "320"), Param("ar", "4:3")}
	expected := "https://test.imgix.net/image.png?ar=4%3A3&dpr=1&h=320&q=75 1x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=2&h=320&q=50 2x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=3&h=320&q=35 3x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=4&h=320&q=23 4x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=5&h=320&q=20 5x"
	actual := c.CreateSrcset("image.png", params)
	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcSetFixedHInDprForm(t *testing.T) {
	c := testClient()
	params := []IxParam{Param("h", "320")}
	expected := [5]string{"1x", "2x", "3x", "4x", "5x"}
	srcset := c.CreateSrcset("image.png", params, WithVariableQuality(true))
	src := strings.Split(srcset, ",")

	for i := 0; i < len(src); i++ {
		dpr := strings.Split(src[i], " ")[1]
		assert.Contains(t, expected, dpr)
	}
}

func TestURLBuilder_CreateSrcSetFixedH(t *testing.T) {
	c := testClient()
	params := []IxParam{Param("h", "320")}
	expected := "https://test.imgix.net/image.png?dpr=1&h=320&q=75 1x,\n" +
		"https://test.imgix.net/image.png?dpr=2&h=320&q=50 2x,\n" +
		"https://test.imgix.net/image.png?dpr=3&h=320&q=35 3x,\n" +
		"https://test.imgix.net/image.png?dpr=4&h=320&q=23 4x,\n" +
		"https://test.imgix.net/image.png?dpr=5&h=320&q=20 5x"
	actual := c.CreateSrcset("image.png", params, WithVariableQuality(true))
	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcSetFluidHighTol(t *testing.T) {
	c := testClient()

	expected := "https://test.imgix.net/image.png?w=100 100w,\n" +
		"https://test.imgix.net/image.png?w=8192 8192w"

	actual := c.CreateSrcset(
		"image.png",
		[]IxParam{},
		WithMinWidth(100),
		WithMaxWidth(8192),
		WithTolerance(1000.0))

	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcSetFluidWidth100to108at2percent(t *testing.T) {
	c := testClient()

	expected := "https://test.imgix.net/image.png?w=100 100w,\n" +
		"https://test.imgix.net/image.png?w=104 104w,\n" +
		"https://test.imgix.net/image.png?w=108 108w"

	actual := c.CreateSrcset(
		"image.png",
		[]IxParam{},
		WithMinWidth(100),
		WithMaxWidth(108),
		WithTolerance(0.02))

	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcsetQOverridesWithVariableQuality(t *testing.T) {
	c := testClient()
	params := []IxParam{Param("h", "800"), Param("ar", "4:3"), Param("q", "99")}

	expected := "https://test.imgix.net/image.png?ar=4%3A3&dpr=1&h=800&q=99 1x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=2&h=800&q=99 2x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=3&h=800&q=99 3x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=4&h=800&q=99 4x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=5&h=800&q=99 5x"

	actual := c.CreateSrcset("image.png", params, WithVariableQuality(true))
	assert.Equal(t, expected, actual)
}

func TestURLBuilder_CreateSrcsetQOverridesWithoutVariableQuality(t *testing.T) {
	c := testClient()
	params := []IxParam{Param("h", "800"), Param("ar", "4:3"), Param("q", "99")}

	expected := "https://test.imgix.net/image.png?ar=4%3A3&dpr=1&h=800&q=99 1x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=2&h=800&q=99 2x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=3&h=800&q=99 3x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=4&h=800&q=99 4x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=5&h=800&q=99 5x"

	actual := c.CreateSrcset("image.png", params, WithVariableQuality(false))
	assert.Equal(t, expected, actual)
}
