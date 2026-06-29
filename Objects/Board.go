package Objects

type Board struct {
	Width  int
	Height int
	Brd    [][]byte
}

// Set writes a cell, ignoring out-of-bounds positions so a stray write can
// never panic the game loop.
func (b *Board) Set(pos Position, cell byte) {
	if !b.InBounds(pos) {
		return
	}
	b.Brd[pos.Y][pos.X] = cell
}

func (b *Board) Get(pos Position) byte {
	return b.Brd[pos.Y][pos.X]
}

func (b *Board) InBounds(pos Position) bool {
	return pos.X >= 0 && pos.X < b.Width && pos.Y >= 0 && pos.Y < b.Height
}
