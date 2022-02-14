package rt

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"

	"github.com/LarsFox/diploma-news-crawler/models"
)

const (
	urlArticle  = "https://russian.rt.com%s"
	urlArticles = "https://russian.rt.com/listing/type.News.tag.novosty-glavnoe/prepare/%s/%d/%d"
)

// Crawler works with Russian rt.com.
type Crawler struct{}

// New returns a new crawler.
func New() *Crawler {
	return &Crawler{}
}

// Category lists articles in a category.
func (c *Crawler) Category(category string, offset, perPage int) ([]*models.Article, error) {
	u := fmt.Sprintf(urlArticles, category, perPage, offset/perPage)
	resp, err := http.DefaultClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("error getting category: %w", err)
	}

	document, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error decoding category: %w", err)
	}
	defer resp.Body.Close()

	linksBlock := extractNode(document, "listing__rows")
	if linksBlock == nil {
		return nil, fmt.Errorf("linksBlock is nil: %s", u)
	}

	cards := gatherNodes(linksBlock, "card__heading")

	articles := make([]*models.Article, 0, len(cards))
	for _, card := range cards {
		for _, attr := range card.FirstChild.Attr {
			if attr.Key != "href" {
				continue
			}

			// /world/news/838374-franciya-koronavirus-statistika
			parts := strings.Split(attr.Val, "/")
			articles = append(articles, &models.Article{
				Category: category,
				ID:       fmt.Sprintf("%s-%s", parts[1], parts[3][0:6]),
				URL:      fmt.Sprintf(urlArticle, attr.Val),
			})
		}
	}
	return articles, nil
}

// Enrich parses article text.
func (c *Crawler) Enrich(article *models.Article) error {
	resp, err := http.DefaultClient.Get(article.URL)
	if err != nil {
		return fmt.Errorf("err getting article: %w", err)
	}

	defer resp.Body.Close()
	document, err := html.Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("err parsing page: %w", err)
	}

	textsBlock := mustNode(
		extractNode(document, "article__text"),
		extractNode(document, "News-view"),
	)

	if textsBlock == nil {
		return nil
	}

	h1 := extractHeader(document)
	if h1 == "" {
		return fmt.Errorf("no header")
	}

	src := parseNode(textsBlock)
	if src == "" {
		return fmt.Errorf("no text")
	}

	src = h1 + "\n\n" + src
	src = strings.ReplaceAll(src, ".", ". ")
	article.Text = src
	return nil
}

func extractHeader(document *html.Node) string {
	h1 := mustNode(
		extractNode(document, "News-view_title"),
		extractNode(document, "article__heading"),
	)

	if h1 == nil {
		return ""
	}
	if h1.FirstChild != nil && h1.FirstChild.Type == html.TextNode {
		return h1.FirstChild.Data
	}
	return ""
}

func mustNode(nodes ...*html.Node) *html.Node {
	for _, node := range nodes {
		if node != nil {
			return node
		}
	}
	return nil
}

func extractNode(node *html.Node, className string) *html.Node {
	for _, attr := range node.Attr {
		if attr.Key == "class" && strings.Contains(attr.Val, className) {
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

func gatherNodes(node *html.Node, className string) []*html.Node {
	nodes := []*html.Node{}
	for _, attr := range node.Attr {
		if attr.Key == "class" && strings.Contains(attr.Val, className) {
			nodes = append(nodes, node)
			break
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		cs := gatherNodes(c, className)
		if len(cs) > 0 {
			nodes = append(nodes, cs...)
		}
	}
	return nodes
}

func parseNode(node *html.Node) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case html.TextNode:
		return node.Data

	case html.ElementNode:
		switch node.Data {
		// Checking inner nodes.
		case "div", "p", "a", "span", "u", "ul", "li", "sup", "quote", "blockquote":

		// Ignoring inner nodes.
		case "img", "em", "br", "hr", "h1", "h2", "h3", "h4", "strong", "figure", "script", "iframe":
			return ""

		default:
			log.Printf("rt: unknown node data: %s", node.Data)
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
