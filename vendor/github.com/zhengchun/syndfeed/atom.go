package syndfeed

import (
	"errors"
	"html"
	"io"

	"github.com/antchfx/xmlquery"
)

// Atom is an Atom feed parser.
// https://validator.w3.org/feed/docs/rfc4287.html
type Atom struct{}

func (a *Atom) parseItemElement(elem *xmlquery.Node, ns map[string]string) *Item {
	item := new(Item)
	for elem := elem.FirstChild; elem != nil; elem = elem.NextSibling {
		if ns, ok := ns[elem.Prefix]; ok {
			item.ElementExtensions = append(item.ElementExtensions, &ElementExtension{elem.Data, elem.Prefix, elem.InnerText()})
			if m := lookupModule(ns); m != nil {
				m.ParseElement(elem, item)
			}
		}
		switch elem.Data {
		case "author":
			item.Authors = append(item.Authors, a.parseAuthorElement(elem))
		case "category":
			if v := elem.SelectAttr("term"); v != "" {
				item.Categories = append(item.Categories, v)
			}
		case "contributor":
			item.Contributors = append(item.Contributors, a.parseAuthorElement(elem))
		case "id":
			item.Id = elem.InnerText()
		case "title":
			item.Title = elem.InnerText()
		case "link":
			item.Links = append(item.Links, a.parseLinkElement(elem))
		case "published":
			if t, err := parseDateString(elem.InnerText()); err == nil {
				item.PublishDate = t
			}
		case "rights":
			item.Copyright = elem.InnerText()
		case "summary":
			item.Summary = html.UnescapeString(elem.InnerText())
		case "updated":
			if t, err := parseDateString(elem.InnerText()); err == nil {
				item.LastUpdatedTime = t
			}
		case "content":
			item.Content = html.UnescapeString(elem.InnerText())
		case "source":
		}
	}
	return item
}

func (a *Atom) parseLinkElement(elem *xmlquery.Node) *Link {
	return &Link{
		URL:       html.UnescapeString(elem.SelectAttr("href")),
		Title:     elem.SelectAttr("title"),
		MediaType: elem.SelectAttr("type"),
		RelType:   elem.SelectAttr("rel"),
	}
}

func (a *Atom) parseAuthorElement(elem *xmlquery.Node) *Person {
	author := new(Person)
	if n := elem.SelectElement("name"); n != nil {
		author.Name = n.InnerText()
	}
	if n := elem.SelectElement("uri"); n != nil {
		author.URL = n.InnerText()
	}
	if n := elem.SelectElement("email"); n != nil {
		author.Email = n.InnerText()
	}
	return author
}

func (a *Atom) parse(doc *xmlquery.Node) (*Feed, error) {
	root := doc.SelectElement("feed")
	if root == nil {
		return nil, errors.New("invalid Atom document without feed element")
	}

	feed := &Feed{Version: "1.0"} // default atom version is 1.0
	feed.Namespace = make(map[string]string)
	// xmlns:prefix = namespace
	for _, attr := range root.Attr {
		switch {
		case attr.Name.Local == "version":
			feed.Version = attr.Value
		case attr.Name.Space == "xmlns":
			feed.Namespace[attr.Name.Local] = attr.Value
		}
	}

	for elem := root.FirstChild; elem != nil; elem = elem.NextSibling {
		if ns, ok := feed.Namespace[elem.Prefix]; ok {
			feed.ElementExtensions = append(feed.ElementExtensions, &ElementExtension{elem.Data, elem.Prefix, elem.InnerText()})
			if m := lookupModule(ns); m != nil {
				m.ParseElement(elem, feed)
			}
			continue
		}
		switch elem.Data {
		case "author":
			feed.Authors = append(feed.Authors, a.parseAuthorElement(elem))
		case "category":
			if v := elem.SelectAttr("term"); v != "" {
				feed.Categories = append(feed.Categories, v)
			}
		case "contributor":
			feed.Contributors = append(feed.Contributors, a.parseAuthorElement(elem))
		case "generator":
			feed.Generator = elem.InnerText()
		case "id":
			feed.Id = elem.InnerText()
		case "link":
			feed.Links = append(feed.Links, a.parseLinkElement(elem))
		case "logo":
			feed.ImageURL = elem.InnerText()
		case "rights":
			feed.Copyright = elem.InnerText()
		case "subtitle":
			feed.Description = elem.InnerText()
		case "title":
			feed.Title = elem.InnerText()
		case "updated":
			if t, err := parseDateString(elem.InnerText()); err == nil {
				feed.LastUpdatedTime = t
			}
		case "entry":
			item := a.parseItemElement(elem, feed.Namespace)
			feed.Items = append(feed.Items, item)
		case "icon":
		case "published":
		case "source":
		}
	}
	return feed, nil
}

// Parse parses an atom feed.
func (a *Atom) Parse(r io.Reader) (*Feed, error) {
	doc, err := xmlquery.Parse(r)
	if err != nil {
		return nil, err
	}
	return a.parse(doc)
}

var atom = &Atom{}

// ParseAtom parses Atom feed from the specified io.Reader r.
func ParseAtom(r io.Reader) (*Feed, error) {
	return atom.Parse(r)
}
