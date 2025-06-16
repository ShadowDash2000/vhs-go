package dto

type PlaylistCreateRequest struct {
	Name   string   `form:"name" json:"name"`
	Videos []string `form:"videos" json:"videos"`
}

type PlaylistCreate struct {
	Name   string
	Videos []string
}

func NewPlaylistCreate(req *PlaylistCreateRequest) *PlaylistCreate {
	return &PlaylistCreate{
		Name:   req.Name,
		Videos: req.Videos,
	}
}

type PlaylistUpdateRequest struct {
	Name   string   `form:"name" json:"name"`
	Videos []string `form:"videos" json:"videos"`
}

type PlaylistUpdate struct {
	Name   string
	Videos []string
}

func NewPlaylistUpdate(req *PlaylistUpdateRequest) *PlaylistUpdate {
	return &PlaylistUpdate{
		Name:   req.Name,
		Videos: req.Videos,
	}
}
