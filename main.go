package main

import (
	"io"
	"bytes"
	"errors"
	"strings"
	"io/ioutil"
	"log"
	"net/http"
	"fmt"
	"golang.org/x/net/html"

	// "github.com/grokify/html-strip-tags-go"
)


func main() {
	// Get the URL/Domain for validation
	// @todo Need to figure out how to get all urls from a domain efficiently
	// Can we use sitemap?
	var URL string
	fmt.Print("Enter the url for schema validation: ")
	fmt.Scan(&URL)

	// Get the page source
	// @todo Loop a list of all urls on a domain and get page source of each page
	pageSource, err := getPageSource(URL)
	if err != nil {
		log.Fatal(err)
	}

	htm, err := html.Parse(strings.NewReader(pageSource))
	if err != nil {
		log.Fatal(err)
	}

	schema, err := getSchema(htm)
	if err != nil {
		log.Fatal(err)
	}

	schemaMarkup := renderNodes(schema)
	fmt.Println(schemaMarkup)
}

func getPageSource(URL string) (string, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return URL, err
	}
	defer resp.Body.Close()

	pageSource, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return URL, err
	}
	
	return string(pageSource), err
}

func getSchema(htm *html.Node) ([]*html.Node, error) {
	var schema []*html.Node
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, script := range n.Attr {
				if script.Val == "application/ld+json" {
					schema = append(schema, n)
				}
			}
		}
		for s := n.FirstChild; s != nil; s = s.NextSibling {
			f(s)
		}
	}
	f(htm)
	if (schema != nil) {
		return schema, nil
	}
	return nil, errors.New("No schema markup found in the page")
}

func renderNodes(nodes []*html.Node) []string {
	var renderedNodes []string
	for _, n := range nodes {
		stripped := stripScriptTags(n, `<script type = "application/ld+json" >`, "</script>")
		renderedNodes = append(renderedNodes, stripped)
		}
	return renderedNodes
}

func stripScriptTags(node *html.Node, firstTrim string, secondTrim string) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, node)
	str := buf.String()
	stripped := strings.Trim(strings.Trim(str, firstTrim), secondTrim)

	return stripped
}