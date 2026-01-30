package frame

import (
	"io"

	"github.com/mewkiz/flac/internal/hashutil/crc16"
	"github.com/mewkiz/flac/internal/hashutil/crc8"
)

// crcReader wraps an io.Reader with inline CRC-16 and CRC-8 computation,
// replacing the nested io.TeeReader chain. CRC-8 is only active during header
// parsing; CRC-16 accumulates for the entire frame.
type crcReader struct {
	r      io.Reader
	crc16  uint16
	crc8   uint8
	doCRC8 bool
}

func (cr *crcReader) Read(p []byte) (n int, err error) {
	n, err = cr.r.Read(p)
	if n > 0 {
		data := p[:n]
		cr.crc16 = crc16.Update(cr.crc16, crc16.IBMTable, data)
		if cr.doCRC8 {
			cr.crc8 = crc8.Update(cr.crc8, crc8.ATMTable, data)
		}
	}

	return n, err
}

// EnableCRC8 resets and enables CRC-8 accumulation (used during header parsing).
func (cr *crcReader) EnableCRC8() {
	cr.crc8 = 0
	cr.doCRC8 = true
}

// DisableCRC8 stops CRC-8 accumulation (after header is verified).
func (cr *crcReader) DisableCRC8() {
	cr.doCRC8 = false
}

// CRC8 returns the accumulated CRC-8 value.
func (cr *crcReader) CRC8() uint8 {
	return cr.crc8
}

// CRC16 returns the accumulated CRC-16 value.
func (cr *crcReader) CRC16() uint16 {
	return cr.crc16
}
