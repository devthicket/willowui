package binutil

import (
	"encoding/binary"
	"io"
	"math"
)

// Reader is a sequential binary reader with error accumulation.
// Once an error is set, all subsequent reads return zero values.
type Reader struct {
	Data []byte
	Pos  int
	Err  error
}

func (r *Reader) ReadBytes(n int) []byte {
	if r.Err != nil {
		return nil
	}
	if r.Pos+n > len(r.Data) {
		r.Err = io.ErrUnexpectedEOF
		return nil
	}
	b := make([]byte, n)
	copy(b, r.Data[r.Pos:r.Pos+n])
	r.Pos += n
	return b
}

func (r *Reader) ReadU8() uint8 {
	if r.Err != nil {
		return 0
	}
	if r.Pos >= len(r.Data) {
		r.Err = io.ErrUnexpectedEOF
		return 0
	}
	v := r.Data[r.Pos]
	r.Pos++
	return v
}

func (r *Reader) ReadU16() uint16 {
	if r.Err != nil {
		return 0
	}
	if r.Pos+2 > len(r.Data) {
		r.Err = io.ErrUnexpectedEOF
		return 0
	}
	v := uint16(r.Data[r.Pos]) | uint16(r.Data[r.Pos+1])<<8
	r.Pos += 2
	return v
}

func (r *Reader) ReadU32() uint32 {
	if r.Err != nil {
		return 0
	}
	if r.Pos+4 > len(r.Data) {
		r.Err = io.ErrUnexpectedEOF
		return 0
	}
	v := uint32(r.Data[r.Pos]) | uint32(r.Data[r.Pos+1])<<8 |
		uint32(r.Data[r.Pos+2])<<16 | uint32(r.Data[r.Pos+3])<<24
	r.Pos += 4
	return v
}

func (r *Reader) ReadFloat64() float64 {
	if r.Err != nil {
		return 0
	}
	if r.Pos+8 > len(r.Data) {
		r.Err = io.ErrUnexpectedEOF
		return 0
	}
	bits := binary.LittleEndian.Uint64(r.Data[r.Pos : r.Pos+8])
	r.Pos += 8
	return math.Float64frombits(bits)
}

func (r *Reader) ReadString() string {
	length := r.ReadU32()
	if r.Err != nil {
		return ""
	}
	if r.Pos+int(length) > len(r.Data) {
		r.Err = io.ErrUnexpectedEOF
		return ""
	}
	s := string(r.Data[r.Pos : r.Pos+int(length)])
	r.Pos += int(length)
	return s
}

func (r *Reader) ReadBool() bool {
	return r.ReadU8() != 0
}
