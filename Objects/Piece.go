package Objects

type Piece struct {
	Position Position
	shp      Shape
	color    Color
}

var (
	Shape1 = Shape{
		Blocks: [][]string{
			{"X", "X", "X"},
			{"X", "X", "X"},
		},
	}
	Shape2 = Shape{
		Blocks: [][]string{
			{"X", "X", "X"},
			{" ", "X", " "},
		},
	}
	Shape3 = Shape{
		Blocks: [][]string{
			{"X", "X", "X"},
			{"X", " ", " "},
		},
	}
	Shape4 = Shape{
		Blocks: [][]string{
			{"X", "X", "X"},
			{" ", " ", "X"},
		},
	}
	Shape5 = Shape{
		Blocks: [][]string{
			{"X", " ", "X"},
			{"X", "X", "X"},
		},
	}
	Shape6 = Shape{
		Blocks: [][]string{
			{"X", "X", "X"},
			{"X", "X", " "},
		},
	}
	Shape7 = Shape{
		Blocks: [][]string{
			{"X", "X", "X"},
			{" ", "X", "X"},
		},
	}
	Shape8 = Shape{
		Blocks: [][]string{
			{"X"},
		},
	}
	Shape9 = Shape{
		Blocks: [][]string{
			{"X", "X", "X", "X"},
		},
	}

	AllShapes = []Shape{Shape1, Shape2, Shape3, Shape4, Shape5, Shape6, Shape7, Shape8, Shape9}
	AllColors = []Color{{Red}, {Green}, {Yellow}, {Blue}, {Purple}, {Orange}, {Brown}}
)

func (p *Piece) Width() int {
	if len(p.shp.Blocks) == 0 {
		return 0
	}
	return len(p.shp.Blocks[0])
}

func (p *Piece) Height() int {
	return len(p.shp.Blocks)
}

func (p *Piece) Cells() []Position {
	out := make([]Position, 0, p.Width()*p.Height())
	for i, row := range p.shp.Blocks {
		for j, cell := range row {
			if cell != " " {
				out = append(out, Position{X: p.Position.X + j, Y: p.Position.Y + i})
			}
		}
	}
	return out
}

func (p *Piece) CellsAt(pos Position, shp Shape) []Position {
	out := make([]Position, 0)
	for i, row := range shp.Blocks {
		for j, cell := range row {
			if cell != " " {
				out = append(out, Position{X: pos.X + j, Y: pos.Y + i})
			}
		}
	}
	return out
}

func (p *Piece) Color() Color   { return p.color }
func (p *Piece) Shape() Shape   { return p.shp }
func (p *Piece) SetShape(s Shape) { p.shp = s }
