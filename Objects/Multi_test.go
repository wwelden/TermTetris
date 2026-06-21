package Objects

import (
	"bytes"
	"testing"
)

func multiTestGame(w, h int) *Game {
	g := &Game{
		DrawBuffer: new(bytes.Buffer),
		GameBoard:  &Board{Width: w, Height: h},
		PressedKey: make(chan byte, 16),
	}
	g.FillBoard(w, h)
	return g
}

func cell(g *Game, x, y int) byte { return g.GameBoard.Brd[y][x] }

// A piece must not fall through or overlap another falling piece: it hovers
// while the piece below it is still in the air.
func TestFallingPieceHoversOnFallingPeer(t *testing.T) {
	g := multiTestGame(10, 12)
	lower := &Piece{Position: Position{X: 4, Y: 6}, shp: Shape8, color: Color{Green}}
	upper := &Piece{Position: Position{X: 4, Y: 5}, shp: Shape8, color: Color{Red}}
	g.Falling = []*Piece{lower, upper}
	g.ActivePiece = upper

	g.gravityStep()

	if len(g.Falling) != 2 {
		t.Fatalf("expected both pieces airborne, got %d", len(g.Falling))
	}
	if lower.Position.Y != 7 {
		t.Errorf("lower piece should keep falling to Y=7, got Y=%d", lower.Position.Y)
	}
	// Upper either follows to 6 or holds at 5, but must never overlap the lower.
	if upper.Position.Y == lower.Position.Y {
		t.Errorf("upper piece overlapped the lower piece at Y=%d", upper.Position.Y)
	}
}

// Two stacked pieces that both rest on the floor in the same step should
// both lock, leaving an intact stack and no airborne pieces.
func TestStackedPiecesLockInOneStep(t *testing.T) {
	g := multiTestGame(10, 6)
	lower := &Piece{Position: Position{X: 4, Y: 4}, shp: Shape8, color: Color{Green}} // on floor row-1
	upper := &Piece{Position: Position{X: 4, Y: 3}, shp: Shape8, color: Color{Red}}
	g.Falling = []*Piece{lower, upper}
	g.ActivePiece = upper

	g.gravityStep()

	if len(g.Falling) != 0 {
		t.Fatalf("expected both pieces locked, %d still airborne", len(g.Falling))
	}
	if cell(g, 4, 4) != GreenCell || cell(g, 4, 3) != RedCell {
		t.Errorf("stack corrupted: (4,4)=%d (4,3)=%d", cell(g, 4, 4), cell(g, 4, 3))
	}
	if g.ActivePiece != nil {
		t.Errorf("ActivePiece should be nil once all pieces lock")
	}
}

// When the controlled piece locks but others are still falling, control
// should pass to a remaining piece rather than going nil.
func TestActiveReassignedAfterLock(t *testing.T) {
	g := multiTestGame(12, 8)
	grounded := &Piece{Position: Position{X: 3, Y: 6}, shp: Shape8, color: Color{Red}} // floor row-1
	high := &Piece{Position: Position{X: 8, Y: 1}, shp: Shape8, color: Color{Blue}}
	g.Falling = []*Piece{grounded, high}
	g.ActivePiece = grounded

	g.gravityStep()

	if g.ActivePiece != high {
		t.Fatalf("control should pass to the remaining airborne piece")
	}
	if len(g.Falling) != 1 || g.Falling[0] != high {
		t.Errorf("only the airborne piece should remain falling")
	}
}

// The player's piece cannot be moved into the column occupied by another
// falling piece.
func TestMoveBlockedByFallingPeer(t *testing.T) {
	g := multiTestGame(12, 12)
	blocker := &Piece{Position: Position{X: 5, Y: 4}, shp: Shape8, color: Color{Green}}
	active := &Piece{Position: Position{X: 4, Y: 4}, shp: Shape8, color: Color{Red}}
	g.Falling = []*Piece{blocker, active}
	g.ActivePiece = active

	g.tryMove(1, 0) // would land on blocker

	if active.Position.X != 4 {
		t.Errorf("move into a falling peer should be refused, X=%d", active.Position.X)
	}
}

// A falling piece must never count toward completing a row; only locked
// cells do. The overlay model guarantees this because falling pieces are
// not written to the board.
func TestFallingPieceDoesNotCompleteRow(t *testing.T) {
	g := multiTestGame(6, 6)
	for x := 1; x <= 3; x++ {
		g.GameBoard.Brd[4][x] = RedCell
	}
	g.Falling = []*Piece{{Position: Position{X: 4, Y: 4}, shp: Shape8, color: Color{Blue}}}
	g.ActivePiece = g.Falling[0]

	if g.clearCompletedRows() != 0 {
		t.Fatalf("a hovering piece must not complete a row")
	}
}

// When clearing rows shifts a locked overhang down onto a hovering piece,
// settleAfterClear rides the piece down so nothing is corrupted.
func TestSettleAfterClearRidesPieceDown(t *testing.T) {
	g := multiTestGame(8, 12)
	// Overhang at (3,5); full row at 6 that will clear and shift the
	// overhang from row 5 down to row 6.
	g.GameBoard.Brd[5][3] = GreenCell
	for x := 1; x <= 6; x++ {
		g.GameBoard.Brd[6][x] = RedCell
	}
	// Hovering piece tucked under the overhang at (3,6) -> after the clear
	// the overhang would land on it unless it rides down.
	piece := &Piece{Position: Position{X: 3, Y: 6}, shp: Shape8, color: Color{Blue}}
	g.Falling = []*Piece{piece}
	g.ActivePiece = piece

	cleared := g.clearCompletedRows()
	if cleared != 1 {
		t.Fatalf("expected 1 cleared row, got %d", cleared)
	}
	g.settleAfterClear(cleared)

	for _, p := range g.Falling {
		if !g.fits(p.shp, p.Position) {
			t.Errorf("falling piece left overlapping a locked cell at %v", p.Position)
		}
	}
	if cell(g, 3, 6) != GreenCell {
		t.Errorf("shifted overhang corrupted, (3,6)=%d", cell(g, 3, 6))
	}
}

// Spawning must pick a column that is clear of both walls and other falling
// pieces, and must never overlap an existing falling piece.
func TestSpawnAvoidsFallingPieces(t *testing.T) {
	g := multiTestGame(8, 12)
	// Flood the top row region with a wide falling bar so only some
	// columns are free; spawn repeatedly and verify no overlaps.
	for range 50 {
		before := len(g.Falling)
		g.spawnNext()
		if g.gameOver {
			break
		}
		if len(g.Falling) == before {
			continue // spawn skipped because top was busy; valid
		}
		// Verify the newly spawned piece overlaps nothing else.
		seen := make(map[Position]int)
		for idx, p := range g.Falling {
			for _, c := range p.Cells() {
				if _, dup := seen[c]; dup {
					t.Fatalf("two falling pieces overlap at %v", c)
				}
				seen[c] = idx
			}
		}
	}
}

// If the locked stack fills every spawn column, spawning ends the game.
func TestSpawnIntoFullTopEndsGame(t *testing.T) {
	g := multiTestGame(8, 10)
	for y := range 3 {
		for x := 1; x <= 6; x++ {
			g.GameBoard.Brd[y][x] = RedCell
		}
	}
	g.isRunning = true

	g.spawnNext()

	if !g.gameOver {
		t.Fatalf("expected game over when no spawn column is free")
	}
}

// A piece coming to rest with a cell on the top row overflows the well and
// ends the game.
func TestLockAtTopEndsGame(t *testing.T) {
	g := multiTestGame(8, 8)
	for y := 1; y <= 6; y++ {
		g.GameBoard.Brd[y][4] = GreenCell // column full up to row 1
	}
	g.isRunning = true
	piece := &Piece{Position: Position{X: 4, Y: 0}, shp: Shape8, color: Color{Red}}
	g.Falling = []*Piece{piece}
	g.ActivePiece = piece

	g.gravityStep()

	if !g.gameOver {
		t.Fatalf("expected game over when a piece locks on the top row")
	}
}

func TestControlsWithNoActivePieceAreSafe(t *testing.T) {
	g := multiTestGame(10, 10)
	g.ActivePiece = nil
	g.applyKey('a')
	g.applyKey('d')
	g.applyKey('s')
	g.applyKey('w')
}
