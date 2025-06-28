package vhs

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"log/slog"
	"os"
	"vhs/internal/assets"
	"vhs/internal/vhs/entities"
	"vhs/pkg/errorcollector"
	"vhs/pkg/ffhelp"
	"vhs/pkg/webvtt"
)

type VideoUploaderBase struct {
	tmpFile      *os.File
	ffhelp       *ffhelp.FFHelp
	bytesWritten int
	data         *VideoUploadData
	video        Video
	logger       *slog.Logger
}

const (
	UploadDir = "upload/video"
	ThumbsDir = UploadDir + "/thumbs"
	WebVTTDir = UploadDir + "/webvtt"

	FrameDuration = 5
)

func NewVideoUploader(logger *slog.Logger) VideoUploader {
	return &VideoUploaderBase{
		logger: logger,
	}
}

type VideoUploaderBaseMock struct {
	TmpFile      *os.File
	Ffhelp       *ffhelp.FFHelp
	BytesWritten int
	Data         *VideoUploadData
	Video        Video
	Logger       *slog.Logger
}

func NewVideoUploaderMock(v *VideoUploaderBaseMock) *VideoUploaderBase {
	return &VideoUploaderBase{
		tmpFile:      v.TmpFile,
		ffhelp:       v.Ffhelp,
		bytesWritten: v.BytesWritten,
		data:         v.Data,
		video:        v.Video,
		logger:       v.Logger,
	}
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
	preview, err := filesystem.NewFileFromBytes(assets.DefaultPreview, "default_preview")
	if err != nil {
		return "", err
	}
	video.SetPreview(preview)

	if err = video.Save(); err != nil {
		return "", err
	}

	v.ffhelp = ffhelp.Input(file.Name())
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
			v.logger.Error(
				"error while video processing: "+err.Error(),
				"video", v.video,
			)
		}

		if err = v.clear(); err != nil {
			v.logger.Error(
				"error while clearing video files: "+err.Error(),
				"video", v.video,
			)
		}
	}()

	if err = v.video.Refresh(); err != nil {
		return err
	}
	if err = v.SetMeta(); err != nil {
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
			v.logger.Error(
				"error while clearing video files: "+ec.Error().Error(),
				"video", v.video,
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

	err := v.ffhelp.SplitVideoToThumbnails(
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
		int(v.ffhelp.GetVideoDuration()),
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
	if v.video.PreviewIsSet() {
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

func (v *VideoUploaderBase) SetMeta() error {
	v.video.SetMeta(v.ffhelp.Probe())
	return v.video.Save()
}

func (v *VideoUploaderBase) SetDuration() error {
	v.video.SetDuration(v.ffhelp.GetVideoDuration())
	return v.video.Save()
}
