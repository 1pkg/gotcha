package gomonkey

// Assembles a jump to a function value
func jmpToFunctionValue(to uintptr) []byte {
	return []byte{
		0x48, 0xBA,
		byte(to),
		byte(to >> 8),
		byte(to >> 16),
		byte(to >> 24),
		byte(to >> 32),
		byte(to >> 40),
		byte(to >> 48),
		byte(to >> 56), // movabs rdx,to
		0xFF, 0x22,     // jmp QWORD PTR [rdx]
	}
}

// Assembles a call to a function body
func callFunctionBody(to uintptr, before, after []byte) []byte {
	load := []byte{
		0x48, 0xBA,
		byte(to),
		byte(to >> 8),
		byte(to >> 16),
		byte(to >> 24),
		byte(to >> 32),
		byte(to >> 40),
		byte(to >> 48),
		byte(to >> 56), // movabs rdx,to
	}
	call := []byte{0xFF, 0x12} // call QWORD PTR [rdx]
	return join(
		load,
		before,
		call,
		after,
	)
}

// Pads assembly function body with nope instructions
func padWithNope(body []byte, size int) []byte {
	padBody := make([]byte, size)
	copy(padBody, body)
	for i := len(body); i < size; i++ {
		padBody[i] = 0x90
	}
	return padBody
}
