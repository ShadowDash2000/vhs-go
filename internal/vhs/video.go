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
	Preview() string
	SetPreview(*filesystem.File)
	Thumbnails() []string
	SetThumbnails([]*filesystem.File)
	Video() string
	SetVideo(*filesystem.File)
	Status() Status
	SetStatus(Status)
	User() string
	SetUser(string)
	WebVTT() string
	SetWebVTT(*filesystem.File)
}

type Status string

const (
	StatusPublic Status = "public"
	StatusLink          = "link"
	StatusClosed        = "closed"
)
