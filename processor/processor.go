package processor

import (
	"log"
	"os"
	"sync"
	"sync/atomic"

	"github.com/LarsFox/diploma-news-crawler/models"
)

// Processor ...
type Processor struct {
	categories   []string
	crawler      crawler
	maxWorkers   int
	name         string
	perPage      int
	storage      storage
	total        int
	wgArticles   sync.WaitGroup
	wgCategories sync.WaitGroup
}

// New returns a new processor for news crawling.
func New(name string, crawler crawler, categories []string, maxWorkers, perPage, total int, storage storage) *Processor {
	return &Processor{
		categories:   categories,
		crawler:      crawler,
		maxWorkers:   maxWorkers,
		name:         name,
		perPage:      perPage,
		storage:      storage,
		total:        total,
		wgArticles:   sync.WaitGroup{},
		wgCategories: sync.WaitGroup{},
	}
}

type crawler interface {
	Category(category string, offset, perPage int) ([]*models.Article, error)
	Enrich(article *models.Article) error
}

type storage interface {
	Save(name string, article *models.Article) error
}

// Go runs all processors concurrently.
func Go(processors ...*Processor) {
	wg := &sync.WaitGroup{}
	wg.Add(len(processors))

	for _, p := range processors {
		p := p
		go func() {
			defer wg.Done()
			p.run()
		}()
	}
	wg.Wait()
}

// Run starts news crawling.
// Each processor runs x goroutines, where x == maxWorkers+len(categories).
func (p *Processor) run() {
	log.Printf("Starting processor %s", p.name)

	ch := make(chan *models.Article, p.maxWorkers)

	var total int64
	for i := 0; i < p.maxWorkers; i++ {
		p.wgArticles.Add(1)
		go p.save(ch, &total)
	}

	for _, cat := range p.categories {
		p.wgCategories.Add(1)
		go p.crawl(ch, cat, &total)
	}

	p.wgCategories.Wait()

	close(ch)

	p.wgArticles.Wait()
}

// crawl passes articles from categories to workers.
func (p *Processor) crawl(ch chan *models.Article, cat string, total *int64) {
	defer p.wgCategories.Done()
	log := log.New(os.Stdout, p.name+"-categories: ", log.Default().Flags())

	var retries int64
	for offset := 0; int(atomic.LoadInt64(total)) < p.total; offset += p.perPage {
		if retries > 100 {
			log.Printf("exhausted, finishing category %s", cat)
			return
		}

		articles, err := p.crawler.Category(cat, offset, p.perPage)
		if err != nil {
			retries++
			log.Println(err)
			continue
		}

		if len(articles) == 0 {
			log.Printf("got none, finishing category %s", cat)
			return
		}

		for _, a := range articles {
			ch <- a
		}
	}
}

// save enriches the article and passes it to storage.
func (p *Processor) save(ch chan *models.Article, total *int64) {
	defer p.wgArticles.Done()
	log := log.New(os.Stdout, p.name+"-articles: ", log.Default().Flags())

	for article := range ch {
		if err := p.crawler.Enrich(article); err != nil {
			log.Printf("%s err enriching err: %v", article.URL, err)
			continue
		}

		if article.Text == "" {
			log.Printf("%s has no text", article.URL)
			continue
		}

		if err := p.storage.Save(p.name, article); err != nil {
			log.Println(err)
			continue
		}

		atomic.AddInt64(total, 1)
	}
}
