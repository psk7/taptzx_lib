package tzx

// Bass boost filter
// Bend square pulses. Required to correct loading SpeedLock and similar protections on real ZX hardware.
// Used some hand-selected magic :)

type boostFilter struct {
	bitsPerSample     int
	frequency         int
	writer            samplesWriter
	first             bool
	previous          int32
	last              int32
	step              int16
	pulseLen          int16
	samplesThreshold1 int16
	samplesThreshold2 int16
	samplesThreshold3 int16
	samplesThreshold4 int16
}

func CreateBoostFilter(freq int, chain samplesWriter) samplesWriter {
	w := boostFilter{writer: chain, first: true, frequency: freq}

	w.samplesThreshold1 = int16(185.e-6 * float32(freq)) // 185 mkSec, 8 @ 44.1kHz
	w.samplesThreshold2 = int16(230.e-6 * float32(freq)) // 230 mkSec, 10 @ 44.1kHz
	w.samplesThreshold3 = int16(365.e-6 * float32(freq)) // 365 mkSec, 16 @ 44.1kHz
	w.samplesThreshold4 = int16(730.e-6 * float32(freq)) // 730 mkSec, 32 @ 44.1kHz

	return &w
}

func (w *boostFilter) writeSample(sample int16) {

	if sample == 0 {
		sample = -32767
	}

	if w.first {
		w.first = false
		w.writer.writeSample(sample)
		w.previous = int32(sample)
		return
	}

	w.pulseLen += 1

	delta := int32(sample) - w.previous

	isRise := delta > 17000
	isFall := delta < -17000

	wp := w.previous
	w.previous = int32(sample)

	if isRise {
		if w.pulseLen > w.samplesThreshold4 { // 32
			w.last = wp
		} else if w.pulseLen > w.samplesThreshold3 { // 16
			w.last = wp + (delta / 2)
		} else {
			w.last = 0
		}

		if w.pulseLen < w.samplesThreshold1 { // 8
			w.step = int16(delta / 8)
		} else {
			w.step = int16(delta / 12)
		}

		w.writer.writeSample(int16(w.last))
		w.pulseLen = 0
		return
	}

	if isFall && w.pulseLen > w.samplesThreshold2 { // 10
		w.last = 0
		w.step = int16(delta / 12)
		w.writer.writeSample(int16(w.last))
		w.pulseLen = 0
		return
	}

	if isFall {
		w.pulseLen = 0
	}

	if w.step != 0 {
		w.last += int32(w.step)

		if w.step > 0 && w.last > int32(sample) {
			w.last = int32(sample)
			w.step = 0
		}

		if w.step < 0 && w.last < int32(sample) {
			w.last = int32(sample)
			w.step = 0
		}

		w.writer.writeSample(int16(w.last))
		return
	}

	w.writer.writeSample(sample)
}

func (w *boostFilter) flush() {
}
