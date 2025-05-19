package vhs

type VideoUploader interface {
	Start(*VideoUploadData) error
	UploadPart([]byte) (bool, error)
	Cancel() error
	Done() error
}

type VideoUploadData struct {
	Size   int    `json:"size"`
	Name   string `json:"name"`
	Token  string `json:"token"`
	UserId string
}
