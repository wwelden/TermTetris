package Objects

import (
	"bytes"
	"testing"
)

func newTestGame(w, h int) *Game {
	g := &Game{
		DrawBuffer: new(bytes.Buffer),
		GameBoard:  &Board{Width: w, Height: h},
	}
	g.FillBoard(w, h)
	return g
}

func fillRow(g *Game, y int, cell byte) {
	for x := 1; x < g.GameBoard.Width-1; x++ {
		g.GameBoard.Brd[y][x] = cell
	}
}

func TestRowFullDetection(t *testing.T) {
	g := newTestGame(8, 6)
	if g.rowFull(2) {
		t.Fatal("empty row reported full")
	}
	fillRow(g, 2, RedCell)
	if !g.rowFull(2) {
		t.Fatal("filled row not detected")
	}
	g.GameBoard.Brd[2][3] = EmptyCell
	if g.rowFull(2) {
		t.Fatal("row with gap reported full")
	}
}

func TestClearSingleRow(t *testing.T) {
	g := newTestGame(8, 6)
	g.GameBoard.Brd[1][2] = BlueCell
	fillRow(g, 3, RedCell)

	cleared := g.clearCompletedRows()
	if cleared != 1 {
		t.Fatalf("cleared = %d, want 1", cleared)
	}
	// Block previously at (2,1) should now be at (2,2) after the shift.
	if g.GameBoard.Brd[2][2] != BlueCell {
		t.Errorf("expected BlueCell to shift down to row 2, got %d", g.GameBoard.Brd[2][2])
	}
	if g.GameBoard.Brd[1][2] != EmptyCell {
		t.Errorf("expected row 1 to be empty after shift, got %d", g.GameBoard.Brd[1][2])
	}
	// Walls preserved on row 0.
	if g.GameBoard.Brd[0][0] != WallCell || g.GameBoard.Brd[0][g.GameBoard.Width-1] != WallCell {
		t.Errorf("walls not preserved on top row")
	}
}

func TestClearMultipleRows(t *testing.T) {
	g := newTestGame(8, 8)
	fillRow(g, 4, RedCell)
	fillRow(g, 5, GreenCell)
	fillRow(g, 6, BlueCell)

	cleared := g.clearCompletedRows()
	if cleared != 3 {
		t.Fatalf("cleared = %d, want 3", cleared)
	}
	// Bottom playable row (height-2 = 6) should now be empty.
	for x := 1; x < g.GameBoard.Width-1; x++ {
		if g.GameBoard.Brd[6][x] != EmptyCell {
			t.Errorf("row 6 col %d = %d, want EmptyCell", x, g.GameBoard.Brd[6][x])
		}
	}
}

func TestScoreFor(t *testing.T) {
	cases := map[int]int{0: 0, 1: 100, 2: 300, 3: 500, 4: 800, 5: 800}
	for n, want := range cases {
		if got := scoreFor(n); got != want {
			t.Errorf("scoreFor(%d) = %d, want %d", n, got, want)
		}
	}
}

func TestFitsBounds(t *testing.T) {
	g := newTestGame(8, 6)
	shp := Shape{Blocks: [][]string{{"X", "X"}}}
	if !g.fits(shp, Position{X: 1, Y: 0}) {
		t.Errorf("piece should fit at (1,0)")
	}
	if g.fits(shp, Position{X: 0, Y: 0}) {
		t.Errorf("piece overlapping left wall should not fit")
	}
	if g.fits(shp, Position{X: g.GameBoard.Width - 2, Y: 0}) {
		t.Errorf("piece overlapping right wall should not fit")
	}
	if g.fits(shp, Position{X: 1, Y: g.GameBoard.Height - 1}) {
		t.Errorf("piece on bottom wall row should not fit")
	}
}

func TestFitsOverlap(t *testing.T) {
	g := newTestGame(8, 6)
	g.GameBoard.Brd[2][3] = RedCell
	shp := Shape{Blocks: [][]string{{"X"}}}
	if g.fits(shp, Position{X: 3, Y: 2}) {
		t.Errorf("piece overlapping existing block should not fit")
	}
	if !g.fits(shp, Position{X: 4, Y: 2}) {
		t.Errorf("piece in empty cell should fit")
	}
}

func addFalling(g *Game, p *Piece) {
	g.Falling = append(g.Falling, p)
	g.ActivePiece = p
}

func TestGravityLocksOnFloor(t *testing.T) {
	g := newTestGame(8, 6)
	addFalling(g, &Piece{
		Position: Position{X: 3, Y: g.GameBoard.Height - 3}, // one above bottom wall
		shp:      Shape{Blocks: [][]string{{"X"}}},
		color:    Color{Red},
	})
	g.gravityStep()
	// Should have fallen one row and not yet locked.
	if g.ActivePiece == nil {
		t.Fatal("active piece should not have locked yet")
	}
	if g.ActivePiece.Position.Y != g.GameBoard.Height-2 {
		t.Errorf("piece Y = %d, want %d", g.ActivePiece.Position.Y, g.GameBoard.Height-2)
	}
	// Next step locks it into the board and removes it from the air.
	g.gravityStep()
	if g.GameBoard.Brd[g.GameBoard.Height-2][3] != RedCell {
		t.Errorf("piece should have locked at (3, h-2), got %d", g.GameBoard.Brd[g.GameBoard.Height-2][3])
	}
	if len(g.Falling) != 0 {
		t.Errorf("locked piece should leave the falling list, %d remain", len(g.Falling))
	}
	if g.ActivePiece != nil {
		t.Errorf("no falling pieces remain, ActivePiece should be nil")
	}
}

func TestPositionMoves(t *testing.T) {
	p := Position{X: 5, Y: 5}
	p.Fall()
	if p.Y != 6 {
		t.Errorf("Fall: Y = %d, want 6", p.Y)
	}
	p.MoveLeft()
	if p.X != 4 {
		t.Errorf("MoveLeft: X = %d, want 4", p.X)
	}
	p.MoveRight()
	if p.X != 5 {
		t.Errorf("MoveRight: X = %d, want 5", p.X)
	}
	p.MoveUp()
	if p.Y != 5 {
		t.Errorf("MoveUp: Y = %d, want 5", p.Y)
	}
}
