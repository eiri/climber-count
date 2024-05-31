package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

type Client struct {
	url string
}

func NewClient(url string) *Client {
	return &Client{
		url: url,
	}
}

func (c *Client) Counters() *Counters {
	file, err := os.Open(c.url)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	doc, err := html.Parse(file)
	if err != nil {
		log.Fatal(err)
	}

	return parse(doc)
}

func parse(n *html.Node) (counters *Counters) {
	if n.Type == html.TextNode && strings.Contains(n.Data, "var data = ") {
		data := extract(n.Data)
		counters = NewCounters()
		if err := json.Unmarshal([]byte(data), counters); err != nil {
			log.Fatal(err)
		}
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if cn := parse(c); cn != nil {
			counters = cn
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
