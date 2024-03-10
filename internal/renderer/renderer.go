package renderer

import (
	"context"
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin/render"
	"net/http"
)

var Default = &Renderer{}

type Renderer struct {
	Ctx context.Context
	Cmp templ.Component
}

func New(ctx context.Context, component templ.Component) *Renderer {
	return &Renderer{
		ctx,
		component,
	}
}

func (t Renderer) Render(w http.ResponseWriter) error {
	t.WriteContentType(w)
	if t.Cmp != nil {
		return t.Cmp.Render(t.Ctx, w)
	}
	return nil
}

func (t Renderer) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

func (t *Renderer) Instance(name string, data any) render.Render {
	templData, ok := data.(templ.Component)
	if !ok {
		return nil
	}
	return &Renderer{
		Ctx: context.Background(),
		Cmp: templData,
	}
}
