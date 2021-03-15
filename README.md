Library for TAP and TZX (ZX Spectrum tape image) files decoding to audio

    Using to generate 44.1 kHz, mono, 16 or 8 bit, PCM audio stream:
    
    	file, err := tzx.OpenFile(tap_or_tzx_file_name)
    	if err != nil {
    		panic("Error open file")
    	}

    	...

    	freq := 44100 // Audio frequency

		// Setup processing pipeline
		// Convert to desired sample size and write out
    	sw := tzx.CreateBitstreamWriter(<bits-per-sample>, target_stream)

		// Optional, apply some audio stream filters to pipeline
		sw = tzx.CreateBassFilter(freq, sw)

		// Generate audio stream
    	file.GenerateAudioTo(sw, freq,
        		func(s string) {  // Trace each block description. May be nil.
        			println(s)
        		})

