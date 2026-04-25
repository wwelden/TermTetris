package Objects

type Shape struct {
	Blocks [][]string
}

func (s Shape) Rotate() Shape {
	if len(s.Blocks) == 0 {
		return s
	}
	rows := len(s.Blocks)
	cols := len(s.Blocks[0])
	rotated := make([][]string, cols)
	for i := range rotated {
		rotated[i] = make([]string, rows)
	}
	for i, row := range s.Blocks {
		for j, block := range row {
			rotated[j][rows-i-1] = block
		}
	}
	return Shape{Blocks: rotated}
}
