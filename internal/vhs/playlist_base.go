package vhs

import (
	"fmt"
	"github.com/pocketbase/pocketbase/core"
	"golang.org/x/exp/slices"
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

func NewPlaylistFromIds(ids []string) ([]Playlist, error) {
	records, err := PocketBase.FindRecordsByIds(entities.PlaylistsCollection, ids)
	if err != nil {
		return nil, err
	}

	playlists := make([]Playlist, len(records))
	for i, record := range records {
		playlists[i] = NewPlaylistFromRecord(record)
	}
	return playlists, nil
}

func NewPlaylistsFromVideoId(videoId string) ([]Playlist, error) {
	records, err := PocketBase.FindRecordsByFilter(
		entities.PlaylistsCollection,
		fmt.Sprintf("videos.id ?= '%s'", videoId),
		"",
		0,
		0,
	)
	if err != nil {
		return nil, err
	}

	playlists := make([]Playlist, len(records))
	for i, record := range records {
		playlists[i] = NewPlaylistFromRecord(record)
	}
	return playlists, nil
}

func (p *PlaylistBase) Save() error {
	return PocketBase.Save(p)
}

func (p *PlaylistBase) Delete() error {
	return PocketBase.Delete(p)
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

func (p *PlaylistBase) AddVideo(id string) {
	p.Set("videos", append(p.Videos(), id))
}

func (p *PlaylistBase) AddVideos(ids []string) {
	p.Set("videos", append(p.Videos(), ids...))
}

func (p *PlaylistBase) RemoveVideo(id string) {
	p.SetVideos(
		slices.DeleteFunc(p.Videos(), func(s string) bool {
			return s == id
		}),
	)
}
