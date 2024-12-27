package Objects

type Piece struct {
	Position Position
	Shape    [][]string
}

var (
	Shape1 = [][]string{
		{"🔳", "🔳", "🔳"},
		{"🔳", "🔳", "🔳"},
	}
	Shape2 = [][]string{
		{"🔳", "🔳", "🔳"},
		{" ", "🔳", " "},
	}
	Shape3 = [][]string{
		{"🔳", "🔳", "🔳"},
		{"🔳", " ", " "},
	}
)

func (p *Piece) MoveLeft() {
	p.Position.MoveLeft()
}

func (p *Piece) MoveRight() {
	p.Position.MoveRight()
}
