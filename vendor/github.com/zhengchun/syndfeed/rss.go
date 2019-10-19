package syndfeed

import (
	"errors"
	"html"
	"io"

	"github.com/antchfx/xmlquery"
)

// RSS is an RSS feed parser.
// https://validator.w3.org/feed/docs/rss2.html
type RSS struct{}

func (s *RSS) parseItemElement(self *xmlquery.Node, ns map[string]string) *Item {
	item := new(Item)
	for elem := self.FirstChild; elem != nil; elem = elem.NextSibling {
		if ns, ok := ns[elem.Prefix]; ok {
			item.ElementExtensions = append(item.ElementExtensions, &ElementExtension{elem.Data, elem.Prefix, elem.InnerText()})
			if m := lookupModule(ns); m != nil {
				m.ParseElement(elem, item)
			}
			continue
		}
		switch elem.Data {
		case "title":
			item.Title = elem.InnerText()
		case "link":
			item.Links = append(item.Links, &Link{URL: elem.InnerText()})
		case "description":
			item.Summary = html.UnescapeString(elem.InnerText())
		case "author":
			item.Authors = append(item.Authors, &Person{Name: elem.InnerText()})
		case "category":
			item.Categories = append(item.Categories, elem.InnerText())
		case "guid":
			item.Id = elem.InnerText()
		case "pubDate":
			if t, err := parseDateString(elem.InnerText()); err == nil {
				item.PublishDate = t
			}
		case "source":
			item.BaseURL = elem.SelectAttr("URL")
		case "comments":
		case "enclosure":
		}
	}
	return item
}

func (s *RSS) parse(doc *xmlquery.Node) (*Feed, error) {
	root := doc.SelectElement("rss")
	if root == nil {
		return nil, errors.New("invalid RSS document without <rss> element")
	}

	feed := new(Feed)
	// xmlns:prefix = namespace
	feed.Namespace = make(map[string]string)
	for _, attr := range root.Attr {
		switch {
		case attr.Name.Local == "version":
			feed.Version = attr.Value
		case attr.Name.Space == "xmlns":
			feed.Namespace[attr.Name.Local] = attr.Value
		}
	}

	var channel = root.SelectElement("channel")
	if channel == nil {
		return nil, errors.New("invalid RSS document without <channel> element")
	}

	for elem := channel.FirstChild; elem != nil; elem = elem.NextSibling {
		if ns, ok := feed.Namespace[elem.Prefix]; ok {
			feed.ElementExtensions = append(feed.ElementExtensions, &ElementExtension{elem.Data, elem.Prefix, elem.InnerText()})
			if m := lookupModule(ns); m != nil {
				m.ParseElement(elem, feed)
			}
			continue
		}
		switch elem.Data {
		case "title":
			feed.Title = elem.InnerText()
		case "description":
			feed.Description = elem.InnerText()
		case "link":
			feed.Links = append(feed.Links, &Link{URL: elem.InnerText()})
		case "language":
			feed.Language = elem.InnerText()
		case "copyright":
			feed.Copyright = elem.InnerText()
		case "lastBuildDate":
			if t, err := parseDateString(elem.InnerText()); err == nil {
				feed.LastUpdatedTime = t
			}
		case "category":
			// <category domain="Syndic8">1765</category>
			// <category>Grateful Dead</category>
			feed.Categories = append(feed.Categories, elem.InnerText())
		case "generator":
			feed.Generator = elem.InnerText()
		case "docs":
			// A URL that points to the documentation for the format used in the RSS file
			feed.BaseURL = elem.InnerText()
		case "image":
			if elem := elem.SelectElement("url"); elem != nil {
				feed.ImageURL = elem.InnerText()
			}
		case "item":
			item := s.parseItemElement(elem, feed.Namespace)
			feed.Items = append(feed.Items, item)
		case "pubDate":
		case "cloud":
		case "ttl":
		}
	}
	return feed, nil
}

// Parse parses an Rss feed.
func (s *RSS) Parse(r io.Reader) (*Feed, error) {
	doc, err := xmlquery.Parse(r)
	if err != nil {
		return nil, err
	}
	return s.parse(doc)
}

var rss = &RSS{}

// ParseRSS parses RSS feed from the specified io.Reader r.
func ParseRSS(r io.Reader) (*Feed, error) {
	return rss.Parse(r)
}
