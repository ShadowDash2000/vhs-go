package vhs

import "github.com/pocketbase/pocketbase/core"

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
}
