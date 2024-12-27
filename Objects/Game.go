package Objects

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"math/rand"
)

type Game struct {
	isRunning   bool
	GameBoard   *Board
	DrawBuffer  *bytes.Buffer
	Pieces      []*Piece
	ActivePiece *Piece
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

func (g *Game) ClearPiece(pos Position) {
	g.GameBoard.Set(pos, EmptyCell)
}

func (g *Game) UpdatePiece() {
	for _, piece := range g.Pieces {
		for i, row := range piece.shp.Blocks {
			for j, cell := range row {
				if cell == "🔳" {
					g.GameBoard.Set(Position{
						X: piece.Position.X + j,
						Y: piece.Position.Y + i,
					}, EmptyCell)
				}
			}
		}

		piece.canFall = true
		g.ActivePiece = piece
		for i, row := range piece.shp.Blocks {
			for j, cell := range row {
				if cell == "🔳" {
					nextPos := Position{
						X: piece.Position.X + j,
						Y: piece.Position.Y + i + 1,
					}
					if nextPos.Y >= g.GameBoard.Height || g.GameBoard.Brd[nextPos.Y][nextPos.X] == BlockCell || g.GameBoard.Brd[nextPos.Y][nextPos.X] == WallCell {
						piece.canFall = false
						break
					}
				}
			}
			if !piece.canFall {
				break
			}
		}
		if piece.canFall {
			piece.Position.Fall()
		}
		for i, row := range piece.shp.Blocks {
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
func (g *Game) checkCompletedRow() bool {
	for i, row := range g.GameBoard.Brd {
		isRowFull := true
		for j, cell := range row {
			if j == 0 || j == len(row)-1 {
				continue
			}
			if cell != BlockCell {
				isRowFull = false
				break
			}
		}
		if isRowFull {
			g.GameBoard.Brd[i] = make([]byte, g.GameBoard.Width)
			g.GameBoard.Brd[i][0] = WallCell
			g.GameBoard.Brd[i][g.GameBoard.Width-1] = WallCell
			return true
		}
	}
	return false
}
func (g *Game) removeCompletedRow() {
	if g.checkCompletedRow() {
		fmt.Println("Removing completed row")
		for i := len(g.GameBoard.Brd) - 1; i >= 0; i-- {
			isEmptyRow := true
			for j := 1; j < len(g.GameBoard.Brd[i])-1; j++ {
				if g.GameBoard.Brd[i][j] != EmptyCell {
					isEmptyRow = false
					break
				}
			}
			if isEmptyRow {
				for k := i; k > 0; k-- {
					g.GameBoard.Brd[k] = make([]byte, g.GameBoard.Width)
					copy(g.GameBoard.Brd[k], g.GameBoard.Brd[k-1])
				}
				g.GameBoard.Brd[0] = make([]byte, g.GameBoard.Width)
				g.GameBoard.Brd[0][0] = WallCell
				g.GameBoard.Brd[0][g.GameBoard.Width-1] = WallCell
			}
		}
	}
	if g.checkCompletedRow() {
		g.removeCompletedRow()
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
			time.Sleep(2 * time.Second) //change this to 2 seconds
		}
	}()
}

func (g *Game) RotatePiece() {
	g.ActivePiece.Rotate()
}

func (g *Game) checkForLoss() {
	for _, piece := range g.Pieces {
		if (piece.Position.Y == 0) && !piece.canFall {
			fmt.Println("Game Over!")
			g.Stop()
		}
	}
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
		// go g.KeyPressed()
		g.UpdatePiece()
		g.Update()
		g.removeCompletedRow()
		g.checkForLoss()
		time.Sleep(time.Millisecond * 100)
		// g.RotatePiece()
	}
}
