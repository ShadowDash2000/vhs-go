package vhs

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"os"
	"vhs/internal/vhs/entities"
	"vhs/pkg/errorcollector"
	"vhs/pkg/ffmpegthumbs"
	"vhs/pkg/webvtt"
)

type VideoUploaderBase struct {
	tmpFile      *os.File
	bytesWritten int
	data         *VideoUploadData
	video        Video
	duration     float64
}

const (
	UploadDir = "upload/video"
	ThumbsDir = UploadDir + "/thumbs"
	WebVTTDir = UploadDir + "/webvtt"

	FrameDuration = 5
)

func NewVideoUploader() VideoUploader {
	return &VideoUploaderBase{}
}

func (v *VideoUploaderBase) Start(data *VideoUploadData) (string, error) {
	err := os.MkdirAll(UploadDir, 0755)
	if err != nil {
		return "", err
	}

	file, err := os.CreateTemp(UploadDir, "video_")
	if err != nil {
		return "", err
	}

	col, err := Collections.Get(entities.VideosCollection)
	if err != nil {
		return "", err
	}

	record := core.NewRecord(col)
	video := NewVideoFromRecord(record)
	video.SetStatus(entities.StatusClosed)
	video.SetUser(data.UserId)
	video.SetName(data.Name)

	if err = video.Save(); err != nil {
		return "", err
	}

	v.tmpFile = file
	v.video = video
	v.data = data

	return video.ID(), nil
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

func (v *VideoUploaderBase) Done() {
	go v.done()
}

func (v *VideoUploaderBase) done() error {
	var err error
	defer func() {
		if err != nil {
			PocketBase.Logger().Error(
				"error while video processing: "+err.Error(),
				map[string]any{
					"video": v.video,
				},
			)
		}

		if err = v.clear(); err != nil {
			PocketBase.Logger().Error(
				"error while clearing video files: "+err.Error(),
				v.video,
			)
		}
	}()

	if err = v.video.Refresh(); err != nil {
		return err
	}
	if v.duration, err = ffmpegthumbs.GetVideoDurationFloat(v.tmpFile.Name()); err != nil {
		return err
	}
	if err = v.SetDuration(); err != nil {
		return err
	}
	if err = v.SaveVideoFile(); err != nil {
		return err
	}
	if err = v.CreateStoryBoard(); err != nil {
		return err
	}
	if err = v.CreateWebVTT(); err != nil {
		return err
	}
	if err = v.SetDefaultPreview(); err != nil {
		return err
	}

	return nil
}

func (v *VideoUploaderBase) clear() error {
	ec := errorcollector.NewErrorCollector()

	defer func() {
		if ec.HasErrors() {
			PocketBase.Logger().Error(
				"error while clearing video files: "+ec.Error().Error(),
				map[string]any{
					"video": v.video,
				},
			)
		}
	}()

	ec.Collect(func() error {
		return v.tmpFile.Close()
	})
	ec.Collect(func() error {
		return os.Remove(v.tmpFile.Name())
	})
	ec.Collect(func() error {
		return os.RemoveAll(ThumbsDir + "/" + v.video.ID())
	})
	ec.Collect(func() error {
		return os.RemoveAll(WebVTTDir + "/" + v.video.ID())
	})

	return ec.Error()
}

func (v *VideoUploaderBase) SaveVideoFile() error {
	file, err := filesystem.NewFileFromPath(v.tmpFile.Name())
	if err != nil {
		return err
	}

	v.video.SetVideo(file)

	return v.video.Save()
}

func (v *VideoUploaderBase) CreateStoryBoard() error {
	basePath := ThumbsDir + "/" + v.video.ID()

	err := ffmpegthumbs.SplitVideoToThumbnails(
		v.tmpFile.Name(),
		basePath,
		FrameDuration,
	)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}

	var files []*filesystem.File
	for _, entry := range entries {
		file, err := filesystem.NewFileFromPath(basePath + "/" + entry.Name())
		if err != nil {
			return err
		}
		files = append(files, file)
	}

	v.video.SetThumbnails(files)

	return v.video.Save()
}

func (v *VideoUploaderBase) CreateWebVTT() error {
	var filePaths []string
	for _, fileId := range v.video.Thumbnails() {
		filePaths = append(filePaths, "/api/files/"+v.video.ProxyRecord().BaseFilesPath()+"/"+fileId)
	}

	file, err := webvtt.CreateFromFilePaths(
		filePaths,
		WebVTTDir+"/"+v.video.ID(),
		int(v.duration),
		FrameDuration,
	)
	if err != nil {
		return err
	}

	f, err := filesystem.NewFileFromPath(file.Name())
	if err != nil {
		return err
	}

	v.video.SetWebVTT(f)

	return v.video.Save()
}

func (v *VideoUploaderBase) SetDefaultPreview() error {
	// If the preview isn't empty, this means it was set by the user.
	if v.video.Preview() != "" {
		return nil
	}

	basePath := ThumbsDir + "/" + v.video.ID()
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return nil
	}

	thumb := entries[len(entries)/2]
	f, err := filesystem.NewFileFromPath(basePath + "/" + thumb.Name())
	if err != nil {
		return err
	}

	v.video.SetPreview(f)

	return v.video.Save()
}

func (v *VideoUploaderBase) SetDuration() error {
	v.video.SetDuration(v.duration)
	return v.video.Save()
}
