package tests

import (
	"github.com/pocketbase/pocketbase/core"
	"io"
	"log/slog"
	"os"
	"testing"
	"vhs/internal/vhs"
	"vhs/internal/vhs/entities"
	"vhs/pkg/ffhelp"
)

func TestCreateStoryBoard(t *testing.T) {
	defer os.RemoveAll(vhs.UploadDir)

	col, err := Collections.Get(entities.VideosCollection)
	if err != nil {
		t.Fatal(err)
	}

	record := core.NewRecord(col)
	videoMock := NewVideoBaseMockFromRecord(record)
	videoMock.Id = "testId"

	ffhelp, err := ffhelp.Input("assets/black_30m.mp4")
	if err != nil {
		t.Fatal(err)
	}

	videoUploader := vhs.NewVideoUploaderMock(&vhs.VideoUploaderBaseMock{
		TmpFile:      nil,
		Ffhelp:       ffhelp,
		BytesWritten: 0,
		Data:         nil,
		Video:        videoMock,
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	err = videoUploader.CreateSprites()
	if err != nil {
		t.Error(err)
	}
}
