## GoCache

[![Go Report Card](https://goreportcard.com/badge/github.com/pkg6/go-cache)](https://goreportcard.com/report/github.com/pkg6/go-cache)
[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/pkg6/go-cache?tab=doc)
[![Sourcegraph](https://sourcegraph.com/github.com/pkg6/go-cache/-/badge.svg)](https://sourcegraph.com/github.com/pkg6/go-cache?badge)
[![Release](https://img.shields.io/github/release/pkg6/go-cache.svg?style=flat-square)](https://github.com/pkg6/go-cache/releases)


## Installation

Make sure you have a working Go environment (Go 1.18 or higher is required). See the [install instructions](https://golang.org/doc/install.html).

To install [GoCache](https://github.com/pkg6/go-cache), simply run:

```
go get github.com/pkg6/go-cache
```

## Example

```
package main

import (
	"github.com/pkg6/go-cache"
)

func main() {
	c := cache.New()
	c.Extend(cache.NewFileCache())
	c.Set("cache", "test", 0)
	c.Get("cache")
	c.GetMulti([]string{"cache"})
	c.Delete("cache")
	c.Has("cache")
	c.Increment("cache_inc", 1)
	c.Decrement("cache_dec", 1)
	c.Clear()
	c.Pull("cache")
	c.Remember("cache", func() any {
		return "test1"
	}, 0)
}
```
