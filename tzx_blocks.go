package tzx

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
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

type tzx_block_19 struct {
	baseBlock

	data []*symdef

	tailMs uint16
}

type tzx_block_20 struct {
	baseBlock

	duration uint16
}

type tzx_block_13 struct {
	baseBlock

	pulses []uint16
}

type tzx_block_24 struct {
	baseBlock

	loops_count int16
}

type tzx_block_25 struct {
	baseBlock
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

type symdef struct {
	t uint8
	d []uint16
}

type bitReader struct {
	src    io.Reader
	bitPos int8
	data   byte
}

func createBitReader(rdr io.Reader) *bitReader {
	return &bitReader{src: rdr, bitPos: -1, data: 0}
}

func (r *bitReader) read() uint8 {
	if r.bitPos < 0 {
		_ = binary.Read(r.src, binary.LittleEndian, &r.data)
		r.bitPos = 7
	}

	v := r.data & (1 << r.bitPos)
	r.bitPos--

	if v != 0 {
		v = 1
	}

	return v
}

func (r *bitReader) readBits(bits uint8) uint8 {
	var v uint8
	var i uint8

	v = 0
	i = 0

	for ; i < bits; i++ {
		v <<= 1
		v |= r.read()
	}

	return v
}

func readSymDefs(r io.Reader, c uint16, p uint8) []symdef {
	pp := make([]symdef, c)

	var i uint16
	for i = 0; i < c; i++ {
		s := symdef{d: make([]uint16, p)}

		_ = binary.Read(r, binary.LittleEndian, &s.t)
		_ = binary.Read(r, binary.LittleEndian, &s.d)

		pp[i] = s
	}

	return pp
}

func createBlock19(rdr *bufio.Reader, num int) *tzx_block_19 {
	b := tzx_block_19{}

	var totalBlockLength uint32

	_ = binary.Read(rdr, binary.LittleEndian, &totalBlockLength)
	_ = binary.Read(rdr, binary.LittleEndian, &b.tailMs)

	var totp uint32
	var npp uint8
	var asp uint16
	var totd uint32
	var npd uint8
	var asd uint16
	var rb uint8
	_ = binary.Read(rdr, binary.LittleEndian, &totp)
	_ = binary.Read(rdr, binary.LittleEndian, &npp)
	_ = binary.Read(rdr, binary.LittleEndian, &rb)
	asp = uint16(rb)
	if asp == 0 {
		asp = 256
	}

	_ = binary.Read(rdr, binary.LittleEndian, &totd)
	_ = binary.Read(rdr, binary.LittleEndian, &npd)
	_ = binary.Read(rdr, binary.LittleEndian, &rb)

	asd = uint16(rb)
	if asd == 0 {
		asd = 256
	}

	pd := make([]*symdef, 0)

	if totp != 0 {
		syms := readSymDefs(rdr, asp, npp)

		var i uint32
		for i = 0; i < totp; i++ {
			var symnum uint8
			var symcnt uint16
			_ = binary.Read(rdr, binary.LittleEndian, &symnum)
			_ = binary.Read(rdr, binary.LittleEndian, &symcnt)
			for ; symcnt > 0; symcnt-- {
				pd = append(pd, &syms[symnum])
			}
		}
	}

	if totd != 0 {
		syms := readSymDefs(rdr, asd, npd)

		nb := uint8(math.Ceil(math.Log2(float64(asd)))) // Symbol bits number
		//ds := uint32(math.Ceil(float64(nb * totd / 2)))  // Data bytes number

		rd := createBitReader(rdr)

		var i uint32
		for i = 0; i < totd; i++ {
			symnum := rd.readBits(nb)
			pd = append(pd, &syms[symnum])
		}
	}

	b.data = pd
	b.description = fmt.Sprintf("TZX block #%v (0x19, Generalized data block) tail = %v", num, b.tailMs)

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

func createBlock24(rdr *bufio.Reader, num int) *tzx_block_24 {
	b := tzx_block_24{}

	_ = binary.Read(rdr, binary.LittleEndian, &b.loops_count)

	b.description = fmt.Sprintf("TZX block #%v (0x24, Loop start %v", num, b.loops_count)

	return &b
}

func createBlock25(rdr *bufio.Reader, num int) *tzx_block_25 {
	b := tzx_block_25{}

	b.description = fmt.Sprintf("TZX block #%v (0x24, Loop end", num)

	return &b
}

func createBlock32(rdr *bufio.Reader, num int) *baseBlock {
	var l uint16
	_ = binary.Read(rdr, binary.LittleEndian, &l)
	rdr.Discard(int(l))

	return &baseBlock{description: fmt.Sprintf("TZX block #%v (0x32, Archive info)", num)}
}

func (b *tzx_block_24) getBlock24() *tzx_block_24 {
	return b
}

func (b *tzx_block_25) getBlock25() *tzx_block_25 {
	return b
}
