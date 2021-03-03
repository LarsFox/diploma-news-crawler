package main

import (
	"log"

	"github.com/LarsFox/diploma-news-crawler/meduza"
	"github.com/LarsFox/diploma-news-crawler/processor"
	"github.com/LarsFox/diploma-news-crawler/rt"
	"github.com/LarsFox/diploma-news-crawler/storage"
)

var cats = map[string][]string{
	"meduza":    {"news"},
	"rt":        {"all-news"},
	"tass":      {"kultura"},
	"tass_full": {"obschestvo", "sport", "kultura", "politika", "ekonomika"},
}

const (
	maxWorkers    = 20
	totalArticles = 2000
)

func main() {
	s, err := storage.New(".tmp", cats)
	if err != nil {
		log.Fatal(err)
	}
	processor := processor.New(maxWorkers, totalArticles, s)

	// tassC := tass.NewCrawler(cats["tass"], 200)
	meduzaC := meduza.NewCrawler(cats["meduza"], 100)
	rtC := rt.NewCrawler(cats["rt"], 100)

	processor.Run(meduzaC, rtC)
}
