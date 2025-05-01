package vhs

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

type Video interface {
	core.RecordProxy
	Save() error
	ID() string
	Name() string
	SetName(string)
	DefaultPreview() string
	SetDefaultPreview(*filesystem.File)
	Preview() string
	SetPreview(*filesystem.File)
	Thumbnails() []string
	SetThumbnails([]*filesystem.File)
	Video() string
	SetVideo(*filesystem.File)
}
