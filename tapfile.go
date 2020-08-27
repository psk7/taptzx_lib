package tzx

import (
	"bufio"
	"container/list"
	"os"
)

func openTAP(file *os.File) (blocks *list.List) {
	b := list.New()

	rdr := bufio.NewReader(file)

	cnt := 0

	for {
		bl := createTAPBlock(rdr, cnt)

		if bl == nil {
			break
		}

		b.PushBack(bl)

		cnt++
	}

	b.PushBack(createPause(2000))

	return b
}
