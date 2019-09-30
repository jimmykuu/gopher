package main

import (
	"gitea.com/lunny/tango"
	"gitea.com/tango/events"
	"gitea.com/tango/renders"

	"github.com/jimmykuu/gopher"
	"github.com/jimmykuu/gopher/modules/static"
	"github.com/jimmykuu/gopher/modules/templates"
)

func main() {
	t := tango.Classic()
	t.Use(
		events.Events(),
		static.Static("./static"),
		renders.New(renders.Options{
			Reload:     true,
			Directory:  "./templates",
			FileSystem: templates.FileSystem("templates"),
			Funcs:      gopher.Funcs,
		}))

	gopher.SetRoutes(t)

	t.Run()
}
