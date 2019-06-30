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
			Funcs:      gopher.Funcs,
			FileSystem: templates.FileSystem("./templates"),
		}))

	gopher.SetRoutes(t)

	t.Run()
}
