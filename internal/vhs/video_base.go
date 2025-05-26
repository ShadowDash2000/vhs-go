package vhs

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

type VideoBase struct {
	core.BaseRecordProxy
}

func NewVideoFromRecord(record *core.Record) Video {
	v := &VideoBase{}
	v.SetProxyRecord(record)
	return v
}

func (v *VideoBase) Save() error {
	return PocketBase.Save(v)
}

func (v *VideoBase) ID() string {
	return v.Id
}

func (v *VideoBase) Name() string {
	return v.GetString("name")
}

func (v *VideoBase) SetName(s string) {
	v.Set("name", s)
}

func (v *VideoBase) Preview() string {
	return v.GetString("preview")
}

func (v *VideoBase) SetPreview(file *filesystem.File) {
	v.Set("preview", file)
}

func (v *VideoBase) Thumbnails() []string {
	return v.GetStringSlice("thumbnails")
}

func (v *VideoBase) SetThumbnails(files []*filesystem.File) {
	v.Set("thumbnails", files)
}

func (v *VideoBase) Video() string {
	return v.GetString("video")
}

func (v *VideoBase) SetVideo(file *filesystem.File) {
	v.Set("video", file)
}

func (v *VideoBase) Status() Status {
	return Status(v.GetString("status"))
}

func (v *VideoBase) SetStatus(status Status) {
	v.Set("status", string(status))
}

func (v *VideoBase) User() string {
	return v.GetString("user")
}

func (v *VideoBase) SetUser(user string) {
	v.Set("user", user)
}

func (v *VideoBase) WebVTT() string {
	return v.GetString("webvtt")
}

func (v *VideoBase) SetWebVTT(file *filesystem.File) {
	v.Set("webvtt", file)
}
