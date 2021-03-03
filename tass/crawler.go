package tass

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/LarsFox/diploma-news-crawler/models"
)

// Crawler собирает статьи из ТАССа.
type Crawler struct {
	*models.BaseCrawler
}

// NewCrawler возвращает новый сборщик статей из ТАССа.
func NewCrawler(categories []string, perPage int) *Crawler {
	return &Crawler{
		BaseCrawler: models.NewBaseCrawler(categories, "tass", perPage),
	}
}

// Category возвращает статьи из категории.
func (c *Crawler) Category(category string, offset int) ([]*models.Article, error) {
	u := fmt.Sprintf(urlArticles, category, c.PerPage(), offset)
	resp, err := http.DefaultClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("error getting category: %w", err)
	}

	articles := &respArticles{}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(articles); err != nil {
		return nil, fmt.Errorf("error decoding category: %w", err)
	}

	result := make([]*models.Article, 0, len(articles.Data.News))
	for _, article := range articles.Data.News {
		article := newArticle(articles.Data.Slug, article[0])
		if article == nil {
			continue
		}
		result = append(result, article)
	}
	return result, nil
}

func newArticle(category string, item *respArticlesNews) *models.Article {
	if strings.Contains(item.Theme, "карто") {
		return nil
	}
	if item.Slug != category {
		return nil
	}
	return &models.Article{
		Category: item.Slug,
		ID:       strconv.FormatInt(item.ID, 10),
		URL:      fmt.Sprintf(urlArticle, item.Slug, item.ID),
	}
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

	textsBlock := extractNode(document, "text-content")
	if textsBlock == nil {
		return nil
	}

	src := parseNode(textsBlock)
	if src == "" {
		return nil
	}

	article.Text = regTASS.ReplaceAllString(src, "")
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

	// log.Println(node.Type, node.Data)
	switch node.Type {
	case html.TextNode:
		return node.Data

	case html.ElementNode:
		switch node.Data {
		case "div":
			for _, attr := range node.Attr {
				if attr.Key == "class" && strings.HasPrefix(attr.Val, "text-include") {
					return ""
				}
			}

		// Проверяем вложенность.
		case "p", "a", "span", "u", "ul", "li", "sup":

		// Не проверяем вложенность.
		case "em", "br", "hr", "h2", "strong":
			return ""

		default:
			log.Printf("tass: unknown node data: %s", node.Data)
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
