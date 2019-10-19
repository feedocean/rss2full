package goreadly

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var (
	// MinTextLength specified the minimum length of content.
	MinTextLength = 25

	blacklistCandidatesRegexp  = regexp.MustCompile(`(?i)popupbody`)
	okMaybeItsACandidateRegexp = regexp.MustCompile(`(?i)and|article|body|column|main|shadow|post`)
	unlikelyCandidatesRegexp   = regexp.MustCompile(`(?i)combx|comment|community|hidden|disqus|modal|extra|foot|header|menu|remark|rss|shoutbox|sidebar|sponsor|ad-break|agegate|pagination|pager|popup`)
	divToPElementsRegexp       = regexp.MustCompile(`(?i)(a|blockquote|dl|div|img|ol|p|pre|table|ul|select)`)

	negativeRegexp = regexp.MustCompile(`(?i)combx|comment|com-|foot|footer|footnote|masthead|media|meta|outbrain|promo|related|scroll|shoutbox|sidebar|sponsor|shopping|tags|tool|widget`)
	positiveRegexp = regexp.MustCompile(`(?i)article|body|content|entry|hentry|main|page|pagination|post|text|blog|story`)

	sentenceRegexp = regexp.MustCompile(`\.( |$)`)

	//normalizeCRLFRegexp       = regexp.MustCompile(`(\r\n|\r|\n)+`)
	//normalizeWhitespaceRegexp = regexp.MustCompile(`\s{2,}`)

	selfClosingHtmlTags = map[string]bool{
		"area":   true,
		"base":   true,
		"embed":  true,
		"input":  true,
		"link":   true,
		"meta":   true,
		"param":  true,
		"source": true,
		"track":  true,
		"hr":     true,
		"img":    true,
		"br":     true,
	}

	allowedHTMLTagAttrs = map[string]bool{
		"src":         true,
		"href":        true,
		"width":       true,
		"height":      true,
		"frameborder": true,
	}
)

// A Document represents an article document object.
type Document struct {
	URL *url.URL

	Title, Body string
}

type candidate struct {
	node  *html.Node
	score float32
}

// hasChildBlockElement determines whether element has any children block level elements.
func hasChildBlockElement(n *html.Node) bool {
	var hasBlock bool
	for _, n := range htmlquery.Find(n, "descendant::*") {
		hasBlock = hasBlock || divToPElementsRegexp.MatchString(n.Data)
	}
	return hasBlock
}

// hasSinglePInsideElement checks if this node has only whitespace and a single P
// element returns false if the DIV node contains non-empty text nodes
// or if it contains no P or more than 1 element.
func hasSinglePInsideElement(n *html.Node) (*html.Node, bool) {
	var c, l int
	var p *html.Node
	for _, n := range htmlquery.Find(n, "p") {
		p = n
		c++
		for _, n2 := range htmlquery.Find(n, "text()") {
			l += len(strings.TrimSpace(n2.Data))
		}
	}
	return p, c == 1 && l > 0
}

func parseTitle(doc *html.Node) string {
	var title, betterTitle string
	if n := htmlquery.FindOne(doc, "//meta[@property='og:title' or @name='twitter:title']"); n != nil {
		title = htmlquery.SelectAttr(n, "content")
	} else if n := htmlquery.FindOne(doc, "//title"); n != nil {
		title = htmlquery.InnerText(n)
	}
	var seps = []string{" | ", " _ ", " - ", "«", "»", "—"}
	for _, sep := range seps {
		if array := strings.Split(title, sep); len(array) > 1 {
			if len(betterTitle) > 0 {
				// conflict with separate character
				betterTitle = title
				break
			}
			betterTitle = strings.TrimSpace(array[0])
		}
	}
	if len(betterTitle) > 10 {
		return betterTitle
	}
	return title
}

func multiExpr(top *html.Node, exprs ...string) []*html.Node {
	var list []*html.Node
	for _, expr := range exprs {
		v := htmlquery.Find(top, expr)
		list = append(list, v...)
	}
	return list
}

func parseBody(self *url.URL, doc *html.Node) string {
	// Replaces 2 or more successive <br> elements with a single <p>.
	// Whitespace between <br> elements are ignored. For example:
	// <div>foo<br>bar<br> <br><br>abc</div>
	// will become:
	// <div>foo<br>bar<p>abc</p></div>
	nextElement := func(n *html.Node) (next *html.Node) {
		// Finds the next element, starting from the given node, and ignoring
		// whitespace in between. If the given node is an element, the same node is
		// returned.
		for next != nil && n.Type == html.TextNode && strings.TrimSpace(n.Data) == "" {
			next = next.NextSibling
		}
		return
	}

	for _, n := range htmlquery.Find(doc, "//br") {
		parent := n.Parent
		replaced := false
		// If we find a <br> chain, remove the <br>s until we hit another element
		// or non-whitespace. This leaves behind the first <br> in the chain
		// (which will be replaced with a <p> later).
		for next := nextElement(n.NextSibling); next != nil && next.Data == "br"; next = nextElement(next) {
			replaced = true
			sibling := next.NextSibling
			parent.RemoveChild(next)
			next = sibling
		}

		// If we removed a <br> chain, replace the remaining <br> with a <p>. Add
		// all sibling nodes as children of the <p> until we hit another <br>
		// chain.
		if replaced {
			p := &html.Node{
				Data: "p",
				Type: html.ElementNode,
				Attr: make([]html.Attribute, 0),
			}
			parent.InsertBefore(p, n)
			parent.RemoveChild(n)
			for next := p.NextSibling; next != nil; {
				// If we've hit another <br><br>, we're done adding children to this <p>.
				if next.Data == "br" {
					if next := nextElement(next); next != nil && next.Data == "br" {
						break
					}
				}
				// Otherwise, make this node a child of the new <p>.
				sibling := next.NextSibling
				parent.RemoveChild(next)
				p.AppendChild(next)
				next = sibling
			}
		}
	}

	// remove unlikely candidates
	for _, n := range htmlquery.Find(doc, "//*") {
		switch n.Data {
		case "script", "style", "noscript":
			removeNodes(n)
			continue
		case "html", "body", "article":
			continue
		}
		str := htmlquery.SelectAttr(n, "class") + htmlquery.SelectAttr(n, "id")
		if blacklistCandidatesRegexp.MatchString(str) || (unlikelyCandidatesRegexp.MatchString(str) && !okMaybeItsACandidateRegexp.MatchString(str)) {
			removeNodes(n)
		}
	}
	// turn all divs that don't have children block level elements into p's
	for _, n := range htmlquery.Find(doc, "//div") {
		// Sites like http://mobile.slate.com encloses each paragraph with a DIV
		// element. DIVs with only a P element inside and no text content can be
		// safely converted into plain P elements to avoid confusing the scoring
		// algorithm with DIVs with are, in practice, paragraphs.
		if p, ok := hasSinglePInsideElement(n); ok {
			n.RemoveChild(p)
			n.Parent.InsertBefore(p, n)
			n.Parent.RemoveChild(n)
		} else if !hasChildBlockElement(n) {
			n.Data = "p"
		} else {
			// EXPERIMENTAL
			for _, n := range htmlquery.Find(n, "text()") {
				if len(strings.TrimSpace(n.Data)) > 0 {
					p := &html.Node{
						Data: "p",
						Type: html.ElementNode,
						Attr: []html.Attribute{
							html.Attribute{
								Key: "class",
								Val: "readability-styled",
							}},
					}

					n.Parent.InsertBefore(p, n)
					n.Parent.RemoveChild(n)
					p.AppendChild(n)
				}
			}
		}
	}
	// loop through all paragraphs, and assign a score to them based on how content-y they look.
	candidates := make(map[*html.Node]*candidate)

	for _, n := range multiExpr(doc, "//p", "//td") {
		text := htmlquery.InnerText(n)
		count := utf8.RuneCountInString(text)
		// if this paragraph is less than x chars, don't count it
		if count < MinTextLength {
			continue
		}

		parent := n.Parent
		grandparent := parent.Parent
		if _, ok := candidates[parent]; !ok {
			candidates[parent] = scoreNode(parent)
		}
		if grandparent != nil {
			if _, ok := candidates[grandparent]; !ok {
				candidates[grandparent] = scoreNode(grandparent)
			}
		}
		contentScore := float32(1.0)
		// for any commas within this paragraph
		contentScore += float32(strings.Count(text, ","))
		contentScore += float32(strings.Count(text, "，")) // gb2312 character
		contentScore += float32(math.Min(float64(int(count/100.0)), 3))

		candidates[parent].score += contentScore
		if grandparent != nil {
			candidates[grandparent].score += contentScore / 2.0
		}
	}

	// scale the final candidates score based on link density. Good content
	// should have a relatively small link density (5% or less) and be mostly
	// unaffected by this operation
	var best *candidate
	for _, candidate := range candidates {
		candidate.score = candidate.score * (1 - getLinkDensity(candidate.node))
		if best == nil || best.score < candidate.score {
			best = candidate
		}
	}
	// if still have no top candidate, just use the body as a last resort.
	if best == nil {
		best = &candidate{htmlquery.FindOne(doc, "//body"), 0}
		return ""
	}

	// now that we have the top candidate, look through its siblings for content that might also be related.
	// like preambles, content split by ads that we removed, etc.
	var list []*html.Node

	siblingScoreThreshold := float32(math.Max(10, float64(best.score*.2)))
	for n := best.node.Parent.FirstChild; n != nil; n = n.NextSibling {
		canAppend := false
		if n == best.node {
			canAppend = true
		} else if c, ok := candidates[n]; ok && c.score >= siblingScoreThreshold {
			canAppend = true
		}

		if n.Data == "p" {
			linkDensity := getLinkDensity(n)
			content := htmlquery.InnerText(n)
			contentLength := utf8.RuneCountInString(content)
			if contentLength >= 80 && linkDensity < .25 {
				canAppend = true
			} else if contentLength < 80 && linkDensity == 0 {
				canAppend = sentenceRegexp.MatchString(content)
			}
		}
		if canAppend {
			list = append(list, n)
		}
	}
	// we have all of the content that we need.
	// now we clean it up for presentation.
	return sanitize(self, list)
}

func sanitize(u *url.URL, a []*html.Node) string {
	// clean out spurious headers from an element.
	b := a[:0]
	for _, n := range a {
		switch n.Data {
		case "h1", "h2", "h3", "h4", "h5", "h6", "h7":
			if classWeight(n) < 0 || getLinkDensity(n) > 0.33 {
				continue
			}
		case "input", "select", "textarea", "button", "object", "iframe", "embed":
			continue
		}
		b = append(b, n)
	}

	c := b[:0]
	for _, n := range b {
		if n.Data == "table" || n.Data == "ul" || n.Data == "div" {
			weight := float32(classWeight(n))
			if weight < 0 {
				continue
			}
			text := htmlquery.InnerText(n)
			if strings.Count(text, ",")+strings.Count(text, "，") < 10 {
				// if there are not very many commas, and the number of
				// non-paragraph elements is more than paragraphs or other ominous signs, remove the element.
				var (
					p     = len(multiExpr(n, "//p", "//br"))
					img   = len(htmlquery.Find(n, "//img"))
					li    = len(htmlquery.Find(n, "//li")) - 100
					embed = len(htmlquery.Find(n, "//embed[@src]"))
					input = len(htmlquery.Find(n, "//input"))
				)

				contentLength := len(strings.TrimSpace(text))
				linkDensity := getLinkDensity(n)
				remove := false
				if img > p && img > 1 {
					remove = true
				} else if li > p && n.Data != "ul" && n.Data != "ol" {
					remove = true
				} else if input > (p / 3.0) {
					remove = true
				} else if contentLength < MinTextLength && (img == 0 || img > 2) {
					remove = true
				} else if weight < 25 && linkDensity > 0.2 {
					remove = true
				} else if weight >= 25 && linkDensity > 0.5 {
					remove = true
				} else if (embed == 1 && contentLength < 75) || embed > 1 {
					remove = true
				}

				if remove {
					continue
				}
			}
		}
		c = append(c, n)
	}

	if len(c) == 0 {
		return ""
	}

	isFakeElement := func(n *html.Node) bool {
		if n.Data != "p" {
			return false
		}
		for _, attr := range n.Attr {
			if attr.Key == "class" && attr.Val == "readability-styled" {
				return true
			}
		}
		return false
	}

	var fn func(*bytes.Buffer, *html.Node)
	fn = func(buf *bytes.Buffer, n *html.Node) {
		switch {
		case n.Type == html.TextNode:
			buf.WriteString(n.Data)
			return
		case n.Type == html.CommentNode:
			return
		}
		// Check element n whether is created by readability package.
		faked := isFakeElement(n)
		if !faked {
			buf.WriteString("<" + n.Data)
			for _, attr := range n.Attr {
				if !allowedHTMLTagAttrs[attr.Key] || strings.HasPrefix(attr.Val, "javascript:") {
					continue
				}
				if attr.Key == "src" || attr.Key == "href" {
					if !(strings.HasPrefix(attr.Val, "http://") ||
						strings.HasPrefix(attr.Val, "https://") ||
						strings.HasPrefix(attr.Val, "ftp://") ||
						strings.HasPrefix(attr.Val, "mailto:")) {
						if u, err := u.Parse(attr.Val); err == nil {
							attr.Val = u.String()
						}
					}
				}
				buf.WriteString(" " + attr.Key + "=\"" + attr.Val + "\"")
			}
			if selfClosingHtmlTags[n.Data] {
				buf.WriteString("/>")
			} else {
				buf.WriteString(">")
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			fn(buf, child)
		}
		if !faked && !selfClosingHtmlTags[n.Data] {
			buf.WriteString("</" + n.Data + ">")
		}
	}

	var buf bytes.Buffer
	for _, node := range c {
		for n := node.FirstChild; n != nil; n = n.NextSibling {
			fn(&buf, n)
		}
	}
	return buf.String()
	//return normalizeCRLFRegexp.ReplaceAllString(normalizeWhitespaceRegexp.ReplaceAllString(text, " "), "\n")
}

func cleanConditionally(n *html.Node, tags ...string) {
	for i, tag := range tags {
		tags[i] = "//" + tag
	}
	selector := strings.Join(tags, "|")
	for _, n := range htmlquery.Find(n, selector) {
		weight := float32(classWeight(n))
		if weight < 0 {
			removeNodes(n)
			return
		}
		text := htmlquery.InnerText(n)
		if strings.Count(text, ",")+strings.Count(text, "，") < 10 {
			// if there are not very many commas, and the number of
			// non-paragraph elements is more than paragraphs or other ominous signs, remove the element.
			var (
				p     = len(multiExpr(n, "//p", "//br"))
				img   = len(htmlquery.Find(n, "//img"))
				li    = len(htmlquery.Find(n, "//li")) - 100
				embed = len(htmlquery.Find(n, "//embed[@src]"))
				input = len(htmlquery.Find(n, "//input"))
			)

			contentLength := len(strings.TrimSpace(text))
			linkDensity := getLinkDensity(n)
			remove := false
			if img > p && img > 1 {
				remove = true
			} else if li > p && n.Data != "ul" && n.Data != "ol" {
				remove = true
			} else if input > (p / 3.0) {
				remove = true
			} else if contentLength < MinTextLength && (img == 0 || img > 2) {
				remove = true
			} else if weight < 25 && linkDensity > 0.2 {
				remove = true
			} else if weight >= 25 && linkDensity > 0.5 {
				remove = true
			} else if (embed == 1 && contentLength < 75) || embed > 1 {
				remove = true
			}

			if remove {
				removeNodes(n)
			}
		}
	}
}

func scoreNode(n *html.Node) *candidate {
	contentScore := classWeight(n)
	switch n.Data {
	case "article":
		contentScore += 10
	case "section":
		contentScore += 8
	case "div":
		contentScore += 5
	case "pre", "td", "blockquote":
		contentScore += 3
	case "address", "ol", "ul", "dl", "dd", "dt", "li", "form":
		contentScore -= 3
	case "h1", "h2", "h3", "h4", "h5", "h6", "th":
		contentScore -= 5
	}
	// checking node has itemscope??
	for _, attr := range n.Attr {
		if attr.Key == "itemscope" {
			contentScore += 5
		}
		if attr.Key == "itemtype" {
			contentScore += 30
		}
	}
	return &candidate{n, float32(contentScore)}
}

func classWeight(n *html.Node) int {
	weight := 0
	if v := htmlquery.SelectAttr(n, "class"); v != "" {
		if negativeRegexp.MatchString(v) {
			weight -= 25
		}

		if positiveRegexp.MatchString(v) {
			weight += 25
		}
	}
	if v := htmlquery.SelectAttr(n, "id"); v != "" {
		if negativeRegexp.MatchString(v) {
			weight -= 25
		}

		if positiveRegexp.MatchString(v) {
			weight += 25
		}
	}
	return weight
}

func getLinkDensity(n *html.Node) float32 {
	textLength := utf8.RuneCountInString(htmlquery.InnerText(n))
	if textLength == 0 {
		return 0
	}
	linkLength := 0
	for _, n := range htmlquery.Find(n, "//a") {
		if v := htmlquery.SelectAttr(n, "href"); v == "" || v == "#" {
			continue
		}
		linkLength += utf8.RuneCountInString(htmlquery.InnerText(n))
	}
	return float32(linkLength) / float32(textLength)
}

func removeNodes(n *html.Node) {
	if n.Parent == nil {
		return
	}
	if n.NextSibling != nil {
		n.NextSibling.PrevSibling = n.PrevSibling
	}
	if n.PrevSibling != nil {
		n.PrevSibling.NextSibling = n.NextSibling
	}
	if n.Parent.FirstChild == n {
		n.Parent.FirstChild = n.NextSibling
	}
}

func parseHTML(self *url.URL, doc *html.Node) (*Document, error) {
	return &Document{
		URL:   self,
		Title: parseTitle(doc),
		Body:  parseBody(self, doc),
	}, nil
}

// ParseResponse parses an HTTP document and convert its to readability.
func ParseResponse(res *http.Response) (*Document, error) {
	doc, err := htmlquery.Parse(res.Body)
	if err != nil {
		return nil, fmt.Errorf("htmlquery.Parse occured error: %s", err)
	}
	return parseHTML(res.Request.URL, doc)
}

// ParseResponse parses an HTTP document and convert its to readability.
func ParseHTML(self *url.URL, doc *html.Node) (*Document, error) {
	if self == nil {
		return nil, errors.New("URL self is nil")
	}
	return parseHTML(self, doc)
}
