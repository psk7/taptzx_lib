package tzx

import (
	"bufio"
	"container/list"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func openTZX(file *os.File) (blocks *list.List) {
	var magic uint64

	rdr := bufio.NewReader(file)

	_ = binary.Read(rdr, binary.LittleEndian, &magic)

	if magic != 0x1a2165706154585a {
		panic("unknown file format")
	}

	var major byte
	var minor byte

	_ = binary.Read(rdr, binary.LittleEndian, &major)
	_ = binary.Read(rdr, binary.LittleEndian, &minor)

	b := list.New()

	var blockType byte

	cnt := 0

	for {
		res := binary.Read(rdr, binary.LittleEndian, &blockType)

		if res == io.EOF {
			break
		}

		switch blockType {
		case 0x10:
			b.PushBack(createBlock10(rdr, cnt))

		case 0x11:
			b.PushBack(createBlock11(rdr, cnt))

		case 0x12:
			b.PushBack(createBlock12(rdr, cnt))

		case 0x13:
			b.PushBack(createBlock13(rdr, cnt))

		case 0x20:
			b.PushBack(createBlock20(rdr, cnt))

		case 0x14:
			b.PushBack(createBlock14(rdr, cnt))

		case 0x21:
			b.PushBack(createBlock21(rdr, cnt))

		case 0x22:
			b.PushBack(&baseBlock{description: fmt.Sprintf("TZX block #%v (0x22, Group end)", cnt)})

		case 0x30:
			b.PushBack(createBlock30(rdr, cnt))

		case 0x32:
			b.PushBack(createBlock32(rdr, cnt))

		default:
			panic("unknown block")
		}

		cnt += 1
	}

	b.PushBack(createPause(2000))

	return b
}
