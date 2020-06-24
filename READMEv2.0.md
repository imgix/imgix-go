<!-- ix-docs-ignore -->
<!-- Badges -->
![imgix logo](https://assets.imgix.net/sdk-imgix-logo.svg)

`imgix-go` is a client library for generating image URLs with [imgix](https://www.imgix.com/).

![Version](https://badge.fury.io/gh/imgix%2Fimgix-go.svg)
[![Godoc](https://godoc.org/github.com/imgix/imgix-go?status.svg)](https://godoc.org/github.com/imgix/imgix-go)
[![License](https://img.shields.io/github/license/imgix/imgix-go)](https://github.com/imgix/imgix-go/blob/master/LICENSE)

---
<!-- /ix-docs-ignore -->

<!-- Table of Contents -->
- [Installation](#installation)
- [Usage](#usage)
- [Signed URLs](#signed-urls)
- [Srcset Generation](#srcset-generation)

<!-- Installation Instructions -->
## Installation

```bash
go get github.com/imgix/imgix-go
```

<!-- Usage Instructions -->
## Usage

To begin creating imgix URLs, import the imgix library and create a URL builder. The URL builder can be reused to create URLs for any images on the domains it is provided.

```go
package main

import (
    "fmt"
    "net/url"
    "github.com/imgix/imgix-go"
)

func main() {
    ub := NewURLBuilder("demo.imgix.net")
    ixUrl := ub.CreateURL("path/to/image.jpg", url.Values{}))
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

The imgix-python package allows for generation of custom srcset attributes, which can be invoked through the `{create_srcset}` method. By default, the generated srcset will allow for responsive size switching by building a list of image-width mappings.

```go
```

The above will produce the following srcset attribute value which can then be served to the client: 

``` html
https://demos.imgix.net/image.png?w=100&s=9abb0d0db5a4901fcb6420a1a37efe5d 100w,
https://demos.imgix.net/image.png?w=116&s=cfea3b9598400fdb5dd273c50a666116 116w,
https://demos.imgix.net/image.png?w=135&s=e749702260debafa9aa71e55524b39ee 135w,
https://demos.imgix.net/image.png?w=156&s=0fb6a5f27dfece682320b73c466e1e39 156w,
										...
https://demos.imgix.net/image.png?w=7401&s=3b2fbb6aa880a260ba650dc773d47216 7401w,
https://demos.imgix.net/image.png?w=8192&s=1288314bbb33a4f441100b899dd67a00 8192w
```