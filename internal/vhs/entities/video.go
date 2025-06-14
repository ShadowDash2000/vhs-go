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

type Info struct {
	Meta     *ffhelp.Probe `json:"meta"`
	Duration float64       `json:"duration"`
	Chapters []*Chapter    `json:"chapters"`
}

type Chapter struct {
	Start int    `json:"start"`
	Title string `json:"title"`
}

func NewChapterFromRaw(c *ChapterRaw) (*Chapter, error) {
	time, err := parsetime.ParseTimeToSeconds(c.Start)
	if err != nil {
		return nil, err
	}

	return &Chapter{
		Start: time,
		Title: c.Title,
	}, nil
}

type ChapterRaw struct {
	_     restructure.Pos
	Start string   `regexp:"(?:(\\d{2}):)?(\\d{2}):(\\d{2})"`
	_     struct{} `regexp:" - "`
	Title string   `regexp:".*"`
	_     restructure.Pos
}
