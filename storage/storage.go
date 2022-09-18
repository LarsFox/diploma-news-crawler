package storage

import (
	"fmt"
	"os"

	"github.com/LarsFox/diploma-news-crawler/models"
)

// Storage is a plain txt file saver.
type Storage struct {
	dir string
}

const (
	dirPath      = "%s/%s/%s"
	saveFilePath = dirPath + "/%s.txt"
)

// New returns a new storage client and creates all the necessary directories for the future files.
func New(dir string, crawlCategories map[string][]string) (*Storage, error) {
	for crawler, cats := range crawlCategories {
		for _, cat := range cats {
			if err := os.MkdirAll(fmt.Sprintf(dirPath, dir, crawler, cat), 0744); err != nil {
				return nil, err
			}
		}
	}
	return &Storage{dir: dir}, nil
}

// Save creates a new .txt file in a specific category with the article text.
func (s *Storage) Save(name string, article *models.Article) error {
	articlePath := fmt.Sprintf(saveFilePath, s.dir, name, article.Category, article.ID)
	if err := os.WriteFile(articlePath, []byte(article.Text), 0744); err != nil {
		return err
	}
	return nil
}
