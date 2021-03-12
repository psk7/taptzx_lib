package tzx

import (
	"bufio"
	"container/list"
	"errors"
	"io"
	"math"
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
	cpuFreq         uint64
	freq            uint64
	cpuTimeStamp    uint64
	sndTimeStamp    uint64
	currentLevel    bool
	currentVol      int16
	lastRiseSamples uint64
	maxRiseSamples  uint64
	cpuTimeBase     uint64
	sndTimeBase     uint64
	wr              *bufio.Writer
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

	// Emit rise or fall
	if s.currentVol != lvl {
		riseSamples := ((s.cpuTimeStamp - s.sndTimeStamp) / s.sndTimeBase) / 2

		if riseSamples > s.maxRiseSamples {
			riseSamples = s.maxRiseSamples
		}

		actualRiseSamples := riseSamples
		if actualRiseSamples > s.lastRiseSamples {
			actualRiseSamples = s.lastRiseSamples
		}

		s.lastRiseSamples = riseSamples

		if actualRiseSamples > 0 {
			phase := 0.0
			phaseStep := math.Pi / float64(actualRiseSamples)
			amp := float64(int32(lvl) - int32(s.currentVol))

			for i := 0; i < int(actualRiseSamples); i++ {
				v := int16((-math.Cos(phase)+1)/2*amp + float64(s.currentVol))
				s.wr.WriteByte(byte(v % 256))
				s.wr.WriteByte(byte(v >> 8))

				phase += phaseStep

				s.sndTimeStamp += s.sndTimeBase
			}
		}
	}

	// Emit sustain
	for s.sndTimeStamp < s.cpuTimeStamp {
		s.wr.WriteByte(0)
		s.wr.WriteByte(byte(lvl >> 8))
		s.sndTimeStamp += s.sndTimeBase
	}

	s.currentVol = lvl
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

func (c *context) GenerateAudioTo(writer io.Writer, freq int, sync bool, trace func(string)) {
	stream := astream{freq: uint64(freq), cpuFreq: 3500000, currentLevel: false}
	stream.maxRiseSamples = uint64(float32(0.00015) * float32(freq)) // 150 mkSec

	var bb *aBuf

	if sync {
		stream.wr = bufio.NewWriter(writer)
	} else {
		bb = createABuf(&writer)
		stream.wr = bufio.NewWriter(bb)
	}

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

	if bb != nil {
		bb.WaitComplete()
	}
}
