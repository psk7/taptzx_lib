package tzx

import (
	"io"
)

type aBuf struct {
	writer *io.Writer
	ch     chan *[]byte
	ex     chan bool
	done   bool
}

func createABuf(writer *io.Writer) *aBuf {
	b := aBuf{writer: writer, ch: make(chan *[]byte, 16), ex: make(chan bool)}

	go func() {
		for !b.done {
			if len(b.ch) < cap(b.ch) {
				continue
			}

			(*b.writer).Write(*<-b.ch)
		}

		for len(b.ch) > 0 {
			(*b.writer).Write(*<-b.ch)
		}

		b.ex <- true
	}()

	return &b
}

func (buf *aBuf) WaitComplete() {
	buf.done = true
	<-buf.ex
}

func (buf *aBuf) Write(p []byte) (n int, err error) {
	if !buf.done {
		cp := make([]byte, len(p))
		copy(cp, p)
		buf.ch <- &cp
	}

	return len(p), nil
}
