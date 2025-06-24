package vhs

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

type Playlist interface {
	core.RecordProxy
	Save() error
	Delete() error
	ID() string
	Name() string
	SetName(string)
	User() string
	SetUser(string)
	Videos() []string
	SetVideos([]string)
	AddVideo(string)
	AddVideos([]string)
	RemoveVideo(string)
	Preview() string
	SetPreview(*filesystem.File)
}
