package vhs

import (
	"github.com/gorilla/websocket"
	"vhs/internal/vhs/entities/dto"
)

type App interface {
	Start() error
	IsDev() bool
	UploadVideo(conn *websocket.Conn) error
	UpdateVideo(id string, userId string, data *dto.VideoUpdate) error
	CreatePlaylist(userId string, data *dto.PlaylistCreate) error
	UpdatePlaylist(id string, userId string, data *dto.PlaylistUpdate) error
}
