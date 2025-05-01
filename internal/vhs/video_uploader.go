package vhs

type VideoUploader interface {
	Start(*VideoUploadData) error
	UploadPart([]byte) (bool, error)
	Cancel() error
	Done() error
}

type VideoUploadData struct {
	FileSize int
	FileName string
}
