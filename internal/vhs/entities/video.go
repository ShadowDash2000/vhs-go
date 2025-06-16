package entities

import (
	"github.com/alexflint/go-restructure"
	"vhs/pkg/ffhelp"
	"vhs/pkg/parsetime"
)

type Status string

const (
	StatusPublic Status = "public"
	StatusLink          = "link"
	StatusClosed        = "closed"
)

type VideoInfo struct {
	Meta     *ffhelp.Probe   `json:"meta"`
	Duration float64         `json:"duration"`
	Chapters []*VideoChapter `json:"chapters"`
}

type VideoChapter struct {
	Start int    `json:"start"`
	Title string `json:"title"`
}

func NewChapterFromRaw(c *VideoChapterRaw) (*VideoChapter, error) {
	time, err := parsetime.ParseTimeToSeconds(c.Start)
	if err != nil {
		return nil, err
	}

	return &VideoChapter{
		Start: time,
		Title: c.Title,
	}, nil
}

type VideoChapterRaw struct {
	_     restructure.Pos
	Start string   `regexp:"(?:(\\d{2}):)?(\\d{2}):(\\d{2})"`
	_     struct{} `regexp:" - "`
	Title string   `regexp:".*"`
	_     restructure.Pos
}
