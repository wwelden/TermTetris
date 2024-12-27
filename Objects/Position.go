package Objects

type Position struct {
	X int
	Y int
}

func (p *Position) Fall() {
	p.Y++
}

func (p *Position) MoveLeft() {
	p.X--
}

func (p *Position) MoveRight() {
	p.X++
}
func (p *Position) MoveDown() {
	p.Y--
}
