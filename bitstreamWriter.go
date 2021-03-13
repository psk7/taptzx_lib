package tzx

import "io"

type bitstreamWriter struct {
	bitsPerSample int
	writer        io.ByteWriter
}

func CreateBitstreamWriter(bitsPerSample int, writer io.ByteWriter) samplesWriter {
	if bitsPerSample != 8 && bitsPerSample != 16 {
		panic("bitsPerSample must be 8 or 16")
	}

	w := bitstreamWriter{bitsPerSample: bitsPerSample, writer: writer}

	return &w
}

func (w *bitstreamWriter) writeSample(sample int16) {
	if w.bitsPerSample == 8 {
		// 8 bit pcm always unsigned
		v8 := byte((sample / 256) + 128)
		_ = w.writer.WriteByte(v8)

	} else if w.bitsPerSample == 16 {
		// 16 bit pcm always signed
		lo := byte(sample % 256)
		hi := byte(sample / 256)
		_ = w.writer.WriteByte(lo)
		_ = w.writer.WriteByte(hi)
	}
}

func (w *bitstreamWriter) flush() {
}
