package main

import (
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"log"
	"vhs/internal/http/handlers/v1"
	"vhs/internal/middleware"
	"vhs/internal/vhs"
)

func main() {
	app := vhs.New()
	handlers := handlers.New(app)

	vhs.PocketBase.OnServe().BindFunc(func(se *core.ServeEvent) error {
		r := se.Router
		api := r.Group("/api")

		upload := api.Group("")
		upload.
			GET("/upload", handlers.UploadVideoHandler)

		video := api.Group("/video/{videoId}")
		video.
			Group("").
			Bind(apis.RequireAuth()).
			POST("/update", handlers.UpdateVideoHandler)
		video.
			Group("").
			Bind(middleware.AuthorizeGet()).
			GET("/stream", handlers.ServeVideoHandler)

		playlist := api.Group("/playlist").Bind(apis.RequireAuth())
		playlist.POST("", handlers.CreatePlaylistHandler)
		playlist.Group("/{playlistId}").POST("", handlers.UpdatePlaylistHandler)

		return se.Next()
	})

	err := app.Start()
	if err != nil {
		log.Fatal(err)
	}
}
