package ebitenui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image"

	"github.com/justclimber/ebitenui/event"
	"github.com/justclimber/ebitenui/input"
	internalinput "github.com/justclimber/ebitenui/internal/input"
	"github.com/justclimber/ebitenui/widget"

	"github.com/hajimehoshi/ebiten/v2"
)

// UI encapsulates a complete user interface that can be rendered onto the screen.
// There should only be exactly one UI per application.
type UI struct {
	// Container is the root container of the UI hierarchy.
	Container *widget.Container

	// ToolTip is used to render mouse hover tool tips. It may be nil to disable rendering.
	ToolTip *widget.ToolTip

	// DragAndDrop is used to render drag widgets while dragging and dropping. It may be nil to disable rendering.
	DragAndDrop *widget.DragAndDrop

	lastRect      image.Rectangle
	focusedWidget widget.HasWidget
	inputLayerers []input.Layerer
	renderers     []widget.Renderer
	windows       []*widget.Window
	nextWindowsId int
	debugMode     widget.DebugMode
}

// RemoveWindowFunc is a function to remove a Window from rendering.
type RemoveWindowFunc func()

// Update updates u. This method should be called in the Ebiten Update function.
func (u *UI) Update() {
	internalinput.Update()
}

// Draw renders u onto screen. This function should be called in the Ebiten Draw function.
//
// If screen's size changes from one frame to the next, u.Container.RequestRelayout is called.
func (u *UI) Draw(screen *ebiten.Image) {
	event.ExecuteDeferred()

	internalinput.Draw()
	defer internalinput.AfterDraw()

	w, h := screen.Size()
	rect := image.Rect(0, 0, w, h)

	defer func() {
		u.lastRect = rect
	}()

	if rect != u.lastRect {
		u.Container.RequestRelayout()
	}

	u.handleFocus()
	u.setupInputLayers()
	u.Container.SetLocation(rect)
	u.render(screen)
	u.drawDebug(screen)
}

// SetDebugMode set debug mode for u to debugMode.
func (u *UI) SetDebugMode(debugMode widget.DebugMode) {
	u.debugMode = debugMode
}

func (u *UI) drawDebug(screen *ebiten.Image) {
	if u.debugMode == widget.DebugModeNone {
		return
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f, FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()))
	switch u.debugMode {
	case widget.DebugModeBorderAlwaysShow:
		u.Container.RenderWidgetSizeDebug(screen)
		for _, w := range u.windows {
			w.Container().RenderWidgetSizeDebug(screen)
		}
	case widget.DebugModeInputLayersAlwaysShow:
		u.Container.RenderInputLayerDebug(screen)
		for _, w := range u.windows {
			w.Container().RenderInputLayerDebug(screen)
		}
	}
}

func (u *UI) handleFocus() {
	if !input.MouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	if u.focusedWidget != nil {
		u.focusedWidget.(widget.Focuser).Focus(false)
		u.focusedWidget = nil
	}

	x, y := input.CursorPosition()

	for i := len(u.windows) - 1; i >= 0; i-- {
		if u.handleContainerFocus(u.windows[i].Container(), x, y) {
			u.windowToTop(i)
			return
		}
	}
	u.handleContainerFocus(u.Container, x, y)
}

func (u *UI) windowToTop(i int) {
	if len(u.windows) == 1 || i == len(u.windows)-1 {
		return
	}
	w := u.windows[i]
	u.windows = append(u.windows[:i], u.windows[i+1:]...)
	u.windows = append(u.windows, w)
}

func (u *UI) handleContainerFocus(c *widget.Container, x int, y int) bool {
	w := c.WidgetAt(x, y)
	if w == nil {
		return false
	}

	if !w.GetWidget().EffectiveInputLayer().ActiveFor(x, y, input.LayerEventTypeMouseButton) {
		return false
	}

	f, ok := w.(widget.Focuser)
	if !ok {
		return true
	}

	f.Focus(true)
	u.focusedWidget = w
	return true
}

func (u *UI) setupInputLayers() {
	num := 1 // u.Container
	if len(u.windows) > 0 {
		num += len(u.windows)
	}
	if u.DragAndDrop != nil {
		num++
	}

	if cap(u.inputLayerers) < num {
		u.inputLayerers = make([]input.Layerer, num)
	}

	u.inputLayerers = u.inputLayerers[:0]
	u.inputLayerers = append(u.inputLayerers, u.Container)
	for _, w := range u.windows {
		u.inputLayerers = append(u.inputLayerers, w)
	}
	if u.DragAndDrop != nil {
		u.inputLayerers = append(u.inputLayerers, u.DragAndDrop)
	}

	// TODO: SetupInputLayersWithDeferred should reside in "internal" subpackage
	input.SetupInputLayersWithDeferred(u.inputLayerers)
}

func (u *UI) render(screen *ebiten.Image) {
	num := 1 // u.Container
	if len(u.windows) > 0 {
		num += len(u.windows)
	}
	if u.ToolTip != nil {
		num++
	}
	if u.DragAndDrop != nil {
		num++
	}

	if cap(u.renderers) < num {
		u.renderers = make([]widget.Renderer, num)
	}

	u.renderers = u.renderers[:0]
	u.renderers = append(u.renderers, u.Container)
	for _, w := range u.windows {
		u.renderers = append(u.renderers, w)
	}
	if u.ToolTip != nil {
		u.renderers = append(u.renderers, u.ToolTip)
	}
	if u.DragAndDrop != nil {
		u.renderers = append(u.renderers, u.DragAndDrop)
	}

	// TODO: RenderWithDeferred should reside in "internal" subpackage
	widget.RenderWithDeferred(screen, u.renderers, u.debugMode)
}

// AddWindow adds window w to u for rendering. It returns a function to remove w from u.
func (u *UI) AddWindow(w *widget.Window) RemoveWindowFunc {
	w.ID = u.nextWindowsId
	u.windows = append(u.windows, w)
	u.nextWindowsId++

	return func() {
		u.removeWindow(w.ID)
	}
}

func (u *UI) removeWindow(id int) {
	for i, uw := range u.windows {
		if uw.ID == id {
			u.windows = append(u.windows[:i], u.windows[i+1:]...)
			break
		}
	}
}
