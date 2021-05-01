package main

import (
	"github.com/blizzy78/ebitenui/event"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
	"log"

	_ "image/png"

	"github.com/blizzy78/ebitenui"
	"github.com/blizzy78/ebitenui/image"
	"github.com/blizzy78/ebitenui/widget"
)

type game struct {
	ui *ebitenui.UI
}

func main() {
	ebiten.SetWindowSize(1300, 800)
	ebiten.SetWindowTitle("Ebiten UI Demo")
	ebiten.SetVsyncEnabled(false)
	ebiten.SetWindowResizable(true)
	ebiten.SetScreenClearedEveryFrame(false)

	ui, closeUI, err := createUI()
	if err != nil {
		log.Fatal(err)
	}

	defer closeUI()

	game := game{
		ui: ui,
	}

	err = ebiten.RunGame(&game)
	if err != nil {
		log.Print(err)
	}
}

func createUI() (*ebitenui.UI, func(), error) {
	res, err := newUIResources()
	if err != nil {
		return nil, nil, err
	}

	rootContainer := widget.NewContainer(
		"root",
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true, false}),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Top:    20,
				Bottom: 20,
			}),
			widget.GridLayoutOpts.Spacing(0, 20))),
		widget.ContainerOpts.BackgroundImage(res.background))

	toolTips := toolTipContents{
		tips: map[widget.HasWidget]string{},
		res:  res,
	}

	toolTip := widget.NewToolTip(
		// @todo: need to set a page container, not a root container. the same as I did it for dnd
		widget.ToolTipOpts.Container(rootContainer),
		widget.ToolTipOpts.ContentsCreater(&toolTips),
	)

	rootContainer.AddChild(headerContainer(res))

	var ui *ebitenui.UI
	demoContainer, dndPage, dragContents, dropHandler := demoContainer(res, &toolTips, toolTip, func() *ebitenui.UI {
		return ui
	})
	rootContainer.AddChild(demoContainer)

	dnd := widget.NewDragAndDrop(
		widget.DragAndDropOpts.Container(dndPage),
		widget.DragAndDropOpts.ContentsCreater(dragContents),
	)
	dnd.DroppedEvent.AddHandler(dropHandler)

	urlContainer := widget.NewContainer("url", widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Padding(widget.Insets{
			Left:  25,
			Right: 25,
		}),
	)))
	rootContainer.AddChild(urlContainer)

	urlContainer.AddChild(widget.NewText(
		widget.TextOpts.Text("github.com/blizzy78/ebitenui", res.text.smallFace, res.text.disabledColor)))

	ui = &ebitenui.UI{
		Container: rootContainer,

		ToolTip: toolTip,

		DragAndDrop: dnd,
	}

	return ui, func() {
		res.close()
	}, nil
}

func headerContainer(res *uiResources) widget.PreferredSizeLocateableWidget {
	c := widget.NewContainer(
		"header wrapper",
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(15))),
	)

	c.AddChild(header("Ebiten UI Demo", res,
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
	))

	c2 := widget.NewContainer(
		"second row in header",
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Left:  25,
				Right: 25,
			}),
		)),
	)
	c.AddChild(c2)

	c2.AddChild(widget.NewText(
		widget.TextOpts.Text("This program is a showcase of Ebiten UI widgets and layouts.", res.text.face, res.text.idleColor)))

	return c
}

func header(label string, res *uiResources, opts ...widget.ContainerOpt) widget.PreferredSizeLocateableWidget {
	c := widget.NewContainer("header with bg", append(opts, []widget.ContainerOpt{
		widget.ContainerOpts.BackgroundImage(res.header.background),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(widget.AnchorLayoutOpts.Padding(res.header.padding))),
	}...)...)

	c.AddChild(widget.NewText(
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
		})),
		widget.TextOpts.Text(label, res.header.face, res.header.color),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	return c
}

func demoContainer(res *uiResources, toolTips *toolTipContents, toolTip *widget.ToolTip,
	ui func() *ebitenui.UI) (widget.PreferredSizeLocateableWidget, widget.Locater, *dragContents, event.HandlerFunc) {

	demoContainer := widget.NewContainer(
		"demo",
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:  25,
				Right: 25,
			}),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(20, 0),
		)))

	dndPage, dragContents, dropHandler := dragAndDropPage(res)
	pages := []interface{}{
		anchorLayoutPage(res),
		rowLayoutPage(res),
		gridLayoutPage(res),
		buttonPage(res),
		checkboxPage(res),
		radioGroupPage(res),
		listPage(res),
		comboButtonPage(res),
		tabBookPage(res),
		sliderPage(res),
		dndPage,
		textInputPage(res),
		toolTipPage(res, toolTips, toolTip),
		windowPage(res, ui),
	}


	windowsManager := newWindowsManager(
		ui,
		res.panel.image,
		res.panel.padding,
		15,
		res.list.face,
		colornames.White)

	pageList := widget.NewList(
		widget.ListOpts.Entries(pages),
		widget.ListOpts.EntryLabelFunc(func(e interface{}) string {
			return e.(*page).title
		}),
		widget.ListOpts.ScrollContainerOpts(widget.ScrollContainerOpts.Image(res.list.image)),
		widget.ListOpts.SliderOpts(
			widget.SliderOpts.Images(res.list.track, res.list.handle),
			widget.SliderOpts.HandleSize(res.list.handleSize),
			widget.SliderOpts.TrackPadding(res.list.trackPadding),
		),
		widget.ListOpts.EntryColor(res.list.entry),
		widget.ListOpts.EntryFontFace(res.list.face),
		widget.ListOpts.EntryTextPadding(res.list.entryPadding),
		widget.ListOpts.HideHorizontalSlider(),
		widget.ListOpts.AllowReselect(),

		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			windowsManager.windowToggle(args.Entry.(*page))
		}))
	demoContainer.AddChild(pageList)

	return demoContainer, dndPage.content.(widget.Locater), dragContents, dropHandler
}

func newCheckbox(label string, changedHandler widget.CheckboxChangedHandlerFunc, res *uiResources) *widget.LabeledCheckbox {
	return widget.NewLabeledCheckbox(
		widget.LabeledCheckboxOpts.Spacing(res.checkbox.spacing),
		widget.LabeledCheckboxOpts.CheckboxOpts(
			widget.CheckboxOpts.ButtonOpts(widget.ButtonOpts.Image(res.checkbox.image)),
			widget.CheckboxOpts.Image(res.checkbox.graphic),
			widget.CheckboxOpts.ChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
				if changedHandler != nil {
					changedHandler(args)
				}
			})),
		widget.LabeledCheckboxOpts.LabelOpts(widget.LabelOpts.Text(label, res.label.face, res.label.text)))
}

func newPageContentContainer() *widget.Container {
	return widget.NewContainer(
		"page content",
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(10),
		)))
}

func newListComboButton(entries []interface{}, buttonLabel widget.SelectComboButtonEntryLabelFunc, entryLabel widget.ListEntryLabelFunc,
	entrySelectedHandler widget.ListComboButtonEntrySelectedHandlerFunc, res *uiResources) *widget.ListComboButton {

	return widget.NewListComboButton(
		widget.ListComboButtonOpts.SelectComboButtonOpts(
			widget.SelectComboButtonOpts.ComboButtonOpts(
				widget.ComboButtonOpts.ButtonOpts(
					widget.ButtonOpts.Image(res.comboButton.image),
					widget.ButtonOpts.TextPadding(res.comboButton.padding),
				),
			),
		),
		widget.ListComboButtonOpts.Text(res.comboButton.face, res.comboButton.graphic, res.comboButton.text),
		widget.ListComboButtonOpts.ListOpts(
			widget.ListOpts.Entries(entries),
			widget.ListOpts.ScrollContainerOpts(
				widget.ScrollContainerOpts.Image(res.list.image),
			),
			widget.ListOpts.SliderOpts(
				widget.SliderOpts.Images(res.list.track, res.list.handle),
				widget.SliderOpts.HandleSize(res.list.handleSize),
				widget.SliderOpts.TrackPadding(res.list.trackPadding)),
			widget.ListOpts.EntryFontFace(res.list.face),
			widget.ListOpts.EntryColor(res.list.entry),
			widget.ListOpts.EntryTextPadding(res.list.entryPadding),
		),
		widget.ListComboButtonOpts.EntryLabelFunc(buttonLabel, entryLabel),
		widget.ListComboButtonOpts.EntrySelectedHandler(entrySelectedHandler))
}

func newList(entries []interface{}, res *uiResources, widgetOpts ...widget.WidgetOpt) *widget.List {
	return widget.NewList(
		widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widgetOpts...)),
		widget.ListOpts.ScrollContainerOpts(widget.ScrollContainerOpts.Image(res.list.image)),
		widget.ListOpts.SliderOpts(
			widget.SliderOpts.Images(res.list.track, res.list.handle),
			widget.SliderOpts.HandleSize(res.list.handleSize),
			widget.SliderOpts.TrackPadding(res.list.trackPadding),
		),
		widget.ListOpts.HideHorizontalSlider(),
		widget.ListOpts.Entries(entries),
		widget.ListOpts.EntryLabelFunc(func(e interface{}) string {
			return e.(string)
		}),
		widget.ListOpts.EntryFontFace(res.list.face),
		widget.ListOpts.EntryColor(res.list.entry),
		widget.ListOpts.EntryTextPadding(res.list.entryPadding),
	)
}

func newSeparator(res *uiResources, ld interface{}) widget.PreferredSizeLocateableWidget {
	c := widget.NewContainer(
		"separator",
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Top:    20,
				Bottom: 20,
			}))),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(ld)))

	c.AddChild(widget.NewGraphic(
		widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch:   true,
			MaxHeight: 2,
		})),
		widget.GraphicOpts.ImageNineSlice(image.NewNineSliceColor(res.separatorColor)),
	))

	return c
}

func (g *game) Layout(outsideWidth int, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (g *game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.ui.SetDebugMode(widget.DebugModeNone)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		g.ui.SetDebugMode(widget.DebugModeBorderOnMouseOver)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		g.ui.SetDebugMode(widget.DebugModeBorderAlwaysShow)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF4) {
		g.ui.SetDebugMode(widget.DebugModeInputLayersAlwaysShow)
	}
	g.ui.Update()
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	g.ui.Draw(screen)
}
