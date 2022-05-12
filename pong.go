package main

import (
	"image"
	"log"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
)

type Side bool

const (
	Left  Side = false
	Right Side = true
)

// Bar implements the player logic
type Bar struct {
	// Position of the bar's center
	Pos f32.Point
	// bar total length
	Len int
	// which side this bar is on
	Side Side
	// the bar's width
	Width int
}

// Ball implements the ball logic
type Ball struct {
	// Position of the ball's center
	Pos f32.Point
	// Current ball speed given as two vectors
	Speed f32.Point
	Fast  bool
}

// Pong implements pong logic.
type Pong struct {
	// Dimensions of play-area, used for bounds-checking
	Size image.Point
	// Left player
	Left *Bar
	// Right player
	Right *Bar
	// The Ball (can be nil)
	*Ball
	BallSideNext Side
	// The game score
	LeftPoints  int
	RightPoints int
}

func NewPong(boardSize image.Point,
	ballDiameter float32, barLen int, barOffset int,
	barWidth int) *Pong {
	return &Pong{
		Size: boardSize,
		Left: &Bar{
			Len:   barLen,
			Pos:   layout.FPt(image.Pt(barOffset, boardSize.Y/2)),
			Side:  Left,
			Width: barWidth,
		},
		Right: &Bar{
			Len:   barLen,
			Pos:   layout.FPt(image.Pt(boardSize.X-barOffset, boardSize.Y/2)),
			Side:  Right,
			Width: barWidth,
		},
		Ball:         nil,
		BallSideNext: Left, // could have also been implicit
	}
}

// Ball logic

func (p *Pong) NewBall() *Ball {
	var m int
	switch p.BallSideNext {
	case Left:
		m = -1
	case Right:
		m = 1
	}
	theBall := &Ball{
		Pos: f32.Pt(
			float32(p.Size.X/2+m*ballInitOffset),
			float32(p.Size.Y/2),
		),
		Speed: f32.Pt(
			float32(m)*ballInitSpeed,
			0,
		),
	}
	p.Ball = theBall
	return theBall
}

func timeToMeet(axisSpeed float32, axisDistance float32) (time.Duration, bool) {
	// check if they have different signs
	if (axisSpeed < 0) != (axisDistance < 0) {
		return 0, false
	}
	if axisSpeed == 0 {
		return 0, false
	}
	millisecsToMeet := axisDistance / axisSpeed
	return getTime(int64(millisecsToMeet)), true
}

type moveToDo byte

const (
	nothing moveToDo = iota
	reflect
	scoreL
	scoreR
	reflectLBar
	reflectRBar
)

// Advances ball based on its current speed
// without checking anything
func (p *Pong) moveNoCheck(t time.Duration) {
	xMove := p.Ball.Speed.X * float32(getTimeDiv(t))
	yMove := p.Ball.Speed.Y * float32(getTimeDiv(t))

	p.Ball.Pos.X += xMove
	p.Ball.Pos.Y += yMove
}

// handles the collision logic
func (p *Pong) handleActionBacklog(action moveToDo) {
	switch action {
	case reflect:
		p.Ball.Speed.Y = -p.Ball.Speed.Y
	case scoreL:
		p.Ball = nil
		p.BallSideNext = !p.BallSideNext
		p.LeftPoints += 1
	case scoreR:
		p.Ball = nil
		p.BallSideNext = !p.BallSideNext
		p.RightPoints += 1
	case reflectLBar:
		angleShiftScale := (p.Ball.Pos.Y - p.Left.Pos.Y) / (barLength / 2)
		angleShift := float64(angleShiftScale * (math.Pi / 4))
		// log.Println(angleShiftScale)
		p.Ball.Speed.X = -p.Ball.Speed.X
		reflectAngle := math.Atan2(
			float64(p.Ball.Speed.Y), float64(p.Ball.Speed.X))
		// log.Println(reflectAngle * 180 / math.Pi)
		newAngle := reflectAngle + angleShift
		// log.Println(newAngle * 180 / math.Pi)
		if newAngle > math.Pi/3 {
			newAngle = math.Pi / 3
		}
		if newAngle < -math.Pi/3 {
			newAngle = -math.Pi / 3
		}
		newAngleTan := math.Tan(newAngle)
		p.Ball.Speed.X = float32(math.Sqrt(
			math.Pow(float64(ballPlaySpeed), 2)) / (1 + math.Pow(newAngleTan, 2)))
		p.Ball.Speed.Y = p.Ball.Speed.X * float32(newAngleTan)
	case reflectRBar:
		angleShiftScale := (p.Ball.Pos.Y - p.Right.Pos.Y) / (barLength / 2)
		angleShift := -float64(angleShiftScale * (math.Pi / 4))
		p.Ball.Speed.X = -p.Ball.Speed.X
		reflectAngle := math.Atan2(
			float64(p.Ball.Speed.Y), float64(p.Ball.Speed.X))
		newAngle := reflectAngle + angleShift
		if newAngle < 0 {
			newAngle += math.Pi * 2
		}
		if newAngle > 8*math.Pi/6 {
			newAngle = 8 * math.Pi / 6
		}
		if newAngle < 4*math.Pi/6 {
			newAngle = 4 * math.Pi / 6
		}
		newAngleTan := math.Tan(newAngle)
		p.Ball.Speed.X = -float32(math.Sqrt(
			math.Pow(float64(ballPlaySpeed), 2) / (1 + math.Pow(newAngleTan, 2))))
		p.Ball.Speed.Y = p.Ball.Speed.X * float32(newAngleTan)
	}
}

// Full ball move logic, checks for collisions etc
func (p *Pong) MoveBall(t time.Duration) {
	// Check if it'll meet a wall or a bar
	var bestTimeToMeet time.Duration = -1
	var actionBacklog moveToDo
	// gets a time and decides whether it's better
	// if it's better, it returns true
	chooseMinTime := func(curTimeToMeet time.Duration, meet bool) bool {
		if meet && (bestTimeToMeet == -1 || curTimeToMeet < bestTimeToMeet) {
			bestTimeToMeet = curTimeToMeet
			return true
		}
		return false
	}

	// time to meet upper wall
	yToTopWall := 0 - (p.Ball.Pos.Y - ballRadius)
	if chooseMinTime(timeToMeet(p.Ball.Speed.Y, yToTopWall)) {
		actionBacklog = reflect
	}

	// time to meet bottom wall
	yToBotWall := float32(p.Size.Y) - (p.Ball.Pos.Y + ballRadius)
	if chooseMinTime(timeToMeet(p.Ball.Speed.Y, yToBotWall)) {
		actionBacklog = reflect
	}

	// time to meet left wall
	xToLeftWall := 0 - (p.Ball.Pos.X - ballRadius)
	if chooseMinTime(timeToMeet(p.Ball.Speed.X, xToLeftWall)) {
		actionBacklog = scoreL
	}

	// time to meet right wall
	xToRightWall := float32(p.Size.X) - (p.Ball.Pos.X + ballRadius)
	if chooseMinTime(timeToMeet(p.Ball.Speed.X, xToRightWall)) {
		actionBacklog = scoreR
	}

	// time to meet left bar
	xToLeftBar := p.Left.Pos.X - (p.Ball.Pos.X - ballRadius)
	lBarT, meet := timeToMeet(p.Ball.Speed.X, xToLeftBar)
	possibleY := p.Ball.Pos.Y + (p.Ball.Speed.Y * float32(getTimeDiv(lBarT)))
	if possibleY <= (p.Left.Pos.Y+barLength/2) && possibleY >= (p.Left.Pos.Y-barLength/2) {
		if chooseMinTime(lBarT, meet) {
			actionBacklog = reflectLBar
		}
	}

	// time to meet right bar
	xToRightBar := p.Right.Pos.X - (p.Ball.Pos.X + ballRadius)
	rBarT, meet := timeToMeet(p.Ball.Speed.X, xToRightBar)
	possibleY = p.Ball.Pos.Y + (p.Ball.Speed.Y * float32(getTimeDiv(rBarT)))
	if possibleY <= (p.Right.Pos.Y+barLength/2) && possibleY >= (p.Right.Pos.Y-barLength/2) {
		if chooseMinTime(rBarT, meet) {
			actionBacklog = reflectRBar
		}
	}

	// if bestTimeToMeet is set, then move so much and handle the
	// action backlog, else just complete the full move

	if (bestTimeToMeet != -1 && actionBacklog == nothing) ||
		(bestTimeToMeet == -1 && actionBacklog != nothing) {
		log.Fatal("Internal error: invalid collision state: ",
			bestTimeToMeet, actionBacklog)
	}
	if actionBacklog == nothing || bestTimeToMeet > t {
		p.moveNoCheck(t)
	} else {
		p.moveNoCheck(bestTimeToMeet)
		p.handleActionBacklog(actionBacklog)
		if p.Ball != nil {
			p.MoveBall(t - bestTimeToMeet)
		}
	}

}

// Bar logic
var (
	highestBarPos = float32(0 + barLength/2)
	lowestBarPos  = float32(boardSize.Y - barLength/2)
)

func (p *Pong) MoveBars(t time.Duration) {
	if keyState.W && !keyState.S {
		p.Left.MoveUp(t)
	}
	if keyState.S && !keyState.W {
		p.Left.MoveDown(t)
	}
	if keyState.Up && !keyState.Down {
		p.Right.MoveUp(t)
	}
	if keyState.Down && !keyState.Up {
		p.Right.MoveDown(t)
	}
}

func (b *Bar) move(m float32, t time.Duration) {
	m = float32(getTimeDiv(t)) * barSpeed * m
	candidateY := b.Pos.Y + m

	if candidateY > lowestBarPos {
		b.Pos.Y = lowestBarPos
		return
	}
	if candidateY < highestBarPos {
		b.Pos.Y = highestBarPos
		return
	}
	b.Pos.Y = candidateY
}
func (b *Bar) MoveUp(t time.Duration) {
	b.move(float32(-1), t)
}
func (b *Bar) MoveDown(t time.Duration) {
	b.move(float32(1), t)
}
