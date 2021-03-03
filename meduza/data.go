package meduza

import "regexp"

const (
	urlArticle  = "https://meduza.io/api/v3/%s"
	urlArticles = "https://meduza.io/api/v3/search?chrono=%s&page=%d&per_page=%d&locale=ru"
)

var regSpaces = regexp.MustCompile(` {2,}`)

type respArticles struct {
	Documents map[string]*respDocument `json:"documents"`
}

type respDocument struct {
	PublishedAt int64  `json:"published_at"`
	Title       string `json:"title"`
	URL         string `json:"url"`
}

type respArticle struct {
	Root *respArticleRoot `json:"root"`
}

type respArticleRoot struct {
	Content *respArticleContent `json:"content"`
	Title   string              `json:"title"`
}

type respArticleContent struct {
	Body string `json:"body"`
}
