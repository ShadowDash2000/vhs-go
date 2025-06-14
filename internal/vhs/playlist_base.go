package vhs

import (
	"github.com/pocketbase/pocketbase/core"
	"vhs/internal/vhs/entities"
)

type PlaylistBase struct {
	core.BaseRecordProxy
}

func NewPlaylist() (Playlist, error) {
	col, err := Collections.Get(entities.PlaylistsCollection)
	if err != nil {
		return nil, err
	}

	return NewPlaylistFromRecord(core.NewRecord(col)), nil
}

func NewPlaylistFromRecord(record *core.Record) Playlist {
	p := &PlaylistBase{}
	p.SetProxyRecord(record)

	return p
}

func NewPlaylistFromId(id string) (Playlist, error) {
	record, err := PocketBase.FindRecordById(entities.PlaylistsCollection, id)
	if err != nil {
		return nil, err
	}

	return NewPlaylistFromRecord(record), nil
}

func (p *PlaylistBase) Save() error {
	return p.Save()
}

func (p *PlaylistBase) Delete() error {
	return p.Delete()
}

func (p *PlaylistBase) ID() string {
	return p.Id
}

func (p *PlaylistBase) Name() string {
	return p.GetString("name")
}

func (p *PlaylistBase) SetName(s string) {
	p.Set("name", s)
}

func (p *PlaylistBase) User() string {
	return p.GetString("user")
}

func (p *PlaylistBase) SetUser(id string) {
	p.Set("user", id)
}

func (p *PlaylistBase) Videos() []string {
	return p.GetStringSlice("videos")
}

func (p *PlaylistBase) SetVideos(ids []string) {
	p.Set("videos", ids)
}
