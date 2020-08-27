Library for TAP and TZX (ZX Spectrum tape image) files decoding to audio 

    Using to generate 44.1 kHz, mono, 16 bit, PCM audio stream:
    
    	file, err := tzx.OpenFile(tap_or_tzx_file_name)
    	if err != nil {
    		panic("Error open file")
    	}

    	...

    	freq := 44100 // Audio frequency
    	file.GenerateAudioTo(target_stream_writer, freq,
        		func(s string) {  // Trace each block description. May be nil.
        			println(s)
        		})

