package vhs

import "github.com/gorilla/websocket"

type App interface {
	Start() error
	UploadVideo(conn *websocket.Conn) error
}
