package binutil

import (
	"encoding/binary"
	"math"
)

// Writer is an append-only binary writer with error accumulation.
// Once an error is set, all subsequent writes are no-ops.
type Writer struct {
	Buf []byte
	Err error
}

func (w *Writer) WriteBytes(b []byte) {
	if w.Err != nil {
		return
	}
	w.Buf = append(w.Buf, b...)
}

func (w *Writer) WriteU8(v uint8) {
	if w.Err != nil {
		return
	}
	w.Buf = append(w.Buf, v)
}

func (w *Writer) WriteU16(v uint16) {
	if w.Err != nil {
		return
	}
	w.Buf = append(w.Buf, byte(v), byte(v>>8))
}

func (w *Writer) WriteU32(v uint32) {
	if w.Err != nil {
		return
	}
	w.Buf = append(w.Buf, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

func (w *Writer) WriteFloat64(v float64) {
	if w.Err != nil {
		return
	}
	bits := math.Float64bits(v)
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], bits)
	w.Buf = append(w.Buf, b[:]...)
}

func (w *Writer) WriteString(s string) {
	if w.Err != nil {
		return
	}
	w.WriteU32(uint32(len(s)))
	w.Buf = append(w.Buf, s...)
}

func (w *Writer) WriteBool(v bool) {
	if v {
		w.WriteU8(1)
	} else {
		w.WriteU8(0)
	}
}
