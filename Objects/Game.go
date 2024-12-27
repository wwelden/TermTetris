package Objects

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"math/rand"
)

type Game struct {
	isRunning  bool
	GameBoard  *Board
	DrawBuffer *bytes.Buffer
	Pieces     []*Piece
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
			if i == height-1 || j == 0 || j == width-1 {
				g.GameBoard.Brd[i][j] = WallCell
			} else {
				g.GameBoard.Brd[i][j] = EmptyCell
			}
		}
		// g.GameBoard.Brd[i][width-1] = WallCell
	}
}

const (
	EmptyCell       = 0
	BlockCell       = 1
	WallCell        = 2
	EmptyCellSymbol = "  "
	BlockCellSymbol = "🔳"
	WallCellSymbol  = "⬜️"
)

func (g *Game) RenderBoard() {
	for _, row := range g.GameBoard.Brd {
		for _, cell := range row {
			if cell == EmptyCell {
				g.DrawBuffer.WriteString(EmptyCellSymbol)
			} else if cell == BlockCell {
				g.DrawBuffer.WriteString(BlockCellSymbol)
			} else if cell == WallCell {
				g.DrawBuffer.WriteString(WallCellSymbol)
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

func (g *Game) Stop() {
	g.isRunning = false
}

func (g *Game) Update() {
	//nothing for now
}

func (g *Game) SpawnPiece(shape Shape, pos Position) {
	g.Pieces = append(g.Pieces, &Piece{Position: pos, shp: shape})
}

func (g *Game) RenderPiece(pos Position) {
	for _, piece := range g.Pieces {
		for i, row := range piece.shp.Shape {
			for j, cell := range row {
				if cell == "🔳" {
					g.GameBoard.Set(Position{
						X: piece.Position.X + j,
						Y: piece.Position.Y + i,
					}, BlockCell)
				}
			}
		}
	}
}

func (g *Game) ClearPiece(pos Position) {
	g.GameBoard.Set(pos, EmptyCell)
}

func (g *Game) UpdatePiece() {
	for _, piece := range g.Pieces {
		for i, row := range piece.shp.Shape {
			for j, cell := range row {
				if cell == "🔳" {
					g.GameBoard.Set(Position{
						X: piece.Position.X + j,
						Y: piece.Position.Y + i,
					}, EmptyCell)
				}
			}
		}

		canFall := true
		for i, row := range piece.shp.Shape {
			for j, cell := range row {
				if cell == "🔳" {
					nextPos := Position{
						X: piece.Position.X + j,
						Y: piece.Position.Y + i + 1,
					}
					if nextPos.Y >= g.GameBoard.Height || g.GameBoard.Brd[nextPos.Y][nextPos.X] == BlockCell || g.GameBoard.Brd[nextPos.Y][nextPos.X] == WallCell {
						canFall = false
						break
					}
				}
			}
			if !canFall {
				break
			}
		}

		if canFall {
			piece.Position.Fall()
		}

		// Render piece at new position
		for i, row := range piece.shp.Shape {
			for j, cell := range row {
				if cell == "🔳" {
					g.GameBoard.Set(Position{
						X: piece.Position.X + j,
						Y: piece.Position.Y + i,
					}, BlockCell)
				}
			}
		}
	}
}

func (g *Game) spawnPieces() {
	go func() {
		for {
			shapes := []Shape{Shape1, Shape2, Shape3}
			rand.Seed(time.Now().UnixNano())
			randomShape := shapes[rand.Intn(len(shapes))]
			randomX := rand.Intn(g.GameBoard.Width-4) + 1 // -4 for 3-wide shape + right wall, +1 to avoid left wall
			g.SpawnPiece(randomShape, Position{X: randomX, Y: 0})
			time.Sleep(2 * time.Second)
		}
	}()
}

func (g *Game) KeyPressed() {

	var b = make([]byte, 1)
	for g.isRunning {
		os.Stdin.Read(b)
		if b[0] == 'q' {
			g.Stop()
		}
	}
}

func (g *Game) loop() {
	g.spawnPieces()
	for g.isRunning {
		g.Render()
		go g.KeyPressed()
		g.RenderPiece(Position{X: 10, Y: 0})
		g.UpdatePiece()
		g.Update()
		time.Sleep(time.Millisecond * 16)
	}
}
