package tzx

func (t *tzx_block_11) generate(stream audioStream) {

	if t.pilotLen != 0 {
		// PILOT
		for i := 0; i < int(t.pilotLen); i++ {
			stream.addEdge(int(t.pilotPulseLen))
		}

		stream.addEdge(int(t.firstSyncLen))
		stream.addEdge(int(t.secondSyncLen))
	}

	for i := 0; i < len(t.data)-1; i++ {
		d := t.data[i]
		for i := 7; i >= 0; i-- {
			bit := d&(1<<i) != 0

			if bit {
				stream.addEdge(int(t.oneLen))
				stream.addEdge(int(t.oneLen))
			} else {
				stream.addEdge(int(t.zeroLen))
				stream.addEdge(int(t.zeroLen))
			}
		}
	}

	// Last byte
	d := t.data[len(t.data)-1]

	for i := 7; i >= (8 - int(t.rem)); i-- {
		bit := d&(1<<i) != 0

		if bit {
			stream.addEdge(int(t.oneLen))
			stream.addEdge(int(t.oneLen))
		} else {
			stream.addEdge(int(t.zeroLen))
			stream.addEdge(int(t.zeroLen))
		}
	}

	if t.tailMs != 0 {
		stream.addPause(int(t.tailMs))
	}
}

func (t *tzx_block_12) generate(stream audioStream) {
	for i := uint16(0); i < t.pulses; i++ {
		stream.addEdge(int(t.pulselen))
	}
}

func (t *tzx_block_13) generate(stream audioStream) {
	plen := len(t.pulses)

	for i := 0; i < plen; i++ {
		stream.addEdge(int(t.pulses[i]))
	}
}

func (t *tzx_block_19) generate(stream audioStream) {
	for _, v := range t.data {
		for i, s := range v.d {

			if s == 0 {
				break
			}

			if i != 0 {
				stream.addEdge(int(s))
				continue
			}

			switch v.t {
			case 0:
				stream.addEdge(int(s))
				break
			case 1:
				stream.continuePrevious(int(s))
				break
			case 2:
				stream.setLevel(true, int(s))
				break
			case 3:
				stream.setLevel(false, int(s))
				break
			}
		}
	}

	if t.tailMs != 0 {
		stream.addPause(int(t.tailMs))
	}
}

func (t *tzx_block_20) generate(stream audioStream) {
	stream.addPause(int(t.duration))
}
