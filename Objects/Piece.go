package Objects

type Piece struct {
	Position Position
	shp      Shape
	canFall  bool
}

var (
	Shape1 = Shape{
		Blocks: [][]string{
			{"🔳", "🔳", "🔳"},
			{"🔳", "🔳", "🔳"},
		},
	}
	Shape2 = Shape{
		Blocks: [][]string{
			{"🔳", "🔳", "🔳"},
			{" ", "🔳", " "},
		},
	}
	Shape3 = Shape{
		Blocks: [][]string{
			{"🔳", "🔳", "🔳"},
			{"🔳", " ", " "},
		},
	}
)

func (p *Piece) MoveLeft() {
	p.Position.MoveLeft()
}

func (p *Piece) MoveRight() {
	p.Position.MoveRight()
}

func (p *Piece) MoveDown() {
	p.Position.MoveDown()
}
func (p *Piece) Rotate() {
	p.shp = p.shp.Rotate()
}
