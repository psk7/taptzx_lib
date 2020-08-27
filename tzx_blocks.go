package tzx

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

type tzx_block_11 struct {
	baseBlock

	pilotPulseLen uint16
	firstSyncLen  uint16
	secondSyncLen uint16
	zeroLen       uint16
	oneLen        uint16
	pilotLen      uint16
	rem           byte
	tailMs        uint16
	data          []byte
}

type tzx_block_12 struct {
	baseBlock

	pulselen uint16
	pulses   uint16
}

type tzx_block_20 struct {
	baseBlock

	duration uint16
}

type tzx_block_13 struct {
	baseBlock

	pulses []uint16
}

func createTAPBlock(rdr *bufio.Reader, num int) *tzx_block_11 {
	b := tzx_block_11{
		pilotPulseLen: 2168,
		firstSyncLen:  667,
		secondSyncLen: 735,
		zeroLen:       855,
		oneLen:        1710,
		tailMs:        1000,
		rem:           8,
		pilotLen:      8063,
	}

	var blockSize uint16
	res := binary.Read(rdr, binary.LittleEndian, &blockSize)

	if res == io.EOF {
		return nil
	}

	b.data = make([]byte, blockSize)

	io.ReadFull(rdr, b.data)

	if b.data[0] >= 128 {
		b.pilotLen = 3223
	}

	b.description = fmt.Sprintf("TAP block #%v size = %v", num, blockSize)

	return &b
}

func createBlock10(rdr *bufio.Reader, num int) *tzx_block_11 {
	b := tzx_block_11{
		pilotPulseLen: 2168,
		firstSyncLen:  667,
		secondSyncLen: 735,
		zeroLen:       855,
		oneLen:        1710,
		tailMs:        1000,
		rem:           8,
		pilotLen:      8063,
	}

	_ = binary.Read(rdr, binary.LittleEndian, &b.tailMs)

	var dl uint16
	_ = binary.Read(rdr, binary.LittleEndian, &dl)

	b.data = make([]byte, dl)
	io.ReadFull(rdr, b.data)

	if b.data[0] >= 128 {
		b.pilotLen = 3223
	}

	b.description = fmt.Sprintf("TZX block #%v (0x10, Standard Speed Data) size = %v", num, dl)

	return &b
}

func createBlock11(rdr *bufio.Reader, num int) *tzx_block_11 {
	b := tzx_block_11{}

	_ = binary.Read(rdr, binary.LittleEndian, &b.pilotPulseLen)
	_ = binary.Read(rdr, binary.LittleEndian, &b.firstSyncLen)
	_ = binary.Read(rdr, binary.LittleEndian, &b.secondSyncLen)
	_ = binary.Read(rdr, binary.LittleEndian, &b.zeroLen)
	_ = binary.Read(rdr, binary.LittleEndian, &b.oneLen)
	_ = binary.Read(rdr, binary.LittleEndian, &b.pilotLen)
	_ = binary.Read(rdr, binary.LittleEndian, &b.rem)
	_ = binary.Read(rdr, binary.LittleEndian, &b.tailMs)

	d := make([]byte, 3)
	io.ReadFull(rdr, d)
	dl := int(d[2])<<16 + int(d[1])<<8 + int(d[0])

	b.data = make([]byte, dl)
	io.ReadFull(rdr, b.data)

	b.description = fmt.Sprintf("TZX block #%v (0x11, Turbo Speed Data) size = %v, tail = %v", num, dl, b.tailMs)

	return &b
}

func createBlock12(rdr *bufio.Reader, num int) *tzx_block_12 {
	b := tzx_block_12{}

	_ = binary.Read(rdr, binary.LittleEndian, &b.pulselen)
	_ = binary.Read(rdr, binary.LittleEndian, &b.pulses)

	b.description = fmt.Sprintf("TZX block #%v (0x12, Pure Tone) %v pulses @ %v", num, b.pulses, b.pulselen)

	return &b
}

func createBlock13(rdr *bufio.Reader, num int) *tzx_block_13 {
	b := tzx_block_13{}

	var l uint8
	_ = binary.Read(rdr, binary.LittleEndian, &l)

	b.pulses = make([]uint16, l)
	_ = binary.Read(rdr, binary.LittleEndian, &b.pulses)

	b.description = fmt.Sprintf("TZX block #%v (0x13, Pulse sequence) %v pulses", num, len(b.pulses))

	return &b
}

func createBlock14(rdr *bufio.Reader, num int) *tzx_block_11 {
	b := tzx_block_11{}

	_ = binary.Read(rdr, binary.LittleEndian, &b.zeroLen)
	_ = binary.Read(rdr, binary.LittleEndian, &b.oneLen)
	_ = binary.Read(rdr, binary.LittleEndian, &b.rem)
	_ = binary.Read(rdr, binary.LittleEndian, &b.tailMs)

	d := make([]byte, 3)
	io.ReadFull(rdr, d)

	dl := int(d[2])<<16 + int(d[1])<<8 + int(d[0])

	b.data = make([]byte, dl)
	io.ReadFull(rdr, b.data)

	b.description = fmt.Sprintf("TZX block #%v (0x14, Pure Data) size = %v, tail = %v", num, dl, b.tailMs)

	return &b
}

func createBlock20(rdr *bufio.Reader, num int) *tzx_block_20 {
	b := tzx_block_20{}

	_ = binary.Read(rdr, binary.LittleEndian, &b.duration)

	b.description = fmt.Sprintf("TZX block #%v (0x20, Pause (silence) or 'Stop the Tape') duration %v", num, b.duration)

	return &b
}

func createPause(lenMs uint16) *tzx_block_20 {
	b := tzx_block_20{duration: lenMs}
	b.description = fmt.Sprintf("Tail pause %v ms", lenMs)
	return &b
}

func createBlock21(rdr *bufio.Reader, num int) *baseBlock {
	var l uint8
	_ = binary.Read(rdr, binary.LittleEndian, &l)
	d := make([]byte, l)

	io.ReadFull(rdr, d)
	return &baseBlock{description: fmt.Sprintf("TZX block #%v (0x21, Group start) - %v", num, string(d))}
}

func createBlock30(rdr *bufio.Reader, num int) *baseBlock {
	var l uint8
	_ = binary.Read(rdr, binary.LittleEndian, &l)
	d := make([]byte, l)

	io.ReadFull(rdr, d)
	s := string(d)

	return &baseBlock{description: fmt.Sprintf("TZX block #%v (0x30, Text description) - %v", num, s)}
}

func createBlock32(rdr *bufio.Reader, num int) *baseBlock {
	var l uint16
	_ = binary.Read(rdr, binary.LittleEndian, &l)
	rdr.Discard(int(l))

	return &baseBlock{description: fmt.Sprintf("TZX block #%v (0x32, Archive info)", num)}
}
