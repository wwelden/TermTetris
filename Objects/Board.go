package Objects

type Board struct {
	Width  int
	Height int
	Brd    [][]byte
}

func (b *Board) Set(pos Position, cell byte) {
	b.Brd[pos.Y][pos.X] = cell
}
