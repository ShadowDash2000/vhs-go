package dto

import (
	"vhs/internal/vhs/entities"

	"github.com/pocketbase/pocketbase/tools/filesystem"
)

type VideoUpdateRequest struct {
	Name        string   `form:"name"`
	Description string   `form:"description"`
	Status      string   `form:"status"`
	PlaylistIds []string `form:"playlists"`
	Preview     *filesystem.File
}

type VideoUpdate struct {
	Name        string
	Description string
	Status      entities.Status
	Preview     *filesystem.File
	PlaylistIds []string
}

func NewVideoUpdate(req *VideoUpdateRequest) *VideoUpdate {
	return &VideoUpdate{
		Name:        req.Name,
		Description: req.Description,
		Status:      entities.Status(req.Status),
		Preview:     req.Preview,
		PlaylistIds: req.PlaylistIds,
	}
}
