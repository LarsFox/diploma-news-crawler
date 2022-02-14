package meduza

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"

	"github.com/LarsFox/diploma-news-crawler/models"
)

// Crawler works with meduza.io.
type Crawler struct{}

// New returns a new crawler.
func New() *Crawler {
	return &Crawler{}
}

// Category lists articles in a category.
func (c *Crawler) Category(category string, offset, perPage int) ([]*models.Article, error) {
	u := fmt.Sprintf(urlArticles, category, offset/perPage, perPage)
	resp, err := http.DefaultClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("error getting category: %w", err)
	}

	articles := &respArticles{}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(articles); err != nil {
		return nil, fmt.Errorf("error decoding category: %w", err)
	}

	result := make([]*models.Article, 0, len(articles.Documents))
	for _, article := range articles.Documents {
		article := newArticle(category, article)
		if article == nil {
			continue
		}
		result = append(result, article)
	}
	return result, nil
}

func newArticle(category string, item *respDocument) *models.Article {
	if !strings.HasPrefix(item.URL, category) {
		return nil
	}
	return &models.Article{
		Category: category,
		ID:       fmt.Sprintf("%s-@-%d", category, item.PublishedAt),
		URL:      fmt.Sprintf(urlArticle, item.URL),
	}
}

// Enrich parses article text.
func (c *Crawler) Enrich(article *models.Article) error {
	resp, err := http.DefaultClient.Get(article.URL)
	if err != nil {
		return fmt.Errorf("err getting article: %w", err)
	}

	item := &respArticle{}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(item); err != nil {
		return fmt.Errorf("error decoding document: %w", err)
	}

	defer resp.Body.Close()
	document, err := html.Parse(strings.NewReader(item.Root.Content.Body))
	if err != nil {
		return fmt.Errorf("err parsing page: %w", err)
	}

	textsBlock := extractNode(document, "Body")
	if textsBlock == nil {
		return nil
	}

	src := item.Root.Title + "\n"
	src += parseNode(textsBlock)
	if src == "" {
		return nil
	}

	src = strings.Trim(src, "\n")
	src = regSpaces.ReplaceAllString(src, "")
	article.Text = src
	return nil
}

func extractNode(node *html.Node, className string) *html.Node {
	for _, attr := range node.Attr {
		if attr.Key == "class" && attr.Val == className {
			return node
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if node := extractNode(c, className); node != nil {
			return node
		}
	}
	return nil
}

func parseNode(node *html.Node) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case html.TextNode:
		return node.Data

	case html.ElementNode:
		// Removing ads.
		for _, attr := range node.Attr {
			if attr.Key != "class" {
				continue
			}
			if strings.HasPrefix(attr.Val, "Related") {
				return ""
			}
			if strings.HasPrefix(attr.Val, "Embed") {
				return ""
			}
		}

		switch node.Data {
		// Checking inner nodes.
		case "div", "p", "a", "span", "u", "ul", "li", "sup", "quote", "blockquote":

		// Ignoring inner nodes.
		case "em", "br", "hr", "h2", "h3", "h4", "strong", "figure", "script", "button", "style", "embed":
			return ""

		default:
			log.Printf("meduza: unknown node data: %s", node.Data)
			return ""
		}

		var src string
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			src += parseNode(c)
		}
		return src
	}

	return ""
}
