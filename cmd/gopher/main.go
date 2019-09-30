package main

import (
	"gitea.com/lunny/tango"
	"gitea.com/tango/events"
	"gitea.com/tango/renders"

	"github.com/jimmykuu/gopher"
)

func main() {
	t := tango.Classic()
	t.Use(
		events.Events(),
		tango.Static(tango.StaticOptions{
			RootPath: "./static",
			Prefix:   "static",
		}),
		renders.New(renders.Options{
			Reload:    true,
			Directory: "./templates",
			Funcs:     gopher.Funcs,
		}))

	gopher.SetRoutes(t)

	t.Run()
}
