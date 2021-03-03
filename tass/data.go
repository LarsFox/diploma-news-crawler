package tass

import (
	"regexp"
)

const (
	urlArticle  = "https://tass.ru/%s/%d/"
	urlArticles = "https://tass.ru/rubric/api/v1/rubric-articles?type=all&slug=%s&tuplesLimit=%d&newsOffset=%d"
)

type respArticles struct {
	Data *respArticlesData `json:"data"`
}

type respArticlesData struct {
	Slug string
	News [][]*respArticlesNews `json:"news"`
}

type respArticlesNews struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
	Theme string `json:"theme"`
}

var regTASS = regexp.MustCompile(`.*/ТАСС/\. `)
