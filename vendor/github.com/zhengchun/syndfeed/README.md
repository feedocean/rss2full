Overview 
===
[![Coverage Status](https://coveralls.io/repos/github/zhengchun/syndfeed/badge.svg?branch=master)](https://coveralls.io/github/zhengchun/syndfeed?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/zhengchun/syndfeed)](https://goreportcard.com/report/github.com/zhengchun/syndfeed)
[![Build Status](https://travis-ci.org/zhengchun/syndfeed.svg?branch=master)](https://travis-ci.org/zhengchun/syndfeed)
[![GoDoc](https://godoc.org/github.com/zhengchun/syndfeed?status.svg)](https://godoc.org/github.com/zhengchun/syndfeed)
[![MIT license](https://img.shields.io/badge/License-MIT-blue.svg)](https://lbesson.mit-license.org/)

`syndfeed` is a Go library for [RSS](https://en.wikipedia.org/wiki/RSS) 2.0 and [Atom](https://en.wikipedia.org/wiki/Atom_(standard)) 1.0 feeds, supported implement extension module to parse any RSS and Atom extension element.

Dependencies
===
- [xmlquery](https://github.com/antchfx/xmlquery)

Getting Started
===

#### Parse a feed from URL

```go
feed, _ := syndfeed.LoadURL("https://cn.engadget.com/rss.xml")
fmt.Println(feed.Title)
```

#### Parse a feed from io.Reade

```go
feedData := `<rss version="2.0">
<channel>
<title>Sample Feed</title>
</channel>
</rss>`
feed, _ := syndfeed.Parse(strings.NewReader(feedData))
fmt.Println(feed.Title)
```

#### Parse an Atom feed into `syndfeed.Feed`

```go
feedData := `<feed xmlns="http://www.w3.org/2005/Atom">
<title>Example Atom</title>
</feed>`
feed, _ := syndfeed.ParseAtom(strings.NewReader(feedData))
fmt.Println(feed.Title)
```

#### Parse a RSS feed into `syndfeed.Feed`

```go
feedData := `<rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/">
<channel>
<dc:creator>example@site.com (Example Name)</dc:creator>
</channel>`
feed, _ := syndfeed.ParseRSS(strings.NewReader(feedData))
fmt.Println(feed.Authors[0].Name)
```

Extension Modules
===

In some syndication feeds, they have some valid XML elements but are not specified in either the Atom 1.0 or RSS 2.0 specifications. You can add extension module to process these elements.

The `syndfeed` build-in supported following modules: [Dublin Core](http://web.resource.org/rss/1.0/modules/dc/), [Content](http://web.resource.org/rss/1.0/modules/content/), [Syndication](http://web.resource.org/rss/1.0/modules/syndication/). 

You can implement your own modules to parse any extension element like: [iTunes RSS](https://rss.itunes.apple.com/en-us), [Media RSS](http://www.rssboard.org/media-rss).

#### iTunes Module

```go
iTunesHandler := func(n *xmlquery.Node, v interface{}) {
    item := v.(*syndfeed.Item)
    switch n.Data {
    case "artist":
        item.Authors = append(item.Authors, &syndfeed.Person{Name: n.InnerText()})
    case "releaseDate":
		item.PublishDate = ParseDate(n.InnerText())
    }
}
// Register a new module to the syndfeed module list.
syndfeed.RegisterExtensionModule("https://rss.itunes.apple.com", syndfeed.ModuleHandlerFunc(iTunesHandler))
// Now can parse iTunes feed. 
feed, err := syndfeed.LoadURL("https://github.com/zhengchun/syndfeed/blob/master/_samples/itunes.atom")
fmt.Println(feed.Items[0].Authors[0].Name)
// Output author name.
```

TODO
===

- Add RSS/Atom format output.

#### Please let me know if you have any questions.
