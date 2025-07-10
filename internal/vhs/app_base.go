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
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"golang.org/x/exp/slices"
	"log/slog"
	"os"
	"strings"
	"vhs/internal/vhs/entities"
	"vhs/internal/vhs/entities/dto"
	"vhs/internal/vhs/helper"
	"vhs/pkg/collections"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

var (
	PocketBase  *pocketbase.PocketBase
	Collections *collections.Collections
)

type AppBase struct {
	logger *slog.Logger
}

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

	app := &AppBase{
		logger: PocketBase.Logger(),
	}

	app.bindHooks()

	return app
}

func (a *AppBase) bindHooks() {
	PocketBase.OnRecordCreate(entities.PlaylistsCollection).BindFunc(a.updatePlaylistPreview)
	PocketBase.OnRecordUpdate(entities.PlaylistsCollection).BindFunc(a.updatePlaylistPreview)
	PocketBase.OnRecordAfterUpdateSuccess(entities.VideosCollection).BindFunc(a.updatePlaylistPreviewFromVideo)
	PocketBase.OnRecordEnrich(entities.VideosCollection).BindFunc(a.enrichVideo)
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
	v = NewVideoUploader(a.logger)

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
			a.logger.Error(
				"error while updating video: "+err.Error(),
				"videoId", id,
				"user", userId,
				"data", data,
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
	err = a.removeVideoFromPlaylists(data.PlaylistIds, video)
	if err != nil {
		return err
	}
	if len(data.PlaylistIds) > 0 {
		err = a.addVideoToPlaylists(userId, data.PlaylistIds, video)
		if err != nil {
			return err
		}
	}

	return video.Save()
}

func (a *AppBase) removeVideoFromPlaylists(playlistIds []string, video Video) error {
	currentPlaylists, err := NewPlaylistsFromVideoId(video.ID())
	if err != nil {
		return err
	}

	for _, playlist := range currentPlaylists {
		if !slices.Contains(playlistIds, playlist.ID()) {
			playlist.RemoveVideo(video.ID())
			err = playlist.Save()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *AppBase) addVideoToPlaylists(userId string, playlistIds []string, video Video) error {
	var playlists []Playlist
	playlists, err := NewPlaylistFromIds(playlistIds)
	if err != nil {
		return err
	}

	err = helper.VerifyFields(
		playlists,
		userId,
		func(playlist Playlist) string {
			return playlist.User()
		},
		func(playlist Playlist) error {
			playlist.AddVideo(video.ID())
			return playlist.Save()
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (a *AppBase) CreatePlaylist(userId string, data *dto.PlaylistCreate) error {
	var err error
	defer func() {
		if err != nil {
			a.logger.Error(
				"error while creating playlist: "+err.Error(),
				"user", userId,
				"data", data,
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
			a.logger.Error(
				"error while updating playlist: "+err.Error(),
				"playlistId", id,
				"user", userId,
				"data", data,
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

func (a *AppBase) updatePlaylistPreviewFromVideoRecord(playlist Playlist, video Video) error {
	key := video.BaseFilesPath() + "/" + video.Preview()

	fs, _ := PocketBase.NewFilesystem()
	defer fs.Close()

	blob, _ := fs.GetReader(key)
	defer blob.Close()

	buff := make([]byte, blob.Size())
	_, err := blob.Read(buff)
	if err != nil {
		return err
	}

	file, err := filesystem.NewFileFromBytes(buff, "playlist-preview-"+playlist.ID())
	if err != nil {
		return err
	}

	playlist.SetPreview(file)
	return nil
}

func (a *AppBase) updatePlaylistPreview(e *core.RecordEvent) error {
	playlist := NewPlaylistFromRecord(e.Record)

	if len(playlist.Videos()) == 0 {
		return e.Next()
	}

	video, err := NewVideoFromId(playlist.Videos()[0])
	if err != nil {
		return err
	}

	err = a.updatePlaylistPreviewFromVideoRecord(playlist, video)
	if err != nil {
		return err
	}

	return e.Next()
}

func (a *AppBase) updatePlaylistPreviewFromVideo(e *core.RecordEvent) error {
	video := NewVideoFromRecord(e.Record)
	playlists, err := NewPlaylistsFromVideoId(video.ID())
	if err != nil {
		return err
	}

	for _, playlist := range playlists {
		playlist.Save()
	}

	return e.Next()
}

func (a *AppBase) enrichVideo(e *core.RecordEnrichEvent) error {
	playlists, err := NewPlaylistsFromVideoId(e.Record.Id)
	if err != nil {
		return err
	}

	playlistIds := make([]string, len(playlists))
	for _, playlist := range playlists {
		playlistIds = append(playlistIds, playlist.ID())
	}

	e.Record.WithCustomData(true)
	e.Record.Set("playlists", playlistIds)

	return e.Next()
}
