package actions

import (
	"github.com/lunny/tango"
	"github.com/tango-contrib/renders"
)

type RenderBase struct {
	renders.Renderer
	tango.Ctx
}

func (r *RenderBase) Render(tmpl string, t ...renders.T) error {
	if len(t) > 0 {
		return r.Renderer.Render(tmpl, t[0].Merge(renders.T{}))
	}
	return r.Renderer.Render(tmpl, renders.T{})
}
