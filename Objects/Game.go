package Objects

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"sort"
	"time"

	"golang.org/x/term"
)

// Game runs a variant of Tetris in which several pieces fall at once: a
// new piece spawns on a timer while earlier pieces are still descending.
// The board holds only locked cells; falling pieces are kept separately
// in Falling and overlaid at render time, so clearing rows never has to
// disturb a piece that is still in the air. The player steers ActivePiece
// (the most recently spawned falling piece); gravity applies to them all.
type Game struct {
	isRunning   bool
	gameOver    bool
	GameBoard   *Board
	DrawBuffer  *bytes.Buffer
	Falling     []*Piece
	ActivePiece *Piece
	PressedKey  chan byte
	Score       int

	fallEvery  int
	tickCount  int
	spawnEvery int
	spawnCount int
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

	tickDuration  = 16 * time.Millisecond
	fallInterval  = 30  // ticks per gravity step (~480ms)
	spawnInterval = 120 // ticks between new pieces (~1.9s)
)

func (g *Game) FillBoard(width, height int) {
	g.GameBoard.Width = width
	g.GameBoard.Height = height
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

	// Overlay every falling piece on top of the locked board. The
	// controlled piece is drawn with a distinct symbol so the player can
	// always tell which one their keys affect.
	overlay := make(map[Position]string)
	for _, p := range g.Falling {
		sym := g.symbolFor(GetColorCell(p.Color()))
		if p == g.ActivePiece {
			sym = ActiveSymbol
		}
		for _, c := range p.Cells() {
			overlay[c] = sym
		}
	}
	for y, row := range g.GameBoard.Brd {
		for x, cell := range row {
			if sym, ok := overlay[Position{X: x, Y: y}]; ok {
				g.DrawBuffer.WriteString(sym)
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
	g.spawnEvery = spawnInterval
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
	if !g.fitsWithFalling(g.ActivePiece.Shape(), newPos, g.ActivePiece) {
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
		if g.fitsWithFalling(rotated, newPos, g.ActivePiece) {
			g.ActivePiece.SetShape(rotated)
			g.ActivePiece.Position = newPos
			return
		}
	}
}

// fits reports whether shp placed at pos lands entirely on empty, in-bounds
// board cells. It considers only locked cells, not other falling pieces.
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

// fitsWithFalling is fits plus the constraint that shp must not overlap any
// other falling piece. The ignore piece (usually the one being moved) is
// excluded so a piece never collides with itself.
func (g *Game) fitsWithFalling(shp Shape, pos Position, ignore *Piece) bool {
	if !g.fits(shp, pos) {
		return false
	}
	occupied := make(map[Position]bool)
	for _, p := range g.Falling {
		if p == ignore {
			continue
		}
		for _, c := range p.Cells() {
			occupied[c] = true
		}
	}
	for i, row := range shp.Blocks {
		for j, cell := range row {
			if cell == " " {
				continue
			}
			if occupied[Position{X: pos.X + j, Y: pos.Y + i}] {
				return false
			}
		}
	}
	return true
}

func (g *Game) lockPiece(p *Piece) {
	color := GetColorCell(p.Color())
	for _, c := range p.Cells() {
		g.GameBoard.Set(c, color)
		if c.Y <= 0 {
			// A piece coming to rest against the ceiling means the stack
			// has overflowed the top of the well.
			g.gameOver = true
		}
	}
	if g.gameOver {
		g.Stop()
	}
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

// spawnNext adds one random piece at a random free column along the top and
// makes it the controlled piece. If the locked stack already fills every
// top column the game is over; if the columns are only momentarily blocked
// by other falling pieces the spawn is skipped until the next interval.
func (g *Game) spawnNext() {
	shp := AllShapes[rand.Intn(len(AllShapes))]
	col := AllColors[rand.Intn(len(AllColors))]
	pieceWidth := 0
	if len(shp.Blocks) > 0 {
		pieceWidth = len(shp.Blocks[0])
	}
	maxX := g.GameBoard.Width - 1 - pieceWidth

	onBoard := false
	var free []int
	for x := 1; x <= maxX; x++ {
		pos := Position{X: x, Y: 0}
		if !g.fits(shp, pos) {
			continue
		}
		onBoard = true
		if g.fitsWithFalling(shp, pos, nil) {
			free = append(free, x)
		}
	}
	if !onBoard {
		g.gameOver = true
		g.Stop()
		return
	}
	if len(free) == 0 {
		return // top is busy with falling pieces; try again next interval
	}
	piece := &Piece{Position: Position{X: free[rand.Intn(len(free))], Y: 0}, shp: shp, color: col}
	g.Falling = append(g.Falling, piece)
	g.ActivePiece = piece
}

// gravityStep lowers every falling piece by one row. Bottom-most pieces are
// processed first so a stack of falling pieces descends together in one
// step. A piece blocked by the board or a locked cell locks in place; a
// piece blocked only by another falling piece hovers, since the piece below
// it may yet move away.
func (g *Game) gravityStep() {
	if len(g.Falling) == 0 {
		return
	}
	order := make([]*Piece, len(g.Falling))
	copy(order, g.Falling)
	sort.SliceStable(order, func(a, b int) bool {
		return order[a].Position.Y > order[b].Position.Y
	})

	survivors := make([]*Piece, 0, len(order))
	locked := false
	for _, p := range order {
		down := Position{X: p.Position.X, Y: p.Position.Y + 1}
		switch {
		case g.fitsWithFalling(p.shp, down, p):
			p.Position = down
			survivors = append(survivors, p)
		case g.fits(p.shp, down):
			// Board allows the move, so only a falling peer is in the way:
			// hover this step rather than locking.
			survivors = append(survivors, p)
		default:
			g.lockPiece(p)
			locked = true
		}
	}
	g.Falling = survivors
	g.reassignActive()

	if locked {
		cleared := g.clearCompletedRows()
		if cleared > 0 {
			g.Score += scoreFor(cleared)
			g.settleAfterClear(cleared)
		}
	}
}

// reassignActive keeps ActivePiece pointing at a live falling piece after
// locks remove pieces from the well. The highest piece is chosen so control
// passes to whatever is nearest the top.
func (g *Game) reassignActive() {
	if slices.Contains(g.Falling, g.ActivePiece) {
		return
	}
	g.ActivePiece = nil
	for _, p := range g.Falling {
		if g.ActivePiece == nil || p.Position.Y < g.ActivePiece.Position.Y {
			g.ActivePiece = p
		}
	}
}

// settleAfterClear corrects falling pieces whose cell was overtaken when the
// locked stack shifted down past an overhang. Such a piece rides down with
// the shift, but only into positions clear of both locked cells and other
// falling pieces, and never past the bottom of the well. A piece with no
// valid spot to drop into is left where it is; the next gravity step resolves
// it. Pieces are never locked here, so a stray position can't be written
// out of bounds.
func (g *Game) settleAfterClear(cleared int) {
	for _, p := range g.Falling {
		for moved := 0; moved < cleared && !g.fitsWithFalling(p.shp, p.Position, p); moved++ {
			down := Position{X: p.Position.X, Y: p.Position.Y + 1}
			if !g.fitsWithFalling(p.shp, down, p) {
				break
			}
			p.Position = down
		}
	}
	g.reassignActive()
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
		g.spawnCount++
		if g.spawnCount >= g.spawnEvery {
			g.spawnCount = 0
			g.spawnNext()
		}
		g.Render()
		time.Sleep(tickDuration)
	}
}
