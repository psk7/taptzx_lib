package tzx

func getLCM(A uint32, B uint32) uint64 {
	if A == B {
		return uint64(A)
	}

	min := uint64(A)
	max := uint64(B)

	if A > B {
		min, max = max, min
	}

	mm := min * max
	c := max

	for c < mm {
		if c%min == 0 && c%max == 0 {
			return c
		}

		c += max
	}

	return mm
}
