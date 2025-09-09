package handlers

import (
	"errors"
	"net/http"
	"vhs/internal/vhs"
	"vhs/internal/vhs/entities"
	"vhs/internal/vhs/entities/dto"

	"github.com/gorilla/websocket"
	"github.com/pocketbase/pocketbase/core"
)

type Handlers struct {
	app vhs.App
}

func New(app vhs.App) *Handlers {
	//if app.IsDev() {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	//}

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

	record, err := vhs.PocketBase.FindRecordById(entities.VideosCollection, videoId)
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

func (h *Handlers) UpdateVideoHandler(e *core.RequestEvent) error {
	var data *dto.VideoUpdateRequest
	if err := e.BindBody(&data); err != nil {
		return e.BadRequestError("invalid request body", err)
	}

	files, err := e.FindUploadedFiles("preview")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return e.InternalServerError("error while processing uploaded file", err)
	} else if len(files) > 0 {
		data.Preview = files[0]
	}

	videoId := e.Request.PathValue("videoId")
	err = h.app.UpdateVideo(videoId, e.Auth.Id, dto.NewVideoUpdate(data))
	if err != nil {
		return e.InternalServerError("error while updating video", err)
	}

	return nil
}

func (h *Handlers) CreatePlaylistHandler(e *core.RequestEvent) error {
	var data *dto.PlaylistCreateRequest
	if err := e.BindBody(&data); err != nil {
		return e.BadRequestError("invalid request body", err)
	}

	err := h.app.CreatePlaylist(e.Auth.Id, dto.NewPlaylistCreate(data))
	if err != nil {
		return e.InternalServerError("error while creating playlist", err)
	}

	return nil
}

func (h *Handlers) UpdatePlaylistHandler(e *core.RequestEvent) error {
	var data *dto.PlaylistUpdateRequest
	if err := e.BindBody(&data); err != nil {
		return e.BadRequestError("invalid request body", err)
	}

	playlistId := e.Request.PathValue("playlistId")
	err := h.app.UpdatePlaylist(playlistId, e.Auth.Id, dto.NewPlaylistUpdate(data))
	if err != nil {
		return e.InternalServerError("error while updating playlist", err)
	}

	return nil
}
