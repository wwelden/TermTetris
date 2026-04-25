package Objects

type Board struct {
	Width  int
	Height int
	Brd    [][]byte
}

func (b *Board) Set(pos Position, cell byte) {
	b.Brd[pos.Y][pos.X] = cell
}

func (b *Board) Get(pos Position) byte {
	return b.Brd[pos.Y][pos.X]
}

func (b *Board) InBounds(pos Position) bool {
	return pos.X >= 0 && pos.X < b.Width && pos.Y >= 0 && pos.Y < b.Height
}
