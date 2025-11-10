package crypto

import (
	"encoding/binary"
)

// Argon2id implementation (hybrid of Argon2i and Argon2d)
// Based on RFC 9106: https://www.rfc-editor.org/rfc/rfc9106.html

const (
	// Argon2 algorithm version
	argon2Version = 0x13 // Version 1.3

	// Block size in bytes (1024 bytes = 128 uint64)
	argon2BlockSize = 1024
	argon2QWords    = argon2BlockSize / 8 // 128 uint64 words

	// Sync points for Argon2id
	argon2SyncPoints = 4
)

// Argon2Params holds Argon2 parameters
type Argon2Params struct {
	Time        uint32 // Number of iterations
	Memory      uint32 // Memory in KiB
	Parallelism uint8  // Degree of parallelism
	KeyLen      uint32 // Output key length
}

// DefaultArgon2Params returns recommended parameters for password managers
func DefaultArgon2Params() Argon2Params {
	return Argon2Params{
		Time:        3,     // 3 iterations
		Memory:      65536, // 64 MiB
		Parallelism: 4,     // 4 parallel lanes
		KeyLen:      32,    // 32 bytes (256 bits)
	}
}

// Argon2idKey derives a key using Argon2id
func Argon2idKey(password, salt []byte, params Argon2Params) []byte {
	// Calculate memory blocks
	memoryBlocks := params.Memory / 4 // Each block is 1 KiB, memory is in KiB
	if memoryBlocks < 8*uint32(params.Parallelism) {
		memoryBlocks = 8 * uint32(params.Parallelism)
	}

	segmentLength := memoryBlocks / (uint32(params.Parallelism) * argon2SyncPoints)
	laneLength := segmentLength * argon2SyncPoints

	// Initialize memory
	memory := make([]block, memoryBlocks)

	// Initial hash H0
	h0 := argon2InitialHash(password, salt, params, memoryBlocks)

	// Fill first two blocks of each lane
	for lane := uint8(0); lane < params.Parallelism; lane++ {
		// H' = H0 || LE32(0) || LE32(lane)
		var buf [72]byte // 64 (H0) + 4 + 4
		copy(buf[:64], h0)
		binary.LittleEndian.PutUint32(buf[64:], 0)
		binary.LittleEndian.PutUint32(buf[68:], uint32(lane))

		laneOffset := uint32(lane) * laneLength

		// First block: H(H' || LE32(0))
		b0 := argon2Blake2bLong(buf[:], 1024)
		copy(memory[laneOffset][:], bytesToBlock(b0))

		// Second block: H(H' || LE32(1))
		binary.LittleEndian.PutUint32(buf[64:], 1)
		b1 := argon2Blake2bLong(buf[:], 1024)
		copy(memory[laneOffset+1][:], bytesToBlock(b1))
	}

	// Fill remaining blocks
	for pass := uint32(0); pass < params.Time; pass++ {
		for slice := uint32(0); slice < argon2SyncPoints; slice++ {
			for lane := uint8(0); lane < params.Parallelism; lane++ {
				argon2FillSegment(memory, pass, lane, slice, params, laneLength, segmentLength)
			}
		}
	}

	// Final block: XOR all last blocks of all lanes
	finalBlock := memory[(laneLength - 1)]
	for lane := uint8(1); lane < params.Parallelism; lane++ {
		laneOffset := uint32(lane) * laneLength
		lastBlock := memory[laneOffset+laneLength-1]
		for i := range finalBlock {
			finalBlock[i] ^= lastBlock[i]
		}
	}

	// Generate output key
	return argon2Blake2bLong(blockToBytes(finalBlock), int(params.KeyLen))
}

// argon2InitialHash computes H0 = H(p, τ, m, t, v, y, |P|, P, |S|, S, |K|, K, |X|, X)
func argon2InitialHash(password, salt []byte, params Argon2Params, memoryBlocks uint32) []byte {
	h, _ := NewBlake2b(64)

	// Encode parameters
	var buf [4]byte

	// Parallelism
	binary.LittleEndian.PutUint32(buf[:], uint32(params.Parallelism))
	h.Write(buf[:])

	// Tag length (key length)
	binary.LittleEndian.PutUint32(buf[:], params.KeyLen)
	h.Write(buf[:])

	// Memory size in KiB
	binary.LittleEndian.PutUint32(buf[:], params.Memory)
	h.Write(buf[:])

	// Iterations
	binary.LittleEndian.PutUint32(buf[:], params.Time)
	h.Write(buf[:])

	// Version
	binary.LittleEndian.PutUint32(buf[:], argon2Version)
	h.Write(buf[:])

	// Type (2 = Argon2id)
	binary.LittleEndian.PutUint32(buf[:], 2)
	h.Write(buf[:])

	// Password length and password
	binary.LittleEndian.PutUint32(buf[:], uint32(len(password)))
	h.Write(buf[:])
	h.Write(password)

	// Salt length and salt
	binary.LittleEndian.PutUint32(buf[:], uint32(len(salt)))
	h.Write(buf[:])
	h.Write(salt)

	// Secret (empty)
	binary.LittleEndian.PutUint32(buf[:], 0)
	h.Write(buf[:])

	// Associated data (empty)
	binary.LittleEndian.PutUint32(buf[:], 0)
	h.Write(buf[:])

	return h.Sum(nil)
}

// block represents a 1024-byte Argon2 block
type block [argon2QWords]uint64

// bytesToBlock converts bytes to a block
func bytesToBlock(b []byte) []uint64 {
	result := make([]uint64, argon2QWords)
	for i := 0; i < argon2QWords && i*8 < len(b); i++ {
		result[i] = binary.LittleEndian.Uint64(b[i*8:])
	}
	return result
}

// blockToBytes converts a block to bytes
func blockToBytes(b block) []byte {
	result := make([]byte, argon2BlockSize)
	for i, v := range b {
		binary.LittleEndian.PutUint64(result[i*8:], v)
	}
	return result
}

// argon2FillSegment fills a segment of the memory
func argon2FillSegment(memory []block, pass uint32, lane uint8, slice uint32, params Argon2Params, laneLength, segmentLength uint32) {
	var dataIndependentAddressing bool

	// Argon2id uses data-independent addressing for first half of first pass
	if pass == 0 && slice < argon2SyncPoints/2 {
		dataIndependentAddressing = true
	}

	startIndex := slice*segmentLength + 2
	if pass == 0 && slice == 0 {
		startIndex = 2 // Skip first two blocks
	}

	laneOffset := uint32(lane) * laneLength

	for i := startIndex; i < (slice+1)*segmentLength; i++ {
		prevIndex := i - 1
		if i == 0 {
			prevIndex = laneLength - 1
		}

		// Get reference block index
		var refLane, refIndex uint32
		if dataIndependentAddressing {
			refLane, refIndex = argon2IndexAlpha(pass, lane, slice, i, laneLength, segmentLength, params.Parallelism, 0, 0)
		} else {
			pseudo := memory[laneOffset+prevIndex][0]
			refLane, refIndex = argon2IndexAlpha(pass, lane, slice, i, laneLength, segmentLength, params.Parallelism, pseudo, pseudo>>32)
		}

		// Compute new block
		refBlock := memory[refLane+refIndex]
		prevBlock := memory[laneOffset+prevIndex]

		var newBlock block
		argon2ComputeBlock(&newBlock, &prevBlock, &refBlock)

		// Store with XOR for passes > 0
		if pass == 0 {
			memory[laneOffset+i] = newBlock
		} else {
			for j := range newBlock {
				memory[laneOffset+i][j] ^= newBlock[j]
			}
		}
	}
}

// argon2IndexAlpha computes reference block indices
func argon2IndexAlpha(pass uint32, lane uint8, slice, index, laneLength, segmentLength uint32, lanes uint8, pseudoRand1, pseudoRand2 uint64) (uint32, uint32) {
	var refLane uint32

	if pass == 0 && slice == 0 {
		refLane = uint32(lane)
	} else {
		refLane = uint32(pseudoRand2 % uint64(lanes))
	}

	// Calculate reference area size
	var areaSize uint32
	if pass == 0 {
		if slice == 0 {
			areaSize = index - 1
		} else {
			areaSize = slice*segmentLength + index - 1
		}
	} else {
		areaSize = laneLength - segmentLength + index - 1
	}

	if areaSize == 0 {
		areaSize = 1
	}

	// Map pseudoRand1 to reference index
	x := pseudoRand1
	x = (x * x) >> 32
	x = (uint64(areaSize) * x) >> 32
	refIndex := areaSize - 1 - uint32(x)

	if pass != 0 && slice != argon2SyncPoints-1 {
		refIndex += (slice + 1) * segmentLength
		if refIndex >= laneLength {
			refIndex -= laneLength
		}
	}

	return refLane * laneLength, refIndex
}

// argon2ComputeBlock computes G(X, Y) -> Z
func argon2ComputeBlock(out, in1, in2 *block) {
	var r, z block

	// R = X XOR Y
	for i := range r {
		r[i] = in1[i] ^ in2[i]
	}
	z = r

	// Apply Blake2b permutation
	argon2BlakeMix(&z, &r)

	// Z = Z XOR R
	for i := range out {
		out[i] = z[i] ^ in1[i] ^ in2[i]
	}
}

// argon2BlakeMix applies the Blake2b-based mixing function
func argon2BlakeMix(out, in *block) {
	// Row-wise mixing
	for i := 0; i < 8; i++ {
		argon2BlakeMixRow(out, in, i*16)
	}

	// Column-wise mixing
	var tmp block
	for i := 0; i < 8; i++ {
		for j := 0; j < 16; j++ {
			tmp[i*16+j] = out[j*8+i]
		}
	}
	for i := 0; i < 8; i++ {
		argon2BlakeMixRow(out, &tmp, i*16)
	}
}

// argon2BlakeMixRow mixes a row using Blake2b round
func argon2BlakeMixRow(out, in *block, offset int) {
	v := in[offset : offset+16]

	// Blake2b round with permutation
	argon2G(&v[0], &v[1], &v[2], &v[3], &v[4], &v[5], &v[6], &v[7])
	argon2G(&v[8], &v[9], &v[10], &v[11], &v[12], &v[13], &v[14], &v[15])

	copy(out[offset:offset+16], v)
}

// argon2G is the Blake2b G mixing function for Argon2
func argon2G(a, b, c, d, e, f, g, h *uint64) {
	*a = *a + *b + 2*uint64(uint32(*a))*uint64(uint32(*b))
	*d = rotr64(*d^*a, 32)
	*c = *c + *d + 2*uint64(uint32(*c))*uint64(uint32(*d))
	*b = rotr64(*b^*c, 24)
	*a = *a + *b + 2*uint64(uint32(*a))*uint64(uint32(*b))
	*d = rotr64(*d^*a, 16)
	*c = *c + *d + 2*uint64(uint32(*c))*uint64(uint32(*d))
	*b = rotr64(*b^*c, 63)
}

// argon2Blake2bLong generates variable-length output using Blake2b
func argon2Blake2bLong(input []byte, outLen int) []byte {
	out := make([]byte, outLen)

	if outLen <= 64 {
		// Short output: single Blake2b hash
		h, _ := NewBlake2b(outLen)
		h.Write(input)
		copy(out, h.Sum(nil))
		return out
	}

	// Long output: iterate Blake2b
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(outLen))

	h, _ := NewBlake2b(64)
	h.Write(buf[:])
	h.Write(input)
	v := h.Sum(nil)

	copy(out[:32], v)
	pos := 32

	for pos < outLen {
		h, _ = NewBlake2b(64)
		h.Write(v)
		v = h.Sum(nil)

		toCopy := 64
		if outLen-pos < 64 {
			toCopy = outLen - pos
		}
		copy(out[pos:], v[:toCopy])
		pos += toCopy
	}

	return out
}
