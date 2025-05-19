package vhs

import (
	"encoding/json"
	"errors"
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

const (
	UploadVideoMessagePart   = "part"
	UploadVideoMessageEnd    = "end"
	UploadVideoMessageCancel = "cancel"
	UploadVideoMessageError  = "error"
)

func (a *AppBase) UploadVideo(c *websocket.Conn) error {
	var (
		v          VideoUploader
		err        error
		mt         int
		message    []byte
		done       = false
		resMessage string
	)

	res := map[string]interface{}{}
	v = NewVideoUploader()

	for {
		mt, message, err = c.ReadMessage()
		if err != nil {
			resMessage = UploadVideoMessageError
			break
		}

		switch mt {
		case websocket.TextMessage:
			err = a.startUpload(message, v)
			resMessage = UploadVideoMessagePart
			break
		case websocket.BinaryMessage:
			done, err = v.UploadPart(message)
			if done {
				resMessage = UploadVideoMessageEnd
			} else {
				resMessage = UploadVideoMessagePart
			}
			break
		case websocket.CloseMessage:
			err = v.Cancel()
			resMessage = UploadVideoMessageCancel
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

	record, err := PocketBase.FindAuthRecordByToken(data.Token, core.TokenTypeAuth)
	if err != nil {
		return err
	}
	if record == nil {
		return errors.New("invalid token")
	}

	data.UserId = record.Id
	err = v.Start(data)
	if err != nil {
		return err
	}

	return nil
}

func (a *AppBase) IsDev() bool {
	return PocketBase.IsDev()
}
