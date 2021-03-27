package tzx

import (
	"container/list"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const MaxSampleValue = 32767

type block interface {
	getDescription() string
	generate(stream audioStream)

	getBlock24() *tzx_block_24
	getBlock25() *tzx_block_25
}

type baseBlock struct {
	description string
}

type audioStream interface {
	addEdge(len int)
	continuePrevious(len int)
	setLevel(level bool, len int)
	addPause(lenMs int)
}

type SamplesWriter interface {
	WriteSample(sample int16)
	Flush()
}

type astream struct {
	cpuFreq      uint64
	freq         uint64
	cpuTimeStamp uint64
	sndTimeStamp uint64
	currentLevel bool
	cpuTimeBase  uint64
	sndTimeBase  uint64
	wr           SamplesWriter
}

type context struct {
	blocks      *list.List
	loopLabel   *list.Element
	loopCounter int
}

func (b *baseBlock) getDescription() string {
	return b.description
}

func (b *baseBlock) generate(stream audioStream) {
}

func (b *baseBlock) getBlock24() *tzx_block_24 {
	return nil
}

func (b *baseBlock) getBlock25() *tzx_block_25 {
	return nil
}

func (s *astream) appendLevel(len int, lvl int16) {
	s.cpuTimeStamp += uint64(len) * s.cpuTimeBase

	for s.sndTimeStamp < s.cpuTimeStamp {
		s.wr.WriteSample(lvl)
		s.sndTimeStamp += s.sndTimeBase
	}
}

func (s *astream) appendLevelBool(len int, level bool) {
	// Generated samples ALWAYS is 16 bit signed. Levels MUST be LOW:-32767, HIGH:32767
	// BitstreamWriter will convert its to desired format later
	lvl := int16(-MaxSampleValue)

	if level {
		lvl = -lvl
	}

	s.appendLevel(len, lvl)
}

func (s *astream) addEdge(len int) {
	s.appendLevelBool(len, s.currentLevel)

	s.currentLevel = !s.currentLevel
}

func (s *astream) continuePrevious(len int) {
	s.appendLevelBool(len, s.currentLevel)
}

func (s *astream) setLevel(level bool, len int) {
	s.appendLevelBool(len, s.currentLevel)

	s.currentLevel = level
}

func (s *astream) addPause(lenMs int) {
	if lenMs == 0 {
		return
	}

	s.appendLevelBool(int(uint64(lenMs)*(s.cpuFreq/1000)), s.currentLevel)

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

func (c *context) GenerateAudioTo(writer SamplesWriter, freq int, trace func(string)) {
	stream := astream{freq: uint64(freq), cpuFreq: 3500000, currentLevel: false, wr: writer}

	timeBase := getLCM(uint32(stream.freq), uint32(stream.cpuFreq))
	stream.cpuTimeBase = timeBase / stream.cpuFreq
	stream.sndTimeBase = timeBase / stream.freq

	for e := c.blocks.Front(); e != nil; {
		b := e.Value.(block)

		loopStart := b.getBlock24()
		if loopStart != nil {
			// loop start
			e = e.Next()
			c.loopLabel = e
			c.loopCounter = int(loopStart.loops_count)
			continue
		}

		loopEnd := b.getBlock25()
		if loopEnd != nil {
			// loop end
			c.loopCounter -= 1

			if c.loopCounter == 0 {
				e = e.Next()
			} else {
				e = c.loopLabel
			}
			continue
		}

		dscr := b.getDescription()

		if trace != nil && dscr != "" {
			trace(dscr)
		}

		b.generate(&stream)

		e = e.Next()
	}

	writer.Flush()
}
