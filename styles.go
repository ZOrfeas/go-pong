package main

import (
	"image"
	"image/color"
	"math"
	"strconv"

	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

var (
	white = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	grey  = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x10}
)

var keyState struct {
	W    bool
	S    bool
	Down bool
	Up   bool
}

func (pong *Pong) Layout(gtx layout.Context) layout.Dimensions {

	// time advanced from last frame
	timeDiff := gtx.Now.Sub(lastFrameTime)

	size := pong.Size
	gtx.Constraints = layout.Exact(size)

	pong.MoveBars(timeDiff)

	// handle inputs
	for _, ev := range gtx.Events(pong) {
		if x, ok := ev.(key.Event); ok {
			if x.State == key.Press {
				if x.Name == "W" {
					keyState.W = true
				}
				if x.Name == "S" {
					keyState.S = true
				}
				if x.Name == key.NameUpArrow {
					keyState.Up = true
				}
				if x.Name == key.NameDownArrow {
					keyState.Down = true
				}
				// place a ball if it doesn't exist
				if pong.Ball == nil && x.Name == key.NameSpace {
					pong.NewBall()
				}
			}
			if x.State == key.Release {
				if x.Name == "W" {
					keyState.W = false

				}
				if x.Name == "S" {
					keyState.S = false

				}
				if x.Name == key.NameUpArrow {
					keyState.Up = false

				}
				if x.Name == key.NameDownArrow {
					keyState.Down = false

				}
			}
		}
	}
	if pong.Ball != nil {
		pong.MoveBall(timeDiff)
	}

	// register to listen for inputs
	key.FocusOp{Tag: pong}.Add(gtx.Ops)
	key.InputOp{Tag: pong}.Add(gtx.Ops)

	// paint the game canvas
	canvas := clip.Rect{Max: size}.Op()
	paint.FillShape(gtx.Ops, color.NRGBA{A: 0xFF}, canvas)

	pong.Left.Layout(gtx)
	pong.Right.Layout(gtx)
	if pong.Ball != nil {
		pong.Ball.Layout(gtx)
	}
	// draw scores
	if pong.Ball == nil {
		th := material.NewTheme(gofont.Collection())

		stackl := op.Offset(f32.Pt(float32(boardSize.X)/2-60, 0+50)).Push(gtx.Ops)
		ll := material.Label(th, unit.Dp(40), strconv.Itoa(pong.LeftPoints))
		ll.Color = white
		ll.Layout(gtx)
		stackl.Pop()

		stackr := op.Offset(f32.Pt(float32(boardSize.X)/2+40, 0+50)).Push(gtx.Ops)
		lr := material.Label(th, unit.Dp(40), strconv.Itoa(pong.RightPoints))
		lr.Color = white
		lr.Layout(gtx)
		stackr.Pop()
	}

	lineWidth := int(math.Trunc(5 / 2))
	splitLine := clip.Rect{
		Min: image.Pt(size.X/2-lineWidth, 0),
		Max: image.Pt(size.X/2+lineWidth, size.Y),
	}.Op()
	paint.FillShape(gtx.Ops, grey, splitLine)

	return layout.Dimensions{Size: size}
}

func (bar *Bar) Layout(gtx layout.Context) layout.Dimensions {
	var min, max image.Point
	switch bar.Side {
	case Left:
		min = image.Pt(
			int(bar.Pos.X)-bar.Width, int(bar.Pos.Y)-(bar.Len/2))
		max = image.Pt(
			int(bar.Pos.X), int(bar.Pos.Y)+(bar.Len/2))
	case Right:
		min = image.Pt(
			int(bar.Pos.X), int(bar.Pos.Y)-(bar.Len/2))
		max = image.Pt(
			int(bar.Pos.X)+bar.Width, int(bar.Pos.Y)+(bar.Len/2))
	}
	theBar := clip.Rect{Min: min, Max: max}.Op()
	paint.FillShape(gtx.Ops, white, theBar)
	return layout.Dimensions{Size: image.Pt(bar.Width, bar.Len)}
}

func (ball *Ball) Layout(gtx layout.Context) layout.Dimensions {
	ballArea := clip.Circle{
		Center: ball.Pos,
		Radius: ballRadius,
	}.Op(gtx.Ops)
	paint.FillShape(gtx.Ops, white, ballArea)
	d := image.Pt(
		int(ballDiameter),
		int(ballDiameter),
	)
	return layout.Dimensions{Size: d}
}
