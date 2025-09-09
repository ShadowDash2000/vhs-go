package vhs

import (
	"vhs/internal/vhs/entities/dto"

	"github.com/gorilla/websocket"
)

type App interface {
	Start() error
	UploadVideo(conn *websocket.Conn) error
	UpdateVideo(id string, userId string, data *dto.VideoUpdate) error
	CreatePlaylist(userId string, data *dto.PlaylistCreate) error
	UpdatePlaylist(id string, userId string, data *dto.PlaylistUpdate) error
}
