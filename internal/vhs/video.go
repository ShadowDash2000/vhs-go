package vhs

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"vhs/internal/vhs/entities"
)

type Video interface {
	core.RecordProxy
	Save() error
	ID() string
	Name() string
	SetName(string)
	Description() string
	SetDescription(string)
	Preview() string
	SetPreview(*filesystem.File)
	SetPreviewFromId(string)
	Thumbnails() []string
	SetThumbnails([]*filesystem.File)
	Video() string
	SetVideo(*filesystem.File)
	Status() entities.Status
	SetStatus(entities.Status)
	User() string
	SetUser(string)
	WebVTT() string
	SetWebVTT(*filesystem.File)
	Chapters() *[]*entities.Chapter
	SetChapters([]*entities.Chapter)
	Duration() float64
	SetDuration(float64)
}
