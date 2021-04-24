package widget

import (
	"image"

	"github.com/blizzy78/ebitenui/input"
	"github.com/hajimehoshi/ebiten/v2"
)

type Window struct {
	ID    int
	Modal bool

	contents *Container
}

type WindowOpt func(w *Window)

type WindowOptions struct {
}

var WindowOpts WindowOptions

func NewWindow(opts ...WindowOpt) *Window {
	w := &Window{}

	for _, o := range opts {
		o(w)
	}

	return w
}

func (o WindowOptions) Contents(c *Container) WindowOpt {
	return func(w *Window) {
		w.contents = c
	}
}

func (o WindowOptions) Modal() WindowOpt {
	return func(w *Window) {
		w.Modal = true
	}
}

func (w *Window) Container() *Container {
	return w.contents
}

func (w *Window) SetLocation(rect image.Rectangle) {
	w.contents.SetLocation(rect)
}

func (w *Window) RequestRelayout() {
	w.contents.RequestRelayout()
}

func (w *Window) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	var l *input.Layer
	if w.Modal {
		l = &input.Layer{
			DebugLabel: "modal window",
			EventTypes: input.LayerEventTypeAll,
			BlockLower: true,
			FullScreen: true,
		}
	} else {
		l = &input.Layer{
			DebugLabel: "window",
			EventTypes: input.LayerEventTypeAll,
			BlockLower: true,
			RectFunc: func() image.Rectangle {
				return w.contents.GetWidget().Rect
			},
		}
	}
	w.contents.GetWidget().ElevateToNewInputLayer(l)
}

func (w *Window) Render(screen *ebiten.Image, def DeferredRenderFunc) {
	w.contents.Render(screen, def)
}
