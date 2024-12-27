package Objects

type Piece struct {
	Position Position
	shp      Shape
	canFall  bool
}

var (
	Shape1 = Shape{
		Shape: [][]string{
			{"🔳", "🔳", "🔳"},
			{"🔳", "🔳", "🔳"},
		},
	}
	Shape2 = Shape{
		Shape: [][]string{
			{"🔳", "🔳", "🔳"},
			{" ", "🔳", " "},
		},
	}
	Shape3 = Shape{
		Shape: [][]string{
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
