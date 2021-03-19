package imgix

import (
	"os"
	"strings"
	"testing"
)

func TestReadMe_main(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net", WithLibParam(false))
	got := ub.CreateURL("path/to/image.jpg")
	want := "https://demo.imgix.net/path/to/image.jpg"

	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestReadMe_usageWithParams(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net", WithLibParam(false))
	got := ub.CreateURL("path/to/image.jpg", Param("w", "320"), Param("auto", "format", "compress"))
	want := "https://demo.imgix.net/path/to/image.jpg?auto=format%2Ccompress&w=320"

	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestReadMe_SecuredURLUsage(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net", WithToken("MYT0KEN"), WithLibParam(false))
	want := "https://demo.imgix.net/path/to/image.jpg?s=c8bd1807209f7f1d96dd7123f92febb4"
	got := ub.CreateURL("path/to/image.jpg")

	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestReadMe_usageSrcsetGeneration(t *testing.T) {
	ub := NewURLBuilder("demos.imgix.net", WithToken("foo123"))
	srcset := ub.CreateSrcset("image.png", []IxParam{})
	splitSrcset := strings.Split(srcset, "\n")
	const want = 31

	if len(splitSrcset) != want {
		t.Errorf("\ngot: %d; want: %d", len(splitSrcset), want)
	}
}

func TestReadMe_SignedSrcSetCreation(t *testing.T) {
	// Instead of using dotenv, just set the environment variable directly.
	const key = "IX_TOKEN"
	const wantToken = "MYT0KEN"
	os.Setenv(key, wantToken)

	gotToken := os.Getenv(key)
	if gotToken != wantToken {
		t.Errorf("\ngot:  %s\nwant: %s", gotToken, wantToken)
	}

	ub := NewURLBuilder("demos.imgix.net",
		WithToken(wantToken),
		WithLibParam(false))
	srcset := ub.CreateSrcset("image.png", []IxParam{})

	wantLength := 31
	splitSrcSet := strings.Split(srcset, ",\n")

	for _, u := range splitSrcSet {
		isSigned := strings.Contains(u, "s=")
		if !isSigned {
			t.Errorf("\ngot: %t; want: true", isSigned)
		}
	}

	gotLength := len(splitSrcSet)
	if wantLength != gotLength {
		t.Errorf("\ngot: %d; want: %d", gotLength, wantLength)
	}
}

func TestReadMe_FixedWidthSrcSetDefault(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net", WithLibParam(false))
	params := []IxParam{Param("h", "800"), Param("ar", "4:3")}
	want := "https://demo.imgix.net/image.png?ar=4%3A3&dpr=1&h=800&q=75 1x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=2&h=800&q=50 2x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=3&h=800&q=35 3x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=4&h=800&q=23 4x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=5&h=800&q=20 5x"
	got := ub.CreateSrcset("image.png", params)

	if got != want {
		t.Errorf("\ngot: \n%s\n\nwant: \n%s", got, want)
	}
}

func TestReadMe_FixedWidthSrcSetVariableQualityDisabled(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net", WithLibParam(false))
	params := []IxParam{Param("h", "800"), Param("ar", "4:3")}
	want := "https://demo.imgix.net/image.png?ar=4%3A3&dpr=1&h=800 1x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=2&h=800 2x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=3&h=800 3x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=4&h=800 4x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=5&h=800 5x"
	got := ub.CreateSrcset("image.png", params, WithVariableQuality(false))

	if got != want {
		t.Errorf("\ngot: \n%s\n\nwant: \n%s", got, want)
	}
}

func TestReadMe_FixedWidthSrcSetNoOpts(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net", WithLibParam(false))
	params := []IxParam{Param("h", "800"), Param("ar", "4:3")}
	want := "https://demo.imgix.net/image.png?ar=4%3A3&dpr=1&h=800&q=75 1x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=2&h=800&q=50 2x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=3&h=800&q=35 3x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=4&h=800&q=23 4x,\n" +
		"https://demo.imgix.net/image.png?ar=4%3A3&dpr=5&h=800&q=20 5x"
	got := ub.CreateSrcset("image.png", params)

	if got != want {
		t.Errorf("\ngot: \n%s\n\nwant: \n%s", got, want)
	}
}

func TestReadMe_FluidWidthSrcSetFromWidths(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net", WithLibParam(false))
	ixParams := []IxParam{Param("mask", "ellipse")}
	got := ub.CreateSrcsetFromWidths("image.jpg", ixParams, []int{100, 200, 300, 400})
	want := "https://demo.imgix.net/image.jpg?mask=ellipse&w=100 100w,\n" +
		"https://demo.imgix.net/image.jpg?mask=ellipse&w=200 200w,\n" +
		"https://demo.imgix.net/image.jpg?mask=ellipse&w=300 300w,\n" +
		"https://demo.imgix.net/image.jpg?mask=ellipse&w=400 400w"

	if got != want {
		t.Errorf("\ngot: \n%s\n\nwant: \n%s", got, want)
	}
}

func TestReadMe_FluidWidthSrcSet(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net", WithLibParam(false))

	got := ub.CreateSrcset(
		"image.png",
		[]IxParam{},
		WithMinWidth(100),
		WithMaxWidth(380),
		WithTolerance(0.08))

	want := "https://demo.imgix.net/image.png?w=100 100w,\n" +
		"https://demo.imgix.net/image.png?w=116 116w,\n" +
		"https://demo.imgix.net/image.png?w=135 135w,\n" +
		"https://demo.imgix.net/image.png?w=156 156w,\n" +
		"https://demo.imgix.net/image.png?w=181 181w,\n" +
		"https://demo.imgix.net/image.png?w=210 210w,\n" +
		"https://demo.imgix.net/image.png?w=244 244w,\n" +
		"https://demo.imgix.net/image.png?w=283 283w,\n" +
		"https://demo.imgix.net/image.png?w=328 328w,\n" +
		"https://demo.imgix.net/image.png?w=380 380w"

	if got != want {
		t.Errorf("\ngot: \n%s\n\nwant: \n%s", got, want)
	}
}

func TestReadMe_FluidWidthSrcsetTolerance20(t *testing.T) {
	ub := NewURLBuilder("demo.imgix.net", WithLibParam(false))

	srcsetOptions := []SrcsetOption{
		WithMinWidth(100),
		WithMaxWidth(384),
		WithTolerance(0.20),
	}

	got := ub.CreateSrcset(
		"image.png",
		[]IxParam{},
		srcsetOptions...)

	want := "https://demo.imgix.net/image.png?w=100 100w,\n" +
		"https://demo.imgix.net/image.png?w=140 140w,\n" +
		"https://demo.imgix.net/image.png?w=196 196w,\n" +
		"https://demo.imgix.net/image.png?w=274 274w,\n" +
		"https://demo.imgix.net/image.png?w=384 384w"

	if got != want {
		t.Errorf("\ngot: \n%s\n\nwant: \n%s", got, want)
	}
}

func TestReadMe_TargetWidths(t *testing.T) {
	want := []int{300, 378, 476, 600, 756, 953, 1200, 1513, 1906, 2401, 3000}
	got := TargetWidths(300, 3000, 0.13)

	for idx, v := range want {
		if got[idx] != v {
			t.Errorf("\ngot: %d; want: %d", got[idx], v)
		}
	}

	sm := want[:3]
	ub := NewURLBuilder("demos.imgix.net")
	ub.SetUseLibParam(false)
	wantSrcset := ub.CreateSrcsetFromWidths("image.png", []IxParam{}, sm)
	gotSrcset := "https://demos.imgix.net/image.png?w=300 300w,\n" +
		"https://demos.imgix.net/image.png?w=378 378w,\n" +
		"https://demos.imgix.net/image.png?w=476 476w"

	if gotSrcset != wantSrcset {
		t.Errorf("\ngot: \n%s\n\nwant: \n%s", gotSrcset, wantSrcset)
	}
}
