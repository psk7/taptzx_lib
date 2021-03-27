package tzx

import "math"

type bassFilter struct {
	writer SamplesWriter

	samplerate float64
	bass       float64
	slope      float64
	hz         float64
	a0         float64
	a1         float64
	a2         float64
	b0         float64
	b1         float64
	b2         float64
	xn1        float64
	xn2        float64
	yn1        float64
	yn2        float64
}

func CreateBassFilter(freq int, chain SamplesWriter) SamplesWriter {
	w := bassFilter{writer: chain, samplerate: float64(freq)}

	bass := 20. // Bass gain (dB)

	// Pre init (like Audacity)
	w.slope = 0.4
	w.hz = 250.0

	w.xn1 = 0
	w.xn2 = 0
	w.yn1 = 0
	w.yn2 = 0

	// Calculate coefficients
	ww := 2. * math.Pi * w.hz / w.samplerate
	a := math.Exp(math.Log(10.0) * bass / 40)
	b := math.Sqrt((a*a+1)/w.slope - (math.Pow(a-1, 2)))

	w.b0 = a * ((a + 1) - (a-1)*math.Cos(ww) + b*math.Sin(ww))
	w.b1 = 2 * a * ((a - 1) - (a+1)*math.Cos(ww))
	w.b2 = a * ((a + 1) - (a-1)*math.Cos(ww) - b*math.Sin(ww))
	w.a0 = (a + 1) + (a-1)*math.Cos(ww) + b*math.Sin(ww)
	w.a1 = -2 * ((a - 1) + (a+1)*math.Cos(ww))
	w.a2 = (a + 1) + (a-1)*math.Cos(ww) - b*math.Sin(ww)

	return &w
}

func (w *bassFilter) WriteSample(sample int16) {
	in := float64(sample) / MaxSampleValue

	out := (w.b0*in + w.b1*w.xn1 + w.b2*w.xn2 - w.a1*w.yn1 - w.a2*w.yn2) / w.a0
	w.xn2 = w.xn1
	w.xn1 = in
	w.yn2 = w.yn1
	w.yn1 = out

	if out > 1 {
		out = 1
	}

	if out < -1 {
		out = -1
	}

	sample = int16(out * MaxSampleValue)

	w.writer.WriteSample(sample)
}

func (w *bassFilter) Flush() {
	w.writer.Flush()
}
