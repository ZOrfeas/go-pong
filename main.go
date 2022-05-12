package main

import (
	"image"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

var (
	boardSize = image.Pt(800, 600)
)

func getTimeDiv(d time.Duration) int64 {
	return d.Milliseconds()
}
func getTime(i int64) time.Duration {
	return time.Duration(i) * timeDiv
}

const (
	timeDiv = time.Millisecond

	barLength = 100
	barOffset = 20
	barWidth  = 10
	barSpeed  = float32(750000) / float32(int64(timeDiv)) // one pixel per microsecond

	ballDiameter   = float32(20)
	ballRadius     = ballDiameter / 2
	ballInitOffset = 20
	ballInitSpeed  = float32(200000) / float32(int64(timeDiv))
	ballPlaySpeed  = float32(350000) / float32(int64(timeDiv))
)

func main() {

	ui := NewUI()

	windowWidth := unit.Dp(float32(boardSize.X + 10))
	WindowHeight := unit.Dp(float32(boardSize.Y + 10))

	go func() {
		w := app.NewWindow(
			app.Title("Pong"),
			app.Size(windowWidth, WindowHeight),
		)
		if err := ui.Run(w); err != nil {
			log.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	app.Main()
}

type UI struct {
	*Pong
}

func NewUI() *UI {
	pong := NewPong(
		boardSize,
		ballDiameter,
		barLength,
		barOffset,
		barWidth,
	)
	return &UI{
		Pong: pong,
	}
}

var lastFrameTime time.Time

func (ui *UI) Run(w *app.Window) error {
	var ops op.Ops

	go func() {
		frameGenerator := time.Tick(time.Second / 120)
		for range frameGenerator {
			w.Invalidate()
		}
	}()

	lastFrameTime = time.Now()
	for e := range w.Events() {
		switch e := e.(type) {
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			ui.Layout(gtx)
			lastFrameTime = gtx.Now

			e.Frame(gtx.Ops)
		case system.DestroyEvent:
			return e.Err
		}
	}
	return nil
}

func (ui *UI) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Center.Layout(gtx,
		ui.Pong.Layout,
	)
}
