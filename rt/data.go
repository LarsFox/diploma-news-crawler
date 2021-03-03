package rt

const (
	urlArticle  = "https://russian.rt.com%s"
	urlArticles = "https://russian.rt.com/listing/type.News.tag.novosty-glavnoe/prepare/%s/%d/%d"
)

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
