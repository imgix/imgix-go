# imgix-go

This is a Go implementation of an imgix url-building library outlined by
[imgix-blueprint](https://github.com/imgix/imgix-blueprint).

[Godoc](https://godoc.org/github.com/parkr/imgix-go)

[![Build Status](https://travis-ci.org/parkr/imgix-go.svg?branch=master)](https://travis-ci.org/parkr/imgix-go)

## Installation

It's a go package. Do this in your terminal:

```bash
go get github.com/imgix/imgix-go
```

## Usage

Something like this:

```go
package main

import (
    "fmt"
    "net/url"
    "github.com/imgix/imgix-go"
)

func main() {
    client := imgix.NewClient("mycompany.imgix.net")

    // Nothing fancy.
    fmt.Println(client.Path("/myImage.jpg"))

    // Throw some params in there!
    fmt.Println(client.PathWithParams("/myImage.jpg", url.Values{
        "w": []string{"400"},
        "h": []string{"400"},
    }))
}
```

That's it at a basic level. More fun features though!
