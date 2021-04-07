package iobench

//go:noinline
func Read(out []byte) int {
	out[0] = 'H'
	out[1] = 'E'
	out[2] = 'L'
	out[3] = 'L'
	out[4] = 'o'
	return 5
}

// ReadWithReturn allocates []byte on heap (if not inlined)
//
// But will not allocate if inlined and it is very likely that codec will inline it
//
//go:noinline
func ReadWithReturn() []byte {
	bytes := make([]byte, 5)
	bytes[0] = 'H'
	bytes[1] = 'E'
	bytes[2] = 'L'
	bytes[3] = 'L'
	bytes[4] = 'o'
	return bytes
}
