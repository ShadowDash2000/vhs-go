package main

import (
	"github.com/pocketbase/pocketbase/core"
	"log"
	"vhs/internal/http/handlers/v1"
	"vhs/internal/vhs"
	_ "vhs/migrations"
)

func main() {
	app := vhs.New()
	handlers.New(app)

	vhs.PocketBase.OnServe().BindFunc(func(se *core.ServeEvent) error {
		return se.Next()
	})

	err := app.Start()
	if err != nil {
		log.Fatal(err)
	}
}
