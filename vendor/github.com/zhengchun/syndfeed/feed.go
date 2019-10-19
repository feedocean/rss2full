package syndfeed

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/antchfx/xmlquery"
)

// Feed is top-level feed object, <feed> in Atom 1.0 and
// <rss> in RSS 2.0.
type Feed struct {
	Authors      []*Person
	BaseURL      string
	Categories   []string
	Contributors []*Person
	Copyright    string
	Namespace    map[string]string // map[namespace-prefix]namespace-url
	Description  string
	Generator    string
	Id           string
	ImageURL     string
	Items        []*Item
	Language     string
	// LastUpdatedTime is the feed was last updated time.
	LastUpdatedTime   time.Time
	Title             string
	Links             []*Link
	Version           string
	ElementExtensions []*ElementExtension
}

// Link represents a link within a syndication
// feed or item.
type Link struct {
	MediaType string
	URL       string
	Title     string
	RelType   string
}

// Item is a feed item.
type Item struct {
	BaseURL      string
	Authors      []*Person
	Contributors []*Person
	Categories   []string
	Content      string
	Copyright    string
	Id           string
	// LastUpdatedTime is the feed item last updated time.
	LastUpdatedTime time.Time
	Links           []*Link
	// PublishDate is the feed item publish date.
	PublishDate       time.Time
	Summary           string
	Title             string
	ElementExtensions []*ElementExtension
	//CommentURL      string
}

// Person is an author or contributor of the feed content.
type Person struct {
	Name  string
	URL   string
	Email string
}

// ElementExtension is an syndication element extension.
type ElementExtension struct {
	Name, Namespace, Value string
}

// Parse parses a syndication feed(RSS,Atom).
func Parse(r io.Reader) (*Feed, error) {
	doc, err := xmlquery.Parse(r)
	if err != nil {
		return nil, err
	}
	if doc.SelectElement("rss") != nil {
		return rss.parse(doc)
	} else if doc.SelectElement("feed") != nil {
		return atom.parse(doc)
	}
	return nil, errors.New("invalid syndication feed without <rss> or <feed> element")
}

// LoadURL loads a syndication feed URL.
func LoadURL(url string) (*Feed, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return Parse(res.Body)
}
