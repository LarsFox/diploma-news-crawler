package models

// Article is a base model for a news article.
type Article struct {
	Category string
	ID       string
	Text     string
	URL      string
}
