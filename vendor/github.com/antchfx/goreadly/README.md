goreadly
===
goreadly is Go package is to makes Web pages more readable.

[![GoDoc](https://godoc.org/github.com/antchfx/goreadly?status.svg)](https://godoc.org/github.com/antchfx/goreadly)

Installation
===
    go get github.com/antchfx/goreadly

Example
===
```go
package main

import (
	"fmt"
	"net/http"

	"github.com/antchfx/goreadly"
)

func main() {
	resp, _ := http.Get("https://www.engadget.com/2017/07/10/google-highlights-pirate-sites/")
	doc, err := goreadly.ParseResponse(resp)
	if err != nil {
		panic(err)
	}
	fmt.Println(doc.Title)
	fmt.Println(doc.Body)
}
```