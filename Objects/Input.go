package Objects

import (
	"os"
	"time"
)

type Input struct {
	pressedKey byte
}

func (i *Input) KeyPressed() {
	b := make([]byte, 1)

	go func() {
		os.Stdin.SetReadDeadline(time.Now().Add(time.Millisecond * 16))
		os.Stdin.Read(b) // block only for above duration
		i.pressedKey = b[0]
	}()
}
