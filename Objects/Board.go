package Objects

type Board struct {
	Width  int
	Height int
	Brd    [][]byte
}

func (b *Board) Set(pos Position, cell byte) {
	if pos.Y < 0 || pos.Y >= b.Height || pos.X < 0 || pos.X >= b.Width {
		return
	}
	b.Brd[pos.Y][pos.X] = cell
}
