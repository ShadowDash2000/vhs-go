package handlers

import (
	"github.com/gorilla/websocket"
	"github.com/pocketbase/pocketbase/core"
	"net/http"
	"vhs/internal/vhs"
)

type Handlers struct {
	app vhs.App
}

func New(app vhs.App) *Handlers {
	if app.IsDev() {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	}

	return &Handlers{
		app: app,
	}
}

var upgrader = websocket.Upgrader{}

func (h *Handlers) UploadVideoHandler(e *core.RequestEvent) error {
	c, err := upgrader.Upgrade(e.Response, e.Request, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	err = h.app.UploadVideo(c)

	return err
}

func (h *Handlers) ServeVideoHandler(e *core.RequestEvent) error {
	videoId := e.Request.PathValue("videoId")

	info, err := e.RequestInfo()
	if err != nil {
		return err
	}

	record, err := vhs.PocketBase.FindRecordById(vhs.VideosCollection, videoId)
	if err != nil {
		return err
	}

	video := vhs.NewVideoFromRecord(record)

	canAccess, err := vhs.PocketBase.CanAccessRecord(record, info, record.Collection().ViewRule)
	if err != nil {
		return err
	}

	if !canAccess {
		return e.NotFoundError("video not found", nil)
	}

	fs, err := vhs.PocketBase.NewFilesystem()
	if err != nil {
		return err
	}

	return fs.Serve(
		e.Response,
		e.Request,
		record.BaseFilesPath()+"/"+video.Video(),
		video.Name(),
	)
}
