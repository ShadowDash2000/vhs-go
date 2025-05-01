package handlers

import (
	"github.com/gorilla/websocket"
	"github.com/pocketbase/pocketbase/core"
	"vhs/internal/vhs"
)

type Handlers struct {
	app vhs.App
}

func New(app vhs.App) *Handlers {
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
