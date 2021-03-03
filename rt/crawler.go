package rt

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"

	"github.com/LarsFox/diploma-news-crawler/models"
)

// Crawler собирает статьи из Раша-Тудэй.
type Crawler struct {
	*models.BaseCrawler
}

// NewCrawler возвращает новый сборщик статей из Раша-Тудэй.
func NewCrawler(categories []string, perPage int) *Crawler {
	return &Crawler{
		BaseCrawler: models.NewBaseCrawler(categories, "rt", perPage),
	}
}

// Category возвращает статьи из категории.
func (c *Crawler) Category(category string, offset int) ([]*models.Article, error) {
	u := fmt.Sprintf(urlArticles, category, c.PerPage(), offset/c.PerPage())
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
		return nil, errors.New("linksBlock is nil: " + u)
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

// Enrich обогащает данные о статье.
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

	textsBlock := extractNode(document, "article__text")
	if textsBlock == nil {
		return nil
	}

	src := parseNode(textsBlock)
	if src == "" {
		return nil
	}

	src = strings.ReplaceAll(src, ".", ". ")
	article.Text = src
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
		// Проверяем вложенность.
		case "div", "p", "a", "span", "u", "ul", "li", "sup", "quote", "blockquote":

		// Не проверяем вложенность.
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
