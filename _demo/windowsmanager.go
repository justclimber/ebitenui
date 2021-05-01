package main

import (
	"github.com/justclimber/ebitenui"
	eimage "github.com/justclimber/ebitenui/image"
	"github.com/justclimber/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"image"
	"image/color"
)

type windowsManager struct {
	windowsOptions map[string]*windowOption
	ui             func() *ebitenui.UI
	bgImage        *eimage.NineSlice
	padding        widget.Insets
	spacing        int
	face           font.Face
	headerColor    color.Color
}

func newWindowsManager(
	ui func() *ebitenui.UI,
	bgImage *eimage.NineSlice,
	padding widget.Insets,
	spacing int,
	face font.Face,
	headerColor color.Color,
) *windowsManager {
	return &windowsManager{
		windowsOptions: make(map[string]*windowOption),
		ui:             ui,
		bgImage:        bgImage,
		padding:        padding,
		spacing:        spacing,
		face:           face,
		headerColor:    headerColor,
	}
}

type openClosedState int8

const (
	stateClosed = iota
	stateOpen
)

type windowOption struct {
	pos             image.Point
	w               int
	h               int
	openClosedState openClosedState
	windowCloseFunc ebitenui.RemoveWindowFunc
}

func (wm *windowsManager) windowToggle(page *page) {
	wo, ok := wm.windowsOptions[page.title]
	if ok {
		if wo.openClosedState == stateOpen {
			wo.windowCloseFunc()
			wo.openClosedState = stateClosed
			return
		}
	} else {
		w, h := page.content.PreferredSize()
		w += wm.padding.Dx()
		// @todo: calculate headerAndExtraSpaceY
		headerAndExtraSpaceY := 12
		h += wm.padding.Dy() + wm.spacing + headerAndExtraSpaceY
		ew, eh := ebiten.WindowSize()
		x := (ew - w) / 2
		y := (eh - h) / 2
		wo = &windowOption{
			w:   w,
			h:   h,
			pos: image.Point{x, y},
		}
		wm.windowsOptions[page.title] = wo
	}
	wo.openClosedState = stateOpen
	wm.addWindow(wo, page)
}

func (wm *windowsManager) addWindow(wo *windowOption, page *page) {
	c := widget.NewContainer(
		"window "+page.title,
		widget.ContainerOpts.BackgroundImage(wm.bgImage),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(wm.padding),
			widget.RowLayoutOpts.Spacing(wm.spacing),
		)),
	)

	mc := widget.NewContainer(
		"window "+page.title+" movable",
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical)),
		),
	)

	mc.AddChild(widget.NewText(
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.TextOpts.Text(page.title, wm.face, wm.headerColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	c.AddChild(page.content)

	w := widget.NewWindow(
		widget.WindowOpts.Movable(mc),
		widget.WindowOpts.Contents(c),
	)

	r := image.Rectangle{image.Point{0, 0}, image.Point{wo.w, wo.h}}
	r = r.Add(wo.pos)
	w.SetLocation(r)

	wo.windowCloseFunc = wm.ui().AddWindow(w)
}
