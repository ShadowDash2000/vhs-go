package vhs

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"os"
)

type VideoUploaderBase struct {
	tmpFile      *os.File
	started      bool
	bytesWritten int
	data         *VideoUploadData
	video        Video
}

func NewVideoUploader() (VideoUploader, error) {
	file, err := os.CreateTemp("upload/video", "video_")
	if err != nil {
		return nil, err
	}

	col, err := Collections.Get(VideosCollection)
	if err != nil {
		return nil, err
	}

	video := &VideoBase{}
	record := core.NewRecord(col)
	video.SetProxyRecord(record)
	if err = video.Save(); err != nil {
		return nil, err
	}

	return &VideoUploaderBase{
		tmpFile: file,
		video:   video,
	}, nil
}

func (v *VideoUploaderBase) Start(data *VideoUploadData) error {
	if v.started {
		return nil
	}

	v.data = data

	v.video.SetName(data.FileName)
	if err := v.video.Save(); err != nil {
		return err
	}

	v.started = true
	return nil
}

func (v *VideoUploaderBase) UploadPart(p []byte) (bool, error) {
	if !v.started {
		return false, nil
	}

	n, err := v.tmpFile.Write(p)
	if err != nil {
		return false, err
	}
	v.bytesWritten += n

	done := false
	if v.bytesWritten >= v.data.FileSize {
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
	v.video.Save()

	return v.clear()
}

func (v *VideoUploaderBase) clear() error {
	if !v.started {
		return nil
	}

	err := v.tmpFile.Close()
	if err != nil {
		return err
	}

	return nil
}
