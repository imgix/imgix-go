<!-- ix-docs-ignore -->
![imgix logo](https://assets.imgix.net/sdk-imgix-logo.svg)

`imgix-go` is a client library for generating image URLs with [imgix](https://www.imgix.com/).

![Version](https://badge.fury.io/gh/imgix%2Fimgix-go.svg)
[![Build Status](https://travis-ci.org/imgix/imgix-go.svg?branch=main)](https://travis-ci.org/parkr/imgix-go)
[![Godoc](https://godoc.org/github.com/imgix/imgix-go?status.svg)](https://godoc.org/github.com/imgix/imgix-go)
[![License](https://img.shields.io/github/license/imgix/imgix-go)](https://github.com/imgix/imgix-go/blob/main/LICENSE)

---
<!-- /ix-docs-ignore -->

- [Installation](#installation)
- [Usage](#usage)

## Installation

```bash
go get github.com/imgix/imgix-go
```

## Usage

```go
package main

import (
    "fmt"
    "net/url"
    "github.com/imgix/imgix-go"
)

func main() {
    client := imgix.NewClient("mycompany.imgix.net")

    fmt.Println(client.Path("/myImage.jpg"))

    fmt.Println(client.PathWithParams("/myImage.jpg", url.Values{
        "w": []string{"400"},
        "h": []string{"400"},
    }))
}
```
