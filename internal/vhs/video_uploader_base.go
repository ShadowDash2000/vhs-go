package vhs

import (
	"fmt"
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
	UploadDir       = "upload/video"
	ThumbsDir       = UploadDir + "/thumbs"
	SpriteSheetsDir = UploadDir + "/sheets"
	WebVTTDir       = UploadDir + "/webvtt"

	FrameDuration = 5

	DefaultPreviewWidth  = 1280
	DefaultPreviewHeight = 720

	SpriteSheetCols     = 10
	SpriteSheetRows     = 40
	SpriteWidth         = 180
	SpriteHeight        = 101
	SpriteSheetImgCount = SpriteSheetCols * SpriteSheetRows
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

	if v.ffhelp, err = ffhelp.Input(v.tmpFile.Name()); err != nil {
		return err
	}
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
	if err = v.CreateSprites(); err != nil {
		return err
	}
	if err = v.CreateSpriteSheet(); err != nil {
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
		return os.RemoveAll(v.thumbsDir())
	})
	ec.Collect(func() error {
		return os.RemoveAll(v.sheetsDir())
	})
	ec.Collect(func() error {
		return os.RemoveAll(v.webvttDir())
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

func (v *VideoUploaderBase) CreateSprites() error {
	err := v.ffhelp.SplitVideoToThumbnails(
		v.thumbsDir(),
		FrameDuration,
		SpriteWidth,
		SpriteHeight,
	)
	if err != nil {
		return err
	}

	return v.video.Save()
}

func (v *VideoUploaderBase) CreateSpriteSheet() error {
	entries, err := v.readThumbsDir()

	sheetsCount := (len(entries) + SpriteSheetImgCount - 1) / SpriteSheetImgCount
	filePaths := make([][]string, sheetsCount)
	i := 0
	for j, entry := range entries {
		filePaths[i] = append(filePaths[i], v.thumbsDir()+"/"+entry.Name())

		if j%SpriteSheetImgCount == SpriteSheetImgCount-1 {
			i++
		}
	}

	var files []*filesystem.File
	for i := 0; i < sheetsCount; i++ {
		sheetPath := fmt.Sprintf("%s/sheet%06d.jpg", v.sheetsDir(), i)

		err = webvtt.CreateSpriteSheet(
			filePaths[i],
			sheetPath,
			SpriteSheetCols, SpriteSheetRows, SpriteWidth, SpriteHeight,
		)
		if err != nil {
			return err
		}

		f, err := filesystem.NewFileFromPath(sheetPath)
		if err != nil {
			return err
		}

		files = append(files, f)
	}

	v.video.SetThumbnails(files)

	return v.video.Save()
}

func (v *VideoUploaderBase) CreateWebVTT() error {
	var sheetPaths []string
	for _, imgId := range v.video.Thumbnails() {
		sheetPaths = append(sheetPaths, "/api/files/"+v.video.BaseFilesPath()+"/"+imgId)
	}

	outFile := fmt.Sprintf("%s/webvtt.vtt", v.webvttDir())
	file, err := webvtt.CreateFromSheets(
		sheetPaths,
		outFile,
		int(v.ffhelp.GetVideoDuration()),
		FrameDuration,
		v.thumbsCount(), SpriteSheetCols, SpriteSheetRows, SpriteWidth, SpriteHeight,
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

	file, err := v.ffhelp.SaveFrame(
		v.defaultPreviewPath(),
		v.ffhelp.GetVideoDuration()/2,
		DefaultPreviewWidth,
		DefaultPreviewHeight,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	f, err := filesystem.NewFileFromPath(file.Name())
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

func (v *VideoUploaderBase) thumbsDir() string {
	return ThumbsDir + "/" + v.video.ID()
}

func (v *VideoUploaderBase) sheetsDir() string {
	return SpriteSheetsDir + "/" + v.video.ID()
}

func (v *VideoUploaderBase) webvttDir() string {
	return WebVTTDir + "/" + v.video.ID()
}

func (v *VideoUploaderBase) defaultPreviewPath() string {
	return UploadDir + "/" + "preview_" + v.video.ID() + ".jpg"
}

func (v *VideoUploaderBase) readThumbsDir() ([]os.DirEntry, error) {
	return os.ReadDir(v.thumbsDir())
}

func (v *VideoUploaderBase) thumbsCount() int {
	entries, err := v.readThumbsDir()
	if err != nil {
		return 0
	}

	return len(entries)
}
