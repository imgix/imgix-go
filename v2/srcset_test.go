package imgix

import (
	"strings"
	"testing"
)

func testClient() URLBuilder {
	return NewURLBuilder("test.imgix.net", WithLibParam(false))
}

func testClientWithToken() URLBuilder {
	return NewURLBuilder("my-social-network.imgix.net", WithToken("FOO123bar"))
}

func TestURLBuilder_CreateSrcSetFromWidths(t *testing.T) {
	c := testClient()
	got := c.CreateSrcsetFromWidths("image.jpg", []IxParam{}, []int{100, 200, 300, 400})
	want := "https://test.imgix.net/image.jpg?w=100 100w,\n" +
		"https://test.imgix.net/image.jpg?w=200 200w,\n" +
		"https://test.imgix.net/image.jpg?w=300 300w,\n" +
		"https://test.imgix.net/image.jpg?w=400 400w"

	if got != want {
		t.Errorf("\ngot: \n%s\n\nwant: \n%s", got, want)
	}
}

func TestURLBuilder_CreateSrcSetFromRange(t *testing.T) {
	c := testClient()
	// Example of setting the useLibParam after initial construction.
	c.SetUseLibParam(false)

	got := c.CreateSrcset(
		"image.png",
		[]IxParam{},
		WithMinWidth(100),
		WithMaxWidth(380),
		WithTolerance(0.08))

	want := "https://test.imgix.net/image.png?w=100 100w,\n" +
		"https://test.imgix.net/image.png?w=116 116w,\n" +
		"https://test.imgix.net/image.png?w=135 135w,\n" +
		"https://test.imgix.net/image.png?w=156 156w,\n" +
		"https://test.imgix.net/image.png?w=181 181w,\n" +
		"https://test.imgix.net/image.png?w=210 210w,\n" +
		"https://test.imgix.net/image.png?w=244 244w,\n" +
		"https://test.imgix.net/image.png?w=283 283w,\n" +
		"https://test.imgix.net/image.png?w=328 328w,\n" +
		"https://test.imgix.net/image.png?w=380 380w"

	if got != want {
		t.Errorf("\ngot: \n%s\n\nwant: \n%s", got, want)
	}
}

func TestURLBuilder_CreateSrcSetFixedW(t *testing.T) {
	c := testClient()
	want := "https://test.imgix.net/image.png?dpr=1&q=75&w=320 1x,\n" +
		"https://test.imgix.net/image.png?dpr=2&q=50&w=320 2x,\n" +
		"https://test.imgix.net/image.png?dpr=3&q=35&w=320 3x,\n" +
		"https://test.imgix.net/image.png?dpr=4&q=23&w=320 4x,\n" +
		"https://test.imgix.net/image.png?dpr=5&q=20&w=320 5x"
	got := c.CreateSrcset("image.png", []IxParam{Param("w", "320")})

	if got != want {
		t.Errorf("\ngot: \n%s\n\nwant: \n%s", got, want)
	}
}

func TestURLBuilder_CreateSrcsetFixedHandAR(t *testing.T) {
	c := testClient()
	params := []IxParam{Param("h", "320"), Param("ar", "4:3")}
	want := "https://test.imgix.net/image.png?ar=4%3A3&dpr=1&h=320&q=75 1x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=2&h=320&q=50 2x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=3&h=320&q=35 3x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=4&h=320&q=23 4x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=5&h=320&q=20 5x"
	got := c.CreateSrcset("image.png", params, WithVariableQuality(true))

	if got != want {
		t.Errorf("\ngot: \n%s\n\nwant: \n%s", got, want)
	}
}

func TestURLBuilder_CreateSrcsetFixedHandARImplicitVarQuality(t *testing.T) {
	// Same as above, but omitting WithVariableQuality(true) to show that variable
	// quality is the implicit-default.
	c := testClient()
	params := []IxParam{Param("h", "320"), Param("ar", "4:3")}
	want := "https://test.imgix.net/image.png?ar=4%3A3&dpr=1&h=320&q=75 1x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=2&h=320&q=50 2x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=3&h=320&q=35 3x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=4&h=320&q=23 4x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=5&h=320&q=20 5x"
	got := c.CreateSrcset("image.png", params)

	if got != want {
		t.Errorf("\ngot:  %s\n\nwant: %s", got, want)
	}
}

func TestURLBuilder_CreateSrcsetFixedHInDprForm(t *testing.T) {
	c := testClient()
	params := []IxParam{Param("h", "320")}
	want := [5]string{"1x", "2x", "3x", "4x", "5x"}
	srcset := c.CreateSrcset("image.png", params, WithVariableQuality(true))
	src := strings.Split(srcset, ",")

	for i, _ := range src {
		gotDPR := strings.Split(src[i], " ")[1]

		if gotDPR != want[i] {
			t.Errorf("got: %s; want: %s", gotDPR, want[i])
		}
	}
}

func TestURLBuilder_CreateSrcSetFixedH(t *testing.T) {
	c := testClient()
	params := []IxParam{Param("h", "320")}
	want := "https://test.imgix.net/image.png?dpr=1&h=320&q=75 1x,\n" +
		"https://test.imgix.net/image.png?dpr=2&h=320&q=50 2x,\n" +
		"https://test.imgix.net/image.png?dpr=3&h=320&q=35 3x,\n" +
		"https://test.imgix.net/image.png?dpr=4&h=320&q=23 4x,\n" +
		"https://test.imgix.net/image.png?dpr=5&h=320&q=20 5x"
	got := c.CreateSrcset("image.png", params, WithVariableQuality(true))

	if got != want {
		t.Errorf("\ngot:  %s\n\nwant: %s", got, want)
	}
}

func TestURLBuilder_CreateSrcSetFluidHighTol(t *testing.T) {
	c := testClient()

	want := "https://test.imgix.net/image.png?w=100 100w,\n" +
		"https://test.imgix.net/image.png?w=8192 8192w"

	got := c.CreateSrcset(
		"image.png",
		[]IxParam{},
		WithMinWidth(100),
		WithMaxWidth(8192),
		WithTolerance(1000.0))

	if got != want {
		t.Errorf("\ngot:  %s\n\nwant: %s", got, want)
	}
}

func TestURLBuilder_CreateSrcSetFluidWidth100to108at2percent(t *testing.T) {
	c := testClient()

	want := "https://test.imgix.net/image.png?w=100 100w,\n" +
		"https://test.imgix.net/image.png?w=104 104w,\n" +
		"https://test.imgix.net/image.png?w=108 108w"

	got := c.CreateSrcset(
		"image.png",
		[]IxParam{},
		WithMinWidth(100),
		WithMaxWidth(108),
		WithTolerance(0.02))

	if got != want {
		t.Errorf("\ngot:  %s\n\nwant: %s", got, want)
	}
}

func TestURLBuilder_CreateSrcsetQOverridesWithVariableQuality(t *testing.T) {
	c := testClient()
	params := []IxParam{Param("h", "800"), Param("ar", "4:3"), Param("q", "99")}

	want := "https://test.imgix.net/image.png?ar=4%3A3&dpr=1&h=800&q=99 1x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=2&h=800&q=99 2x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=3&h=800&q=99 3x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=4&h=800&q=99 4x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=5&h=800&q=99 5x"

	got := c.CreateSrcset("image.png", params, WithVariableQuality(true))

	if got != want {
		t.Errorf("\ngot:  %s\n\nwant: %s", got, want)
	}
}

func TestURLBuilder_CreateSrcsetQOverridesWithoutVariableQuality(t *testing.T) {
	c := testClient()
	params := []IxParam{Param("h", "800"), Param("ar", "4:3"), Param("q", "99")}

	want := "https://test.imgix.net/image.png?ar=4%3A3&dpr=1&h=800&q=99 1x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=2&h=800&q=99 2x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=3&h=800&q=99 3x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=4&h=800&q=99 4x,\n" +
		"https://test.imgix.net/image.png?ar=4%3A3&dpr=5&h=800&q=99 5x"

	got := c.CreateSrcset("image.png", params, WithVariableQuality(false))

	if got != want {
		t.Errorf("\ngot:  %s\n\nwant: %s", got, want)
	}
}
