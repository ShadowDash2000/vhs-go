package vhs

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ncruces/go-sqlite3/driver"
	"github.com/ncruces/go-sqlite3/ext/unicode"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"os"
	"strings"
	"vhs/internal/vhs/entities/dto"
	"vhs/pkg/collections"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
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
	PocketBase = pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDev: inspectRuntime(),
		DBConnect: func(dbPath string) (*dbx.DB, error) {
			const pragmas = "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=temp_store(MEMORY)&_pragma=cache_size(-16000)"

			db, err := driver.Open("file:"+dbPath+pragmas, unicode.Register)
			if err != nil {
				return nil, err
			}

			return dbx.NewFromDB(db, "sqlite3"), nil
		},
	})
	Collections = collections.NewCollections(PocketBase)

	return &AppBase{}
}

func (a *AppBase) Start() error {
	return PocketBase.Start()
}

func (a *AppBase) IsDev() bool {
	return PocketBase.IsDev()
}

func inspectRuntime() (withGoRun bool) {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		// probably ran with go run
		withGoRun = true
	} else {
		// probably ran with go build
		withGoRun = false
	}
	return
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
		videoId    string
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
			videoId, err = a.startUpload(message, v)
			resMessage = UploadVideoMessagePart
			res["videoId"] = videoId
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

		clear(res)
	}

	if done {
		v.Done()
	}

	return err
}

func (a *AppBase) startUpload(message []byte, v VideoUploader) (string, error) {
	var data *VideoUploadData
	if err := json.Unmarshal(message, &data); err != nil {
		return "", err
	}

	record, err := PocketBase.FindAuthRecordByToken(data.Token, core.TokenTypeAuth)
	if err != nil {
		return "", err
	}
	if record == nil {
		return "", errors.New("invalid token")
	}

	data.UserId = record.Id
	videoId, err := v.Start(data)
	if err != nil {
		return "", err
	}

	return videoId, nil
}

func (a *AppBase) UpdateVideo(id string, userId string, data *dto.VideoUpdate) error {
	var err error
	defer func() {
		if err != nil {
			PocketBase.Logger().Error(
				"error while updating video: "+err.Error(),
				map[string]any{
					"videoId": id,
					"user":    userId,
					"data":    data,
				},
			)
		}
	}()

	video, err := NewVideoFromId(id)
	if err != nil {
		return err
	}

	if video.User() != userId {
		err = fmt.Errorf("expected user %s, got %s", video.User(), userId)
		return err
	}

	if data.Name != "" {
		video.SetName(data.Name)
	}
	if data.Description != "" {
		video.SetDescription(data.Description)
	}
	if data.Status != "" {
		video.SetStatus(data.Status)
	}
	if data.Preview.Size > 0 {
		video.SetPreview(data.Preview)
	}

	return video.Save()
}

func (a *AppBase) CreatePlaylist(userId string, data *dto.PlaylistCreate) error {
	var err error
	defer func() {
		if err != nil {
			PocketBase.Logger().Error(
				"error while creating playlist: "+err.Error(),
				map[string]any{
					"user": userId,
					"data": data,
				},
			)
		}
	}()

	playlist, err := NewPlaylist()
	if err != nil {
		return err
	}

	playlist.SetName(data.Name)
	playlist.SetUser(userId)
	playlist.SetVideos(data.Videos)

	return playlist.Save()
}

func (a *AppBase) UpdatePlaylist(id string, userId string, data *dto.PlaylistUpdate) error {
	var err error
	defer func() {
		if err != nil {
			PocketBase.Logger().Error(
				"error while updating playlist: "+err.Error(),
				map[string]any{
					"playlistId": id,
					"user":       userId,
					"data":       data,
				},
			)
		}
	}()

	playlist, err := NewPlaylistFromId(id)
	if err != nil {
		return err
	}

	if playlist.User() != userId {
		err = fmt.Errorf("expected user %s, got %s", playlist.User(), userId)
		return err
	}

	if data.Name != "" {
		playlist.SetName(data.Name)
	}
	if data.Videos != nil {
		playlist.SetVideos(data.Videos)
	}

	return playlist.Save()
}
