package main

import (
	"html"
	"io"

	"github.com/zhengchun/syndfeed"
)

func outputRss20(sw io.StringWriter, feed *syndfeed.Feed) {
	sw.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sw.WriteString(`<?xml-stylesheet type="text/xsl" href="/assets/rss2full.xsl"?>`)
	sw.WriteString(`<rss xmlns:content="http://purl.org/rss/1.0/modules/content/" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/" version="2.0">`)
	// channel
	sw.WriteString(`<channel>`)
	// title
	sw.WriteString(`<title><![CDATA[` + feed.Title + `]]></title>`)
	// link
	if len(feed.Links) > 0 {
		sw.WriteString(`<link>` + feed.Links[0].URL + `</link>`)
	}
	if feed.Language != "" {
		sw.WriteString("<language>" + feed.Language + "</language>")
	}
	if feed.Description != "" {
		sw.WriteString(`<description>` + feed.Description + "</description>")
	}
	if !feed.LastUpdatedTime.IsZero() {
		sw.WriteString("<lastBuildDate>" + feed.LastUpdatedTime.Format("Mon, 02 Jan 2006 15:04:05 GMT") + "</lastBuildDate>")
	}
	sw.WriteString(`<generator>full-rss(https://github.com/feedocean/full-rss)</generator>`)
	// image
	if feed.ImageURL != "" {
		sw.WriteString("<image>")
		sw.WriteString(`<url>` + feed.ImageURL + `</url>`)
		sw.WriteString("</image>")
	}
	for i := 0; i < len(feed.Items); i++ {
		item := feed.Items[i]
		sw.WriteString("<item>")
		// title
		sw.WriteString(`<title><![CDATA[` + item.Title + `]]></title>`)
		// link
		sw.WriteString("<link>" + html.EscapeString(item.Links[0].URL) + "</link>")
		// description
		if item.Summary != "" {
			sw.WriteString(`<description><![CDATA[` + item.Summary + `]]></description>`)
		}
		// authors
		for _, v := range item.Authors {
			sw.WriteString(`<dc:creator><![CDATA[` + v.Name + `]]></dc:creator>`)
		}
		if item.Content != "" {
			sw.WriteString(`<content:encoded><![CDATA[` + item.Content + `]]></content:encoded>`)
		}
		for _, v := range item.Categories {
			sw.WriteString(`<category>` + v + `</category>`)
		}
		// pubDate
		if !item.PublishDate.IsZero() {
			sw.WriteString("<pubDate>" + item.PublishDate.Format("Mon, 02 Jan 2006 15:04:05 GMT") + "</pubDate>")
		}
		sw.WriteString("</item>")
	}
	sw.WriteString(`</channel>`)
	sw.WriteString(`</rss>`)
}
