package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

const USER_AGENT = "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0"

type ClientRoundTripper struct{}

func (crt ClientRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	logger := slog.Default().With("component", "client")
	logger.Info("sending request", "url", req.URL)

	resp, err = http.DefaultTransport.RoundTrip(req)
	if err != nil {
		logger.Error("bad reply", "msg", err)
		return
	}

	logger.Info("got response", "status", resp.Status)
	return
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	req    *http.Request
	client HTTPClient
}

func NewClient(cfg *Config) *Client {
	url := fmt.Sprintf("https://portal.rockgympro.com/portal/public/%s/occupancy?iframeid=occupancyCounter&fId=%s", cfg.PGK, cfg.FID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", USER_AGENT)

	rt := ClientRoundTripper{}
	return &Client{req, &http.Client{Transport: rt}}
}

func (c *Client) Counters() (*Counters, error) {
	logger := slog.Default().With("component", "client")
	counters := NewCounters()

	body, err := c.fetch()
	if err != nil {
		logger.Error("can't fetch page", "msg", err)
		return counters, err
	}
	defer body.Close()

	doc, err := html.Parse(body)
	if err != nil {
		logger.Error("can't parse page", "msg", err)
		return counters, err
	}

	data := parse(doc)

	err = json.Unmarshal([]byte(data), counters)

	return counters, err
}

func (c *Client) fetch() (io.ReadCloser, error) {
	resp, err := c.client.Do(c.req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to fetch URL: %s, status: %d", c.req.RequestURI, resp.StatusCode)
	}

	return resp.Body, nil
}

func parse(n *html.Node) (data string) {
	if n.Type == html.TextNode && strings.Contains(n.Data, "var data = ") {
		data = extract(n.Data)
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if d := parse(c); d != "" {
			data = d
			break
		}
	}
	return
}

func extract(str string) string {
	var b strings.Builder
	// 640 ought to be enough
	b.Grow(640)
	skip := true
	for _, ch := range str {
		if ch == '=' {
			skip = false
			continue
		}
		if skip {
			continue
		}
		if ch == ';' {
			break
		}
		if ch == '\'' {
			ch = '"'
		}
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return strings.ReplaceAll(b.String(), "},}", "}}")
}
