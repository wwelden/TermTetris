package main

import (
	"bytes"
	"termtetris/Objects"
)

func newGame(width, height int) *Objects.Game {
	return &Objects.Game{
		DrawBuffer: new(bytes.Buffer),
		GameBoard: &Objects.Board{
			Width:  width,
			Height: height,
			Brd:    make([][]byte, height),
		},
	}
}

func main() {
	game := newGame(40, 65)
	game.Start()
}
