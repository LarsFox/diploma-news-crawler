package main

import (
	"fmt"
	"log"
	"os"

	"github.com/LarsFox/diploma-news-crawler/meduza"
	"github.com/LarsFox/diploma-news-crawler/processor"
	"github.com/LarsFox/diploma-news-crawler/rt"
	"github.com/LarsFox/diploma-news-crawler/storage"
	"github.com/LarsFox/diploma-news-crawler/tass"
)

var cats = map[string][]string{
	"meduza": {"news"},
	"rt":     {"all-news"},
	"tass":   {"obschestvo", "sport", "kultura", "politika", "ekonomika"},
}

const (
	maxWorkers    = 20
	totalArticles = 4000
)

func main() {
	s, err := storage.New(".tmp", cats)
	if err != nil {
		log.Fatal(err)
	}

	tassP := processor.New("tass", tass.New(), cats["tass"], maxWorkers, 200, 10000, s)
	meduzaP := processor.New("meduza", meduza.New(), cats["meduza"], maxWorkers, 100, 0, s)
	rtP := processor.New("rt", rt.New(), cats["rt"], maxWorkers, 100, 0, s)

	processor.Go(tassP, meduzaP, rtP)

	for name, cc := range cats {
		for _, cat := range cc {
			path := fmt.Sprintf(".tmp/%s/%s", name, cat)
			files, _ := os.ReadDir(path)
			log.Println(path, len(files))
		}
	}
}
