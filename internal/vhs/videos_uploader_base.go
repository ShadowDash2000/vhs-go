package vhs

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"os"
)

type VideoUploaderBase struct {
	tmpFile      *os.File
	bytesWritten int
	data         *VideoUploadData
	video        Video
}

const UploadDir = "upload/video"

func NewVideoUploader() VideoUploader {
	return &VideoUploaderBase{}
}

func (v *VideoUploaderBase) Start(data *VideoUploadData) error {
	err := os.MkdirAll(UploadDir, 0755)
	if err != nil {
		return err
	}

	file, err := os.CreateTemp(UploadDir, "video_")
	if err != nil {
		return err
	}

	col, err := Collections.Get(VideosCollection)
	if err != nil {
		return err
	}

	record := core.NewRecord(col)
	video := NewVideoFromRecord(record)
	video.SetStatus(StatusClosed)
	video.SetUser(data.UserId)
	video.SetName(data.Name)

	if err = video.Save(); err != nil {
		return err
	}

	v.tmpFile = file
	v.video = video
	v.data = data

	return nil
}

func (v *VideoUploaderBase) UploadPart(p []byte) (bool, error) {
	n, err := v.tmpFile.Write(p)
	if err != nil {
		return false, err
	}
	v.bytesWritten += n

	done := false
	if v.bytesWritten >= v.data.Size {
		done = true
	}

	return done, nil
}

func (v *VideoUploaderBase) Cancel() error {
	return v.clear()
}

func (v *VideoUploaderBase) Done() error {
	file, err := filesystem.NewFileFromPath(v.tmpFile.Name())
	if err != nil {
		return err
	}

	v.video.SetVideo(file)
	err = v.video.Save()

	v.clear()

	return err
}

func (v *VideoUploaderBase) clear() error {
	err := v.tmpFile.Close()
	if err != nil {
		return err
	}

	err = os.Remove(v.tmpFile.Name())
	if err != nil {
		return err
	}

	return nil
}
