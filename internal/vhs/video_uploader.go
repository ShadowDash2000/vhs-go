package vhs

type VideoUploader interface {
	Start(*VideoUploadData) (string, error)
	UploadPart([]byte) (bool, error)
	Cancel() error
	Done()
}

type VideoUploadData struct {
	Size   int    `json:"size"`
	Name   string `json:"name"`
	Token  string `json:"token"`
	UserId string
}
