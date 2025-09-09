package vhs

import (
	"vhs/internal/vhs/entities"
	"vhs/pkg/ffhelp"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

type Video interface {
	core.RecordProxy
	Save() error
	Refresh() error
	ID() string
	Name() string
	SetName(string)
	Description() string
	SetDescription(string)
	Preview() string
	SetPreview(*filesystem.File)
	Thumbnails() []string
	SetThumbnails([]*filesystem.File)
	Video() string
	SetVideo(*filesystem.File)
	SetVideoPath(string)
	Status() entities.Status
	SetStatus(entities.Status)
	User() string
	SetUser(string)
	WebVTT() string
	SetWebVTT(*filesystem.File)
	Chapters() *[]*entities.VideoChapter
	SetChapters([]*entities.VideoChapter)
	Meta() *ffhelp.Probe
	SetMeta(*ffhelp.Probe)
	Duration() float64
	SetDuration(float64)
	BaseFilesPath() string
	PreviewIsSet() bool
}
