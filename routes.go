package main

import (
	"github.com/lunny/tango"

	"github.com/jimmykuu/gopher/actions"
	"github.com/jimmykuu/gopher/apis"
)

func setRoutes(t *tango.Tango) {
	t.Any("/signin", new(actions.Signin))
	t.Any("/api/signin", new(apis.Signin))
}
