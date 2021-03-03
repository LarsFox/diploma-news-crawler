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
	maxWorkers int
	total      int
	storage    storage
	wg         sync.WaitGroup
}

// New ...
func New(maxWorkers, total int, storage storage) *Processor {
	return &Processor{
		maxWorkers: maxWorkers,
		total:      total,
		storage:    storage,
		wg:         sync.WaitGroup{},
	}
}

type crawler interface {
	Categories() []string
	Name() string
	PerPage() int

	Category(category string, offset int) ([]*models.Article, error)
	Enrich(article *models.Article) error
}

type storage interface {
	Save(name string, article *models.Article) error
}

// Run запускает все сборщики.
// На каждого сборщика создается икс рутин по запросу и сохранению, плюс по одной на запрос каждой категории.
func (p *Processor) Run(crawlers ...crawler) {
	log.Println("starting processor")

	for _, c := range crawlers {
		ch := make(chan *models.Article, p.maxWorkers)
		log := log.New(os.Stdout, c.Name()+": ", log.Default().Flags())

		var total int64
		for i := 0; i < p.maxWorkers; i++ {
			p.wg.Add(1)
			go p.save(log, c, ch, &total)
		}

		for _, cat := range c.Categories() {
			p.wg.Add(1)
			go p.crawl(log, c, ch, cat, &total)
		}
	}
	p.wg.Wait()
}

// crawl собирает статьи группами по определенной категории и передает их сборщикам.
func (p *Processor) crawl(log *log.Logger, c crawler, ch chan *models.Article, cat string, total *int64) {
	defer p.wg.Done()
	defer close(ch)

	var retries int64
	for offset := 0; int(atomic.LoadInt64(total)) < p.total; offset += c.PerPage() {
		if retries > 100 {
			log.Printf("exhausted, finishing category %s", cat)
			return
		}

		articles, err := c.Category(cat, offset)
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

// save ожидает статьи и сохраняет их.
func (p *Processor) save(log *log.Logger, c crawler, ch chan *models.Article, total *int64) {
	for article := range ch {
		if err := c.Enrich(article); err != nil {
			log.Println(err)
			continue
		}

		if article.Text == "" {
			log.Printf("%s has no text", article.URL)
			continue
		}

		if err := p.storage.Save(c.Name(), article); err != nil {
			log.Println(err)
			continue
		}

		atomic.AddInt64(total, 1)
	}
	p.wg.Done()
}
