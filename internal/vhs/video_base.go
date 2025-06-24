package vhs

import (
	"github.com/alexflint/go-restructure"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"vhs/internal/vhs/entities"
	"vhs/pkg/ffhelp"
)

type VideoBase struct {
	core.BaseRecordProxy
	info entities.VideoInfo
}

func NewVideoFromRecord(record *core.Record) Video {
	v := &VideoBase{}
	v.SetProxyRecord(record)

	return v
}

func NewVideoFromId(id string) (Video, error) {
	record, err := PocketBase.FindRecordById(entities.VideosCollection, id)
	if err != nil {
		return nil, err
	}

	return NewVideoFromRecord(record), nil
}

func (v *VideoBase) SetProxyRecord(record *core.Record) {
	v.BaseRecordProxy.SetProxyRecord(record)
	v.UnmarshalJSONField("info", &v.info)
}

func (v *VideoBase) Save() error {
	v.parseDescription()

	v.Set("info", v.info)

	return PocketBase.Save(v)
}

func (v *VideoBase) Refresh() error {
	record, err := PocketBase.FindRecordById(entities.VideosCollection, v.Id)
	if err != nil {
		return err
	}

	v.SetProxyRecord(record)

	return nil
}

func (v *VideoBase) parseDescription() {
	regexp := restructure.MustCompile(entities.VideoChapterRaw{}, restructure.Options{})
	var chaptersRaw []*entities.VideoChapterRaw
	regexp.FindAll(&chaptersRaw, v.Description(), -1)

	var chapters []*entities.VideoChapter
	for _, chapterRaw := range chaptersRaw {
		chapter, err := entities.NewChapterFromRaw(chapterRaw)
		if err != nil {
			PocketBase.Logger().Error("parseDescription", map[string]interface{}{
				"chapter": chapter,
				"error":   err,
			})
			continue
		}

		chapters = append(chapters, chapter)
	}

	v.SetChapters(chapters)
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

func (v *VideoBase) Description() string {
	return v.GetString("description")
}

func (v *VideoBase) SetDescription(s string) {
	v.Set("description", s)
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

func (v *VideoBase) Status() entities.Status {
	return entities.Status(v.GetString("status"))
}

func (v *VideoBase) SetStatus(status entities.Status) {
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

func (v *VideoBase) Chapters() *[]*entities.VideoChapter {
	return &v.info.Chapters
}

func (v *VideoBase) SetChapters(chapters []*entities.VideoChapter) {
	v.info.Chapters = chapters
}

func (v *VideoBase) Meta() *ffhelp.Probe {
	return v.info.Meta
}

func (v *VideoBase) SetMeta(meta *ffhelp.Probe) {
	v.info.Meta = meta
}

func (v *VideoBase) Duration() float64 {
	return v.info.Duration
}

func (v *VideoBase) SetDuration(duration float64) {
	v.info.Duration = duration
}

func (v *VideoBase) BaseFilesPath() string {
	return v.BaseRecordProxy.BaseFilesPath()
}
