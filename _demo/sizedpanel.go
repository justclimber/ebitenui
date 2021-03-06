package main

import (
	"image"

	"github.com/justclimber/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

type sizedPanel struct {
	width     int
	height    int
	container *widget.Container
}

func newSizedPanel(w int, h int, opts ...widget.ContainerOpt) *sizedPanel {
	return &sizedPanel{
		width:     w,
		height:    h,
		container: widget.NewContainer("sized panel", opts...),
	}
}

func (p *sizedPanel) GetWidget() *widget.Widget {
	return p.container.GetWidget()
}

func (p *sizedPanel) PreferredSize() (int, int) {
	return p.width, p.height
}

func (p *sizedPanel) RequestRelayout() {
	p.container.RequestRelayout()
}

func (p *sizedPanel) SetLocation(rect image.Rectangle) {
	p.container.SetLocation(rect)
}

func (p *sizedPanel) Render(screen *ebiten.Image, def widget.DeferredRenderFunc, debugMode widget.DebugMode) {
	p.container.Render(screen, def, debugMode)
}

func (p *sizedPanel) Container() *widget.Container {
	return p.container
}
