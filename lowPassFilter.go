package tzx

type lowPassFilter struct {
	bitsPerSample int
	writer        SamplesWriter
	value         int32
	smoothing     float32
	first         bool
}

func CreateLowPassFilter(chain SamplesWriter) SamplesWriter {
	w := lowPassFilter{writer: chain, value: 0, first: true, smoothing: 5}

	return &w
}

func (w *lowPassFilter) WriteSample(sample int16) {
	if w.first {
		w.writer.WriteSample(sample)
		w.first = false
		w.value = int32(sample)
		return
	}

	currentValue := int32(sample)
	w.value += int32(float32(currentValue-w.value) / w.smoothing)

	w.writer.WriteSample(int16(w.value))
}

func (w *lowPassFilter) Flush() {
	w.writer.Flush()
}
