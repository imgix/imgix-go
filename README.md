<!-- ix-docs-ignore -->
<!-- Badges -->
![imgix logo](https://assets.imgix.net/sdk-imgix-logo.svg)

`imgix-go` is a client library for generating image URLs with [imgix](https://www.imgix.com/).

![Version](https://badge.fury.io/gh/imgix%2Fimgix-go.svg)
[![Build Status](https://travis-ci.org/imgix/imgix-go.svg?branch=main)](https://travis-ci.org/imgix/imgix-go)
[![Godoc](https://godoc.org/github.com/imgix/imgix-go?status.svg)](https://godoc.org/github.com/imgix/imgix-go)
[![License](https://img.shields.io/github/license/imgix/imgix-go)](https://github.com/imgix/imgix-go/blob/main/LICENSE)

---
<!-- /ix-docs-ignore -->

<!-- Table of Contents -->
- [Installation](#installation)
- [Usage](#usage)
- [Signed URLs](#signed-urls)
- [Srcset Generation](#srcset-generation)
    - [Fixed-Width Images](#fixed-width-images)
        - [Variable Quality](#variable-quality)
    - [Fluid-Width Images](#fluid-width-images)
        - [Custom Widths](#custom-widths)
        - [Width Ranges](#width-ranges)
        - [Width Tolerance](#width-tolerance)
        - [Explore Target Widths](#explore-target-widths)
- [The `ixlib` Parameter](#the-ixlib-parameter)
- [Testing](#testing)

<!-- Installation Instructions -->
## Installation

```bash
go get github.com/imgix/imgix-go
```

<!-- Usage Instructions -->
## Usage

To begin creating imgix URLs, import the imgix library and create a URL builder. The URL builder can be reused to create URLs for any images on the domain it is provided.

```go
package main

import (
    "fmt"
    "net/url"
    "github.com/imgix/imgix-go"
)

func main() {
    ub := NewURLBuilder("demo.imgix.net")
    ixUrl := ub.CreateURL("path/to/image.jpg", url.Values{})
    // ixUrl == "https://demo.imgix.net/path/to/image.jpg"
}
```

```go
ub := NewURLBuilder("demo.imgix.net")
ub.CreateURL("path/to/image.jpg", url.Values{"h": []string{"100"}, "w": []string{"100"}})
// "https://demo.imgix.net/path/to/image.jpg?h=100&w=100"
```

_HTTPS_ support is enabled by default. _HTTP_ can be toggled on by setting `use_https` to `False`:

```go
ub := NewURLBuilder("demo.imgix.net")
ub.SetUseHTTPS(false)
ub.CreateURL("path/to/image.jpg", url.Values{})
// "http://demo.imgix.net/path/to/image.jpg"
```

## Signed URLs

To produce a signed URL, you must enable secure URLs on your source and then provide your signature key to the URL builder.

First, be sure to keep your secrets safe.

**.env**
```text
IX_TOKEN="token"
```

**main.go**
```go
package main

import (
	"log"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"github.com/imgix/imgix-go"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ixToken := os.Getenv("IX_TOKEN")
	ub := NewURLBuilderWithToken("demo.imgix.net", ixToken)

	expected := "https://demo.imgix.net/path/to/image.jpg?s=5dde0b0e48067925082d670d0e987fcb"
	actual := ub.CreateSignedURL("path/to/image.jpg", url.Values{})
}
```

## Srcset Generation

The imgix-go package allows for generation of custom srcset attributes, which can be invoked through the `CreateSrcSet` method. By default, the generated srcset will allow for responsive size switching by building a list of image-width mappings.

```go
ub := NewURLBuilderWithToken("demos.imgix.net", token)
srcset := ub.CreateSrcSet("image.png", url.Values{}, DefaultOpts)
```

The above will produce a srcset the following srcset attribute value which can then be served to the client: 

``` html
https://demos.imgix.net/image.png?w=100&s=9abb0d0db5a4901fcb6420a1a37efe5d 100w,
https://demos.imgix.net/image.png?w=116&s=cfea3b9598400fdb5dd273c50a666116 116w,
https://demos.imgix.net/image.png?w=135&s=e749702260debafa9aa71e55524b39ee 135w,
https://demos.imgix.net/image.png?w=156&s=0fb6a5f27dfece682320b73c466e1e39 156w,
										...
https://demos.imgix.net/image.png?w=7401&s=3b2fbb6aa880a260ba650dc773d47216 7401w,
https://demos.imgix.net/image.png?w=8192&s=1288314bbb33a4f441100b899dd67a00 8192w
```


### Fixed-Width Images

In cases where enough information is provided about an image's dimensions, `CreateSrcSet` will build a srcset that will allow for an image to be served at different resolutions. The parameters taken into consideration when determining if an image is fixed-width are `w`, `h`, and `ar`.

By invoking `CreateSrcSet` with either a width **or** the height and aspect ratio in the parameters, a fixed-width srcset will be generated. Wherein, the image width is fixed, but the pixel density varies.

```go
ub := NewURLBuilder("demo.imgix.net")
params := url.Values{"h": []string{"800"}, "ar": []string{"4:3"}}
options := SrcSetOpts{disableVariableQuality: false}
srcset := ub.CreateSrcSet("image.png", params, options)
```

Will produce the following attribute value:

``` html
https://demo.imgix.net/image.png?ar=4%3A3&dpr=1&h=800&q=75 1x
https://demo.imgix.net/image.png?ar=4%3A3&dpr=2&h=800&q=50 2x
https://demo.imgix.net/image.png?ar=4%3A3&dpr=3&h=800&q=35 3x
https://demo.imgix.net/image.png?ar=4%3A3&dpr=4&h=800&q=23 4x
https://demo.imgix.net/image.png?ar=4%3A3&dpr=5&h=800&q=20 5x
```

For more information to better understand srcset, we highly recommend
[Eric Portis' "Srcset and sizes" article](https://ericportis.com/posts/2014/srcset-sizes/) which goes into depth about the subject.

#### Variable Quality

This library will automatically append a variable `q` parameter mapped to each `dpr` parameter when generating a [fixed-width image](#fixed-width-images) srcset. This technique is commonly used to compensate for the increased file size of high-DPR images.

Since high-DPR images are displayed at a higher pixel density on devices, image quality can be lowered to reduce overall file size––without sacrificing perceived visual quality. For more information and examples of this technique in action, see [this blog post](https://blog.imgix.com/2016/03/30/dpr-quality).

This behavior will respect any overriding `q` value passed in as a parameter. Additionally, it can be disabled altogether by passing `SrcSetOpts{disableVariableQuality: true}` to `CreateSrcSet`.


### Fluid-Width Images

#### Custom Widths

In situations where specific widths are desired when generating `srcset` pairs, a user can specify them by passing an array of positive integers to `CreateSrcSetFromWidths`:

```go
ub := NewURLBuilder("demo.imgix.net")
srcset := ub.CreateSrcSetFromWidths("image.jpg", url.Values{}, []int{100, 200, 300, 400})
```
```html
https://demo.imgix.net/image.jpg?w=100 100w,
https://demo.imgix.net/image.jpg?w=200 200w,
https://demo.imgix.net/image.jpg?w=300 300w,
https://demo.imgix.net/image.jpg?w=400 400w
```

#### Width Ranges

In certain circumstances, you may want to limit the minimum or maximum value of the non-fixed `srcset` generated by the `CreateSrcSet` method. To do this, you can specify the widths at which a srcset should `begin` and `end` by specifying a `WidthRange` in your `SrcSetOpts`:

```go
ub := NewURLBuilder("demo.imgix.net")
options := SrcSetOpts{widthRange: WidthRange{begin: 100, end: 380, tol: 0.08}}
srcset := ub.CreateSrcSet("image.png", url.Values{}, options)
```

```html
https://demo.imgix.net/image.png?w=100 100w,
https://demo.imgix.net/image.png?w=116 116w,
https://demo.imgix.net/image.png?w=135 135w,
https://demo.imgix.net/image.png?w=156 156w,
https://demo.imgix.net/image.png?w=181 181w,
https://demo.imgix.net/image.png?w=210 210w,
https://demo.imgix.net/image.png?w=244 244w,
https://demo.imgix.net/image.png?w=283 283w,
https://demo.imgix.net/image.png?w=328 328w,
https://demo.imgix.net/image.png?w=380 380w
```

#### Width Tolerance

The `srcset` width `tol`erance dictates the maximum tolerated difference between an image's downloaded size and its rendered size.

For example, setting this value to `0.10` means that an image will not render more than 10% larger or smaller than its native size. In practice, the image URLs generated for a width-based srcset attribute will grow by twice this rate.

A lower tolerance means images will render closer to their native size (thereby increasing perceived image quality), but a large srcset list will be generated and consequently users may experience lower rates of cache-hit for pre-rendered images on your site.

By default, srcset width tolerance is set to 0.08 (8 percent), which we consider to be the ideal rate for maximizing cache hits without sacrificing visual quality. Users can specify their own width tolerance by providing a positive scalar value as width tolerance.

In this case, the width tolerance is set to 20 percent:

```go
options := SrcSetOpts{widthRange: WidthRange{begin: 100, end: 384, tol: 0.20}}
```

```html
https://demo.imgix.net/image.jpg?w=100 100w,
https://demo.imgix.net/image.jpg?w=140 140w,
https://demo.imgix.net/image.jpg?w=196 196w,
https://demo.imgix.net/image.jpg?w=274 274w,
https://demo.imgix.net/image.jpg?w=384 384w
```

#### Explore Target Widths

The `TargetWidths` function is used internally to generate lists of target widths to be used in calls to `CreateSrcSet`.

It is a way to generate, play with, and explore different target widths separately from srcset attributes. We've already seen how to generate srcset attributes when the begin, end, and tolerance values are known.

Another approach is to use `TargetWidths` to determine which combination of values for `start`, `end`, and `tol`erance work best.

```go
// Create
widths := TargetWidths(300, 3000, 0.13)

// Explore
sm := widths[:3]
expectedSm := []int{300, 378, 476}
assert.Equal(t, expectedSm, sm)

md := widths[3:7]
expectedMd := []int{600, 756, 953, 1200}
assert.Equal(t, expectedMd, md)

lg := widths[7:]
expectedLg := []int{1513, 1906, 2401, 3000}
assert.Equal(t, expectedLg, lg)

// Serve
ub := NewURLBuilder("demos.imgix.net")
srcset := ub.CreateSrcSetFromWidths("image.png", url.Values{}, sm)
// "https://demos.imgix.net/image.png?w=300 300w,\nhttps://demos.imgix.net/image.png?w=378 378w,\nhttps://demos.imgix.net/image.png?w=476 476w"
```

<!-- FAQs -->
## The `ixlib` Parameter

For security and diagnostic purposes, we sign all requests with the language and version of library used to generate the URL.

The `ixlib` parameter can be toggled off by setting `useLibParam` via `SetUseLibParam`:

```go
ub := NewURLBuilder("demo.imgix.net")
ub.SetUseLibParam(false)
```

<!-- Test Instructions -->
## Testing

You can go test this code with:

``` bash
$ go test
```
