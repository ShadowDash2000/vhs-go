package vhs

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"vhs/pkg/collections"
)

var (
	PocketBase  *pocketbase.PocketBase
	Collections *collections.Collections
)

type AppBase struct{}

type Components struct {
	App  core.App
	Cols *collections.Collections
}

func New() App {
	PocketBase = pocketbase.New()
	Collections = collections.NewCollections(PocketBase)

	return &AppBase{}
}

func (a *AppBase) Start() error {
	return PocketBase.Start()
}

type UploadVideoMessage int

const (
	UploadVideoMessageStart UploadVideoMessage = iota
	UploadVideoMessagePart
	UploadVideoMessageEnd
	UploadVideoMessageCancel
	UploadVideoMessageError
)

func (a *AppBase) UploadVideo(c *websocket.Conn) error {
	var (
		v          VideoUploader
		err        error
		mt         int
		message    []byte
		done       = false
		resMessage UploadVideoMessage
	)

	res := map[string]interface{}{}

	v, err = NewVideoUploader()
	if err != nil {
		return err
	}

	for {
		mt, message, err = c.ReadMessage()
		if err != nil {
			resMessage = UploadVideoMessageError
			break
		}

		switch UploadVideoMessage(mt) {
		case UploadVideoMessageStart:
			err = a.startUpload(message, v)
			resMessage = UploadVideoMessagePart
			break
		case UploadVideoMessagePart:
			done, err = v.UploadPart(message)
			if done {
				resMessage = UploadVideoMessageEnd
			} else {
				resMessage = UploadVideoMessagePart
			}
			break
		case UploadVideoMessageCancel:
			err = v.Cancel()
			resMessage = UploadVideoMessageEnd
			break
		default:
			return nil
		}

		if err != nil {
			resMessage = UploadVideoMessageError
			res["error"] = err.Error()
		}

		res["type"] = resMessage
		err = c.WriteJSON(res)
		if err != nil || done {
			break
		}
	}

	if done {
		go v.Done()
	}

	return err
}

func (a *AppBase) startUpload(message []byte, v VideoUploader) error {
	var data *VideoUploadData
	if err := json.Unmarshal(message, &data); err != nil {
		return err
	}

	err := v.Start(data)
	if err != nil {
		return err
	}

	return nil
}
