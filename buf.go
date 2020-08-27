package tzx

import (
	"io"
)

type abuf struct {
	writer *io.Writer
	ch     *chan []byte
	ex     *chan bool
	done   bool
}

func createABuf(writer *io.Writer) *abuf {
	c := make(chan []byte, 16)
	ce := make(chan bool)
	b := abuf{writer: writer, ch: &c, ex: &ce}

	go func() {
		for !b.done {
			if len(*b.ch) < 16 {
				continue
			}

			(*b.writer).Write(<-*b.ch)
		}

		for len(*b.ch) > 16 {
			(*b.writer).Write(<-*b.ch)
		}

		*b.ex <- true
	}()

	return &b
}

func (b *abuf) WaitComplete() {
	b.done = true

	<-*b.ex
}

func (a *abuf) Write(p []byte) (n int, err error) {
	*a.ch <- p
	return len(p), nil
}
