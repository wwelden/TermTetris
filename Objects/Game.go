package Objects

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"math/rand"
)

type Game struct {
	isRunning     bool
	GameBoard     *Board
	DrawBuffer    *bytes.Buffer
	Pieces        []*Piece
	ActivePiece   *Piece
	PressedKey    chan byte
	Score         int
	RowHasCleared bool
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
	EmptyCell        = 0
	BlockCell        = 1
	WallCell         = 2
	RedCell          = 3
	GreenCell        = 4
	YellowCell       = 5
	BlueCell         = 6
	PurpleCell       = 7
	OrangeCell       = 8
	BrownCell        = 9
	EmptyCellSymbol  = "  "
	BlockCellSymbol  = "🔳"
	WallCellSymbol   = "⬜️"
	RedCellSymbol    = "🟥"
	GreenCellSymbol  = "🟩"
	YellowCellSymbol = "🟨"
	BlueCellSymbol   = "🟦"
	PurpleCellSymbol = "🟪"
	OrangeCellSymbol = "🟧"
	BrownCellSymbol  = "🟫"
)

func (g *Game) RenderBoard() {
	for _, row := range g.GameBoard.Brd {
		for _, cell := range row {
			switch cell {
			case EmptyCell:
				g.DrawBuffer.WriteString(EmptyCellSymbol)
			case BlockCell:
				g.DrawBuffer.WriteString(BlockCellSymbol)
			case WallCell:
				g.DrawBuffer.WriteString(WallCellSymbol)
			case RedCell:
				g.DrawBuffer.WriteString(RedCellSymbol)
			case GreenCell:
				g.DrawBuffer.WriteString(GreenCellSymbol)
			case YellowCell:
				g.DrawBuffer.WriteString(YellowCellSymbol)
			case BlueCell:
				g.DrawBuffer.WriteString(BlueCellSymbol)
			case PurpleCell:
				g.DrawBuffer.WriteString(PurpleCellSymbol)
			case OrangeCell:
				g.DrawBuffer.WriteString(OrangeCellSymbol)
			case BrownCell:
				g.DrawBuffer.WriteString(BrownCellSymbol)
			}
		}
		g.DrawBuffer.WriteString("\n")
	}
}

func (g *Game) Start() {
	g.Score = 0
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

func (g *Game) SpawnPiece(shape Shape, pos Position, color Color) {
	g.Pieces = append(g.Pieces, &Piece{Position: pos, shp: shape, color: color})
}

func (g *Game) ClearPiece(pos Position) {
	g.GameBoard.Set(pos, EmptyCell)
}

func (g *Game) UpdatePiece() {
	for _, piece := range g.Pieces {
		// Clear previous position
		for i, row := range piece.shp.Blocks {
			for j, cell := range row {
				if cell != "  " {
					g.GameBoard.Set(Position{
						X: piece.Position.X + j,
						Y: piece.Position.Y + i,
					}, EmptyCell)
				}
			}
		}

		piece.canFall = true
		g.ActivePiece = piece

		// Check if piece can fall
		for i, row := range piece.shp.Blocks {
			for j, cell := range row {
				if cell != "  " {
					nextPos := Position{
						X: piece.Position.X + j,
						Y: piece.Position.Y + i + 1,
					}
					if nextPos.Y >= g.GameBoard.Height || g.GameBoard.Brd[nextPos.Y][nextPos.X] == BlockCell || g.GameBoard.Brd[nextPos.Y][nextPos.X] == WallCell || g.GameBoard.Brd[nextPos.Y][nextPos.X] == RedCell || g.GameBoard.Brd[nextPos.Y][nextPos.X] == GreenCell || g.GameBoard.Brd[nextPos.Y][nextPos.X] == YellowCell || g.GameBoard.Brd[nextPos.Y][nextPos.X] == BlueCell || g.GameBoard.Brd[nextPos.Y][nextPos.X] == PurpleCell || g.GameBoard.Brd[nextPos.Y][nextPos.X] == OrangeCell || g.GameBoard.Brd[nextPos.Y][nextPos.X] == BrownCell {
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

		// Draw piece in new position
		g.SetNewPosition(piece)
	}
}

func (g *Game) GetColorCell(color Color) byte {
	switch color.color {
	case Red:
		return RedCell
	case Green:
		return GreenCell
	case Yellow:
		return YellowCell
	case Blue:
		return BlueCell
	case Purple:
		return PurpleCell
	case Orange:
		return OrangeCell
	case Brown:
		return BrownCell
	default:
		return BlockCell
	}
}

func (g *Game) SetNewPosition(piece *Piece) {
	for i, row := range piece.shp.Blocks {
		for j, cell := range row {
			if cell != "  " {
				g.GameBoard.Set(Position{
					X: piece.Position.X + j,
					Y: piece.Position.Y + i,
				}, g.GetColorCell(piece.color))
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
			if cell != RedCell && cell != GreenCell && cell != YellowCell && cell != BlueCell && cell != PurpleCell && cell != OrangeCell && cell != BrownCell {
				isRowFull = false
				break
			}
		}
		if isRowFull {
			g.Score += 100

			g.GameBoard.Brd[i] = make([]byte, g.GameBoard.Width)
			g.GameBoard.Brd[i][0] = WallCell
			g.GameBoard.Brd[i][g.GameBoard.Width-1] = WallCell
			return true
		}
	}
	return false
}
func (g *Game) allpiecesActive() {
	g.RowHasCleared = true
	for _, piece := range g.Pieces {
		piece.canFall = true
	}
}

func (g *Game) removeCompletedRow() {
	if g.checkCompletedRow() {
		fmt.Println("Removing completed row")
		g.allpiecesActive()
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

func (g *Game) printBoard() {
	for _, row := range g.GameBoard.Brd {
		fmt.Println(row)
		// fmt.Println("\n")
	}
}
func (g *Game) spawnPieces() {
	go func() {
		for {
			shapes := []Shape{Shape1, Shape2, Shape3, Shape4, Shape5, Shape6, Shape7, Shape8, Shape9}
			colors := []Color{{Red}, {Green}, {Yellow}, {Blue}, {Purple}, {Orange}, {Brown}}
			rand.Seed(time.Now().UnixNano())
			randomShape := shapes[rand.Intn(len(shapes))]
			randomColor := colors[rand.Intn(len(colors))]
			randomX := rand.Intn(g.GameBoard.Width-4) + 1 // -4 for 3-wide shape + right wall, +1 to avoid left wall
			piece := &Piece{
				Position: Position{X: randomX, Y: 0},
				shp:      randomShape,
				color:    randomColor,
			}
			g.SpawnPiece(piece.shp, piece.Position, piece.color)
			time.Sleep(time.Second * 2) //change this to 4 seconds
		}
	}()
}

func (g *Game) RotatePiece() {
	if g.ActivePiece.rotated {
		// g.ActivePiece.Rotate()
		// g.ActivePiece.rotated = false
	} else {
		g.ActivePiece.Rotate()
		g.ActivePiece.rotated = true
	}
}

func (g *Game) checkForLoss() {
	for _, piece := range g.Pieces {
		if (piece.Position.Y == 0) && !piece.canFall {
			fmt.Println("Game Over!")
			fmt.Println("Score:", g.Score)
			// g.printBoard()
			g.Stop()
		}
	}
}

func (g *Game) MoveRight() {
	g.ActivePiece.Position.MoveRight()
}

func (g *Game) MoveLeft() {
	g.ActivePiece.Position.MoveLeft()
}

func (g *Game) MoveDown() {
	g.ActivePiece.Position.Fall()
}

func (g *Game) KeyPressed() {
	b := make([]byte, 1)

	go func() {
		os.Stdin.SetReadDeadline(time.Now().Add(time.Millisecond * 16))
		os.Stdin.Read(b) // block only for above duration
		g.PressedKey <- b[0]
	}()
}
func (g *Game) GetKeyPressed() {
	key := <-g.PressedKey
	switch key {
	case 'q':
		g.Stop()
	case 'w':
		g.RotatePiece()
	case 'a':
		g.MoveLeft()
	case 'd':
		g.MoveRight()
	case 's':
		g.MoveDown()
	}
}

func (g *Game) loop() {
	g.spawnPieces()
	for g.isRunning {
		g.Render()
		g.UpdatePiece()
		g.Update()
		g.removeCompletedRow()
		g.checkForLoss()
		time.Sleep(time.Millisecond * 16)
		// g.KeyPressed()
		// g.GetKeyPressed()

	}
}
