package vhs

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

type VideoBase struct {
	core.BaseRecordProxy
}

func NewVideoFromRecord(record *core.Record) Video {
	return &VideoBase{}
}

func (v *VideoBase) Save() error {
	return PocketBase.Save(v)
}

func (v *VideoBase) ID() string {
	//TODO implement me
	panic("implement me")
}

func (v *VideoBase) Name() string {
	//TODO implement me
	panic("implement me")
}

func (v *VideoBase) SetName(s string) {
	//TODO implement me
	panic("implement me")
}

func (v *VideoBase) DefaultPreview() string {
	//TODO implement me
	panic("implement me")
}

func (v *VideoBase) SetDefaultPreview(file *filesystem.File) {
	//TODO implement me
	panic("implement me")
}

func (v *VideoBase) Preview() string {
	//TODO implement me
	panic("implement me")
}

func (v *VideoBase) SetPreview(file *filesystem.File) {
	//TODO implement me
	panic("implement me")
}

func (v *VideoBase) Thumbnails() []string {
	//TODO implement me
	panic("implement me")
}

func (v *VideoBase) SetThumbnails(files []*filesystem.File) {
	//TODO implement me
	panic("implement me")
}

func (v *VideoBase) Video() string {
	return v.GetString("video")
}

func (v *VideoBase) SetVideo(file *filesystem.File) {
	v.Set("video", file)
}
