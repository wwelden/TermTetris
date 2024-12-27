package Objects

type Shape struct {
	Blocks [][]string
}

func (s *Shape) Rotate() Shape {
	rotated := make([][]string, len(s.Blocks[0]))
	for i := range rotated {
		rotated[i] = make([]string, len(s.Blocks))
	}
	for i, row := range s.Blocks {
		for j, block := range row {
			rotated[j][len(s.Blocks)-i-1] = block
		}
	}
	return Shape{Blocks: rotated}
}
