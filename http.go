package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/goreadly"
	"github.com/antchfx/htmlquery"
	"github.com/sirupsen/logrus"

	"github.com/julienschmidt/httprouter"
	"github.com/zhengchun/syndfeed"
	"golang.org/x/net/html/charset"
)

var httpClient = &http.Client{
	Timeout: time.Second * 45,
}

func httpGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
	// bug has fixed： https://github.com/golang/go/issues/18779
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	r, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	resp.Body = &responseReader{rc: resp.Body, r: r}
	return resp, nil
}

type responseReader struct {
	rc io.ReadCloser
	r  io.Reader
}

func (r *responseReader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func (r *responseReader) Close() error {
	return r.rc.Close()
}

func Version(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte(appVersion))
}

func FullRss(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	source := strings.TrimLeft(p.ByName("feed"), "/")
	if source == "" || !(strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")) {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Invalid source feed(%s)", source)))
		return
	}
	handleError := func(err error) {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}
	resp, err := httpGet(source)
	if err != nil {
		handleError(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		handleError(fmt.Errorf("%s got status-code is not 200(%d)", source, resp.StatusCode))
		return
	}
	v := resp.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(v)
	var feed *syndfeed.Feed
	switch mediatype {
	case "text/xml",
		"application/xml",
		"application/rss+xml",
		"application/atom+xml":
		var err error
		feed, err = syndfeed.Parse(resp.Body)
		if err != nil {
			handleError(err)
			return
		}
	default:
		handleError(fmt.Errorf("%s got mediatype is not supported(%s)", source, mediatype))
		return
	}
	if len(feed.Items) > 0 {
		var wg sync.WaitGroup
		c := make(chan struct{})
		// create 2 worker to work.
		var queue = make(chan *syndfeed.Item, 2)
		for n := 0; n < 2; n++ {
			go func() {
				for {
					select {
					case item := <-queue:
						link := item.Links[0].URL
						logrus.Debugf("full-text: %s", link)
						if err := fulltext(item, link); err != nil {
							wg.Done()
							logrus.Warnf("GET %s failed. %s", link, err)
						}
						wg.Done()
					case <-c:
						return
					}
				}
			}()
		}
		for _, item := range feed.Items {
			if len(item.Links) > 0 {
				wg.Add(1)
				queue <- item
			}
		}
		wg.Wait()
		close(c)
	}
	// outout RSS 2.0
	w.Header().Set("Content-Type", "application/xml")
	outputRss20(w, feed)
}

func fulltext(item *syndfeed.Item, link string) error {
	resp, err := httpGet(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	htmlDoc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return err
	}
	doc, err := goreadly.ParseHTML(resp.Request.URL, htmlDoc)
	if err != nil {
		return err
	}
	item.Content = doc.Body
	return nil
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}
