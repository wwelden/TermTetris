package Objects

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"time"

	"golang.org/x/term"
)

type Game struct {
	isRunning   bool
	gameOver    bool
	GameBoard   *Board
	DrawBuffer  *bytes.Buffer
	ActivePiece *Piece
	PressedKey  chan byte
	Score       int

	fallEvery  int
	tickCount  int
	prevTermSt *term.State
}

const (
	EmptyCell        = 0
	WallCell         = 2
	RedCell          = 3
	GreenCell        = 4
	YellowCell       = 5
	BlueCell         = 6
	PurpleCell       = 7
	OrangeCell       = 8
	BrownCell        = 9
	EmptyCellSymbol  = "  "
	WallCellSymbol   = "[]"
	RedCellSymbol    = "🟥"
	GreenCellSymbol  = "🟩"
	YellowCellSymbol = "🟨"
	BlueCellSymbol   = "🟦"
	PurpleCellSymbol = "🟪"
	OrangeCellSymbol = "🟧"
	BrownCellSymbol  = "🟫"
	ActiveSymbol     = "🔳"

	tickDuration = 16 * time.Millisecond
	fallInterval = 30 // ticks per gravity step (~480ms)
)

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

func (g *Game) symbolFor(cell byte) string {
	switch cell {
	case EmptyCell:
		return EmptyCellSymbol
	case WallCell:
		return WallCellSymbol
	case RedCell:
		return RedCellSymbol
	case GreenCell:
		return GreenCellSymbol
	case YellowCell:
		return YellowCellSymbol
	case BlueCell:
		return BlueCellSymbol
	case PurpleCell:
		return PurpleCellSymbol
	case OrangeCell:
		return OrangeCellSymbol
	case BrownCell:
		return BrownCellSymbol
	default:
		return EmptyCellSymbol
	}
}

func (g *Game) Render() {
	g.DrawBuffer.Reset()
	g.DrawBuffer.WriteString("\033[H\033[2J")
	fmt.Fprintf(g.DrawBuffer, "TermTetris  —  Score: %d   (a/d move, s soft-drop, w rotate, q quit)\r\n", g.Score)

	overlay := make(map[Position]bool)
	if g.ActivePiece != nil {
		for _, c := range g.ActivePiece.Cells() {
			overlay[c] = true
		}
	}
	pieceSym := ActiveSymbol
	if g.ActivePiece != nil {
		pieceSym = g.symbolFor(GetColorCell(g.ActivePiece.Color()))
	}
	for y, row := range g.GameBoard.Brd {
		for x, cell := range row {
			if overlay[Position{X: x, Y: y}] {
				g.DrawBuffer.WriteString(pieceSym)
			} else {
				g.DrawBuffer.WriteString(g.symbolFor(cell))
			}
		}
		g.DrawBuffer.WriteString("\r\n")
	}
	os.Stdout.Write(g.DrawBuffer.Bytes())
}

func (g *Game) Start() {
	g.Score = 0
	g.isRunning = true
	g.fallEvery = fallInterval
	g.PressedKey = make(chan byte, 16)
	if g.GameBoard.Width < 8 {
		g.GameBoard.Width = 8
	}
	if g.GameBoard.Height < 6 {
		g.GameBoard.Height = 6
	}
	g.FillBoard(g.GameBoard.Width, g.GameBoard.Height)

	if err := g.enterRawMode(); err != nil {
		fmt.Fprintln(os.Stderr, "warning: could not enter raw mode:", err)
	}
	defer g.restoreTerminal()
	defer fmt.Print("\033[?25h")
	fmt.Print("\033[?25l")

	go g.readInput()

	g.spawnNext()
	g.loop()

	g.Render()
	if g.gameOver {
		fmt.Printf("\r\nGame Over!  Final score: %d\r\n", g.Score)
	}
}

func (g *Game) Stop() {
	g.isRunning = false
}

func (g *Game) enterRawMode() error {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil
	}
	st, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	g.prevTermSt = st
	return nil
}

func (g *Game) restoreTerminal() {
	if g.prevTermSt == nil {
		return
	}
	_ = term.Restore(int(os.Stdin.Fd()), g.prevTermSt)
}

func (g *Game) readInput() {
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			return
		}
		select {
		case g.PressedKey <- buf[0]:
		default:
		}
	}
}

func (g *Game) handleInput() {
	for {
		select {
		case k := <-g.PressedKey:
			g.applyKey(k)
		default:
			return
		}
	}
}

func (g *Game) applyKey(k byte) {
	switch k {
	case 'q', 3: // q or Ctrl-C
		g.Stop()
	case 'a':
		g.tryMove(-1, 0)
	case 'd':
		g.tryMove(1, 0)
	case 's':
		g.softDrop()
	case 'w':
		g.tryRotate()
	}
}

func (g *Game) tryMove(dx, dy int) bool {
	if g.ActivePiece == nil {
		return false
	}
	newPos := Position{X: g.ActivePiece.Position.X + dx, Y: g.ActivePiece.Position.Y + dy}
	if !g.fits(g.ActivePiece.Shape(), newPos) {
		return false
	}
	g.ActivePiece.Position = newPos
	return true
}

func (g *Game) softDrop() {
	if g.tryMove(0, 1) {
		g.tickCount = 0
	}
}

func (g *Game) tryRotate() {
	if g.ActivePiece == nil {
		return
	}
	rotated := g.ActivePiece.Shape().Rotate()
	for _, dx := range []int{0, -1, 1, -2, 2} {
		newPos := Position{X: g.ActivePiece.Position.X + dx, Y: g.ActivePiece.Position.Y}
		if g.fits(rotated, newPos) {
			g.ActivePiece.SetShape(rotated)
			g.ActivePiece.Position = newPos
			return
		}
	}
}

func (g *Game) fits(shp Shape, pos Position) bool {
	for i, row := range shp.Blocks {
		for j, cell := range row {
			if cell == " " {
				continue
			}
			p := Position{X: pos.X + j, Y: pos.Y + i}
			if !g.GameBoard.InBounds(p) {
				return false
			}
			if g.GameBoard.Get(p) != EmptyCell {
				return false
			}
		}
	}
	return true
}

func (g *Game) lockPiece() {
	color := GetColorCell(g.ActivePiece.Color())
	for _, c := range g.ActivePiece.Cells() {
		g.GameBoard.Set(c, color)
	}
	g.ActivePiece = nil
}

func GetColorCell(color Color) byte {
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
		return RedCell
	}
}

func (g *Game) spawnNext() {
	shp := AllShapes[rand.Intn(len(AllShapes))]
	col := AllColors[rand.Intn(len(AllColors))]
	pieceWidth := 0
	if len(shp.Blocks) > 0 {
		pieceWidth = len(shp.Blocks[0])
	}
	maxX := g.GameBoard.Width - 1 - pieceWidth
	if maxX < 1 {
		maxX = 1
	}
	startX := 1 + rand.Intn(maxX)
	pos := Position{X: startX, Y: 0}
	if !g.fits(shp, pos) {
		g.gameOver = true
		g.Stop()
		return
	}
	g.ActivePiece = &Piece{Position: pos, shp: shp, color: col}
}

func (g *Game) gravityStep() {
	if g.ActivePiece == nil {
		return
	}
	if g.tryMove(0, 1) {
		return
	}
	g.lockPiece()
	cleared := g.clearCompletedRows()
	g.Score += scoreFor(cleared)
	g.spawnNext()
}

func scoreFor(rows int) int {
	switch rows {
	case 0:
		return 0
	case 1:
		return 100
	case 2:
		return 300
	case 3:
		return 500
	default:
		return 800
	}
}

func (g *Game) rowFull(i int) bool {
	row := g.GameBoard.Brd[i]
	for j := 1; j < len(row)-1; j++ {
		switch row[j] {
		case RedCell, GreenCell, YellowCell, BlueCell, PurpleCell, OrangeCell, BrownCell:
			continue
		default:
			return false
		}
	}
	return true
}

func (g *Game) clearCompletedRows() int {
	cleared := 0
	for i := g.GameBoard.Height - 2; i >= 0; i-- {
		for g.rowFull(i) {
			cleared++
			for k := i; k > 0; k-- {
				copy(g.GameBoard.Brd[k], g.GameBoard.Brd[k-1])
			}
			g.GameBoard.Brd[0] = make([]byte, g.GameBoard.Width)
			g.GameBoard.Brd[0][0] = WallCell
			g.GameBoard.Brd[0][g.GameBoard.Width-1] = WallCell
		}
	}
	return cleared
}

func (g *Game) loop() {
	for g.isRunning {
		g.handleInput()
		g.tickCount++
		if g.tickCount >= g.fallEvery {
			g.tickCount = 0
			g.gravityStep()
		}
		g.Render()
		time.Sleep(tickDuration)
	}
}
