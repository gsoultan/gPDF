package doc

import "github.com/gsoultan/gpdf/model"

// ContentSearcher searches for keywords and replaces text within content streams.
type ContentSearcher interface {
	Search(keywords ...string) ([]model.SearchResult, error)
	Replace(old, new string) error
	Replaces(replacements map[string]string) error
}
