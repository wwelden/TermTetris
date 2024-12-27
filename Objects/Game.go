package Objects

import (
	"bytes"
	"fmt"
	"os"
	"time"
)

type Game struct {
	isRunning  bool
	GameBoard  *Board
	DrawBuffer *bytes.Buffer
}

func (g *Game) Render() {
	g.DrawBuffer.Reset()
	fmt.Fprint(g.DrawBuffer, "\033[H\033[2J")
	g.RenderBoard()
	fmt.Fprint(os.Stdout, g.DrawBuffer.String())
}

func (g *Game) FillBoard(width, height int) {
	g.GameBoard.Brd = make([][]byte, height)
	for i := range g.GameBoard.Brd {
		g.GameBoard.Brd[i] = make([]byte, width)
		for j := range g.GameBoard.Brd[i] {
			g.GameBoard.Brd[i][j] = EmptyCell
		}
	}
}

const (
	EmptyCell       = 0
	BlockCell       = 1
	EmptyCellSymbol = " "
	BlockCellSymbol = "🔳"
)

func (g *Game) RenderBoard() {
	for _, row := range g.GameBoard.Brd {
		for _, cell := range row {
			if cell == EmptyCell {
				g.DrawBuffer.WriteString(EmptyCellSymbol)
			} else {
				g.DrawBuffer.WriteString(BlockCellSymbol)
			}
		}
		g.DrawBuffer.WriteString("\n")
	}
}

func (g *Game) Start() {
	g.isRunning = true
	g.FillBoard(g.GameBoard.Width, g.GameBoard.Height)
	g.loop()
}
func (g *Game) Update() {
	//nothing for now
}

func (g *Game) loop() {
	for g.isRunning {
		g.Update()
		g.Render()
		time.Sleep(time.Millisecond * 16)

	}
}
