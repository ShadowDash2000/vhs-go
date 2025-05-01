package collections

import (
	"github.com/pocketbase/pocketbase/core"
	"vhs/pkg/cache"
)

type Collections struct {
	app  core.App
	cols *cache.MemoryCache[string, *core.Collection]
}

func NewCollections(app core.App) *Collections {
	return &Collections{
		app:  app,
		cols: cache.NewMemoryCache[string, *core.Collection](0, true),
	}
}

func (c *Collections) Get(name string) (*core.Collection, error) {
	if col, ok := c.cols.Get(name); ok {
		return col, nil
	}

	col, err := c.fetchCollection(name)
	if err != nil {
		return nil, err
	}

	c.cols.Set(name, col)
	return col, nil
}

func (c *Collections) fetchCollection(name string) (*core.Collection, error) {
	return c.app.FindCollectionByNameOrId(name)
}
