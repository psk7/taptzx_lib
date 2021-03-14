package tzx

import "math"

type bassFilter struct {
	writer samplesWriter

	samplerate float64
	bass       float64
	slope      float64
	hzBass     float64
	a0Bass     float64
	a1Bass     float64
	a2Bass     float64
	b0Bass     float64
	b1Bass     float64
	b2Bass     float64
	xn1Bass    float64
	xn2Bass    float64
	yn1Bass    float64
	yn2Bass    float64
}

func CreateBassFilter(freq int, chain samplesWriter) samplesWriter {
	w := bassFilter{writer: chain, samplerate: float64(freq)}

	bass := 20. // Bass gain (dB)

	// Pre init (like Audacity)
	w.slope = 0.4
	w.hzBass = 250.0

	w.xn1Bass = 0
	w.xn2Bass = 0
	w.yn1Bass = 0
	w.yn2Bass = 0

	// Calculate coefficients
	ww := 2. * math.Pi * w.hzBass / w.samplerate
	a := math.Exp(math.Log(10.0) * bass / 40)
	b := math.Sqrt((a*a+1)/w.slope - (math.Pow(a-1, 2)))

	w.b0Bass = a * ((a + 1) - (a-1)*math.Cos(ww) + b*math.Sin(ww))
	w.b1Bass = 2 * a * ((a - 1) - (a+1)*math.Cos(ww))
	w.b2Bass = a * ((a + 1) - (a-1)*math.Cos(ww) - b*math.Sin(ww))
	w.a0Bass = (a + 1) + (a-1)*math.Cos(ww) + b*math.Sin(ww)
	w.a1Bass = -2 * ((a - 1) + (a+1)*math.Cos(ww))
	w.a2Bass = (a + 1) + (a-1)*math.Cos(ww) - b*math.Sin(ww)

	return &w
}

func (w *bassFilter) writeSample(sample int16) {
	in := float64(sample) / 32768.

	out := (w.b0Bass*in + w.b1Bass*w.xn1Bass + w.b2Bass*w.xn2Bass - w.a1Bass*w.yn1Bass - w.a2Bass*w.yn2Bass) / w.a0Bass
	w.xn2Bass = w.xn1Bass
	w.xn1Bass = in
	w.yn2Bass = w.yn1Bass
	w.yn1Bass = out

	if out > 1 {
		out = 1
	}

	if out < -1 {
		out = -1
	}

	sample = int16(out * 32767)

	w.writer.writeSample(sample)
}

func (w *bassFilter) flush() {
}
