package models

// Article — статья.
type Article struct {
	Category string
	ID       string
	Text     string
	URL      string
}

// BaseCrawler — заготовка сборщика.
type BaseCrawler struct {
	categories []string
	name       string
	perPage    int
}

// NewBaseCrawler возвращает заготовку сборщика.
func NewBaseCrawler(categories []string, name string, perPage int) *BaseCrawler {
	return &BaseCrawler{categories: categories, name: name, perPage: perPage}
}

// Categories возвращает категории новостей для сборщика.
func (c *BaseCrawler) Categories() []string {
	return c.categories
}

// Name возвращает название сборщика.
func (c *BaseCrawler) Name() string {
	return c.name
}

// PerPage возвращает количество новостей в одном запросе к категории.
func (c *BaseCrawler) PerPage() int {
	return c.perPage
}
