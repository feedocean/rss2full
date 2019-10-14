package main

import (
	"bufio"
	"html"
	"io"

	"github.com/zhengchun/syndfeed"
)

func outputRss20(w io.Writer, feed *syndfeed.Feed) {
	writer := bufio.NewWriter(w)
	writer.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	writer.WriteString(`<?xml-stylesheet type="text/xsl" href="/assets/rss2full.xsl"?>`)
	writer.WriteString(`<rss xmlns:content="http://purl.org/rss/1.0/modules/content/" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/" version="2.0">`)
	// channel
	writer.WriteString(`<channel>`)
	// title
	writer.WriteString(`<title><![CDATA[` + feed.Title + `]]></title>`)
	// link
	if len(feed.Links) > 0 {
		writer.WriteString(`<link>` + feed.Links[0].URL + `</link>`)
	}
	if feed.Language != "" {
		writer.WriteString("<language>" + feed.Language + "</language>")
	}
	if feed.Description != "" {
		writer.WriteString(`<description>` + feed.Description + "</description>")
	}
	if !feed.LastUpdatedTime.IsZero() {
		writer.WriteString("<lastBuildDate>" + feed.LastUpdatedTime.Format("Mon, 02 Jan 2006 15:04:05 GMT") + "</lastBuildDate>")
	}
	writer.WriteString(`<generator>full-rss(https://github.com/feedocean/full-rss)</generator>`)
	// image
	if feed.ImageURL != "" {
		writer.WriteString("<image>")
		writer.WriteString(`<url>` + feed.ImageURL + `</url>`)
		writer.WriteString("</image>")
	}
	for i := 0; i < len(feed.Items); i++ {
		item := feed.Items[i]
		writer.WriteString("<item>")
		// title
		writer.WriteString(`<title><![CDATA[` + item.Title + `]]></title>`)
		// link
		writer.WriteString("<link>" + html.EscapeString(item.Links[0].URL) + "</link>")
		// description
		if item.Summary != "" {
			writer.WriteString(`<description><![CDATA[` + item.Summary + `]]></description>`)
		}
		// authors
		for _, v := range item.Authors {
			writer.WriteString(`<dc:creator><![CDATA[` + v.Name + `]]></dc:creator>`)
		}
		if item.Content != "" {
			writer.WriteString(`<content:encoded><![CDATA[` + item.Content + `]]></content:encoded>`)
		}
		for _, v := range item.Categories {
			writer.WriteString(`<category>` + v + `</category>`)
		}
		// pubDate
		if !item.PublishDate.IsZero() {
			writer.WriteString("<pubDate>" + feed.LastUpdatedTime.Format("Mon, 02 Jan 2006 15:04:05 GMT") + "</pubDate>")
		}
		writer.WriteString("</item>")
	}
	writer.WriteString(`</channel>`)
	writer.WriteString(`</rss>`)
	writer.Flush()
}
