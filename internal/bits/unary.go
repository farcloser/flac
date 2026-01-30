package bits

import (
	"math/bits"

	"github.com/icza/bitio"
)

// ReadUnary decodes and returns an unary coded integer, whose value is
// represented by the number of leading zeros before a one.
//
// Examples of unary coded binary on the left and decoded decimal on the right:
//
//	1       => 0
//	01      => 1
//	001     => 2
//	0001    => 3
//	00001   => 4
//	000001  => 5
//	0000001 => 6
func (br *Reader) ReadUnary() (x uint64, err error) {
	// Check buffered bits first (0-7 bits remaining from a prior read).
	if br.n > 0 {
		// Count leading zeros in the br.n high bits of br.x.
		// br.x stores bits left-aligned within the low br.n bits,
		// so shift to the top of a byte to use LeadingZeros8.
		lz := uint(bits.LeadingZeros8(br.x << (8 - br.n)))
		if lz < br.n {
			// Found a 1-bit within the buffered bits.
			br.n -= lz + 1
			// Clear consumed bits: keep only the low br.n bits.
			br.x &= (1 << br.n) - 1
			return uint64(lz), nil
		}
		// All buffered bits are zero.
		x = uint64(br.n)
		br.n = 0
		br.x = 0
	}

	// Scan whole bytes from the internal buffer.
	for {
		if br.available() == 0 {
			if err = br.fill(); err != nil {
				return 0, err
			}
		}

		b := br.buf[br.pos]
		br.pos++
		br.consumeBytes(br.pos-1, br.pos)

		if b == 0 {
			// Entire byte is zeros.
			x += 8
			continue
		}

		// Found a byte with a 1-bit. Count leading zeros.
		lz := uint(bits.LeadingZeros8(b))
		x += uint64(lz)

		// Buffer the remaining bits after the terminating 1-bit.
		br.n = 7 - lz
		br.x = b & ((1 << br.n) - 1)

		return x, nil
	}
}

// WriteUnary encodes x as an unary coded integer, whose value is represented by
// the number of leading zeros before a one.
//
// Examples of unary coded binary on the left and decoded decimal on the right:
//
//	0 => 1
//	1 => 01
//	2 => 001
//	3 => 0001
//	4 => 00001
//	5 => 000001
//	6 => 0000001
func WriteUnary(bw *bitio.Writer, x uint64) error {
	for ; x > 8; x -= 8 {
		if err := bw.WriteByte(0x0); err != nil {
			return err
		}
	}

	bits := uint64(1)
	n := byte(x + 1)
	if err := bw.WriteBits(bits, n); err != nil {
		return err
	}
	return nil
}
