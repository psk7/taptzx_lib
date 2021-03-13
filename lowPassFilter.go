package tzx

type lowPassFilter struct {
	bitsPerSample int
	writer        samplesWriter
	value         int32
	smoothing     float32
	first         bool
}

func CreateLowPassFilter(chain samplesWriter) samplesWriter {
	w := lowPassFilter{writer: chain, value: 0, first: true, smoothing: 5}

	return &w
}

func (w *lowPassFilter) writeSample(sample int16) {
	if w.first {
		w.writer.writeSample(sample)
		w.first = false
		w.value = int32(sample)
		return
	}

	currentValue := int32(sample)
	w.value += int32(float32(currentValue-w.value) / w.smoothing)

	w.writer.writeSample(int16(w.value))
}

func (w *lowPassFilter) flush() {
}
