package tzx

import (
	"bufio"
	"container/list"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type block interface {
	getDescription() string
	generate(stream audioStream)
}

type baseBlock struct {
	description string
}

type audioStream interface {
	addEdge(len int)
	addPause(lenMs int)
}

type astream struct {
	cpuFreq      uint64
	freq         uint64
	cpuTimeStamp uint64
	sndTimeStamp uint64
	currentLevel bool
	cpuTimeBase  uint64
	sndTimeBase  uint64
	wr           *bufio.Writer
}

type context struct {
	blocks *list.List
}

func (b *baseBlock) getDescription() string {
	return b.description
}

func (b *baseBlock) generate(stream audioStream) {
}

func (s *astream) appendLevel(len int, lvl int16) {
	s.cpuTimeStamp += uint64(len) * s.cpuTimeBase

	for s.sndTimeStamp < s.cpuTimeStamp {
		s.wr.WriteByte(0)
		s.wr.WriteByte(byte(lvl >> 8))
		s.sndTimeStamp += s.sndTimeBase
	}
}

func (s *astream) addEdge(len int) {
	lvl := int16(-16384)

	if s.currentLevel {
		lvl = 16384
	}

	s.appendLevel(len, lvl)

	s.currentLevel = !s.currentLevel
}

func (s *astream) addPause(lenMs int) {
	ll := uint64(lenMs) - 1
	msl := s.cpuFreq / 1000
	s.addEdge(int(msl))

	// if last edge is fall, issue another rise for 2 ms
	if s.currentLevel {
		s.addEdge(int(msl) * 2)
		ll -= 2
	}

	s.appendLevel(int(ll*msl), 0)
	s.currentLevel = false
}

func OpenFile(fileName string) (*context, error) {
	c := context{}

	file, err := os.Open(fileName)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	fext := strings.ToLower(filepath.Ext(fileName))

	if fext == ".tap" {
		c.blocks = openTAP(file)
	} else if fext == ".tzx" {
		c.blocks = openTZX(file)
	} else {
		return nil, errors.New("unknown file format")
	}

	return &c, nil
}

func (c *context) GenerateAudioTo(writer io.Writer, freq int, trace func(string)) {
	stream := astream{freq: uint64(freq), cpuFreq: 3500000, currentLevel: false}
	stream.wr = bufio.NewWriter(writer)

	timeBase := getLCM(uint32(stream.freq), uint32(stream.cpuFreq))
	stream.cpuTimeBase = timeBase / stream.cpuFreq
	stream.sndTimeBase = timeBase / stream.freq

	for e := c.blocks.Front(); e != nil; e = e.Next() {
		b := e.Value.(block)

		dscr := b.getDescription()

		if trace != nil && dscr != "" {
			trace(dscr)
		}

		b.generate(&stream)
	}

	stream.wr.Flush()
}
