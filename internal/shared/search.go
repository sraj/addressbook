package shared

import "context"

type SearchIndex interface {
	IndexDocument(ctx context.Context, index string, id string, document interface{}) error
	Search(ctx context.Context, index string, query string, from, size int) ([]string, int64, error)
	DeleteDocument(ctx context.Context, index string, id string) error
}
