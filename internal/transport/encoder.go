package transport

import (
	"fmt"
	"io"
)

type FrameEncoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *FrameEncoder {
	return &FrameEncoder{w: w}
}

func (e *FrameEncoder) Encode(frm *Frame) error {
	if err := e.writeUInt32(uint32(len(frm.Payload))); err != nil {
		return err
	}
	if err := e.writeUInt32(uint32(frm.Kind)); err != nil {
		return err
	}
	return e.write(frm.Payload)
}

func (e *FrameEncoder) write(b []byte) error {
	var written int
	for written != len(b) {
		n, err := e.w.Write(b[written:])
		if err != nil {
			return err
		}
		written += n
	}
	// fmt.Printf("encode buf: ")
	// printBuf(b)
	return nil
}

func printBuf(buf []byte) {
	fmt.Print("[")
	for i, b := range buf {
		fmt.Printf("0b%b", b)
		if i < len(buf)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Println("]")
}

func (e *FrameEncoder) writeUInt32(val uint32) error {
	b := make([]byte, 4)
	bitmask := uint32(0xFF) // Mask for first 8 bits of uint32
	for i := range b {
		// Start with biggest most significant byte (big-endian). Shift the
		// currently selected byte into the least significant 8 bits of val,
		// and zero out all other bits with mask.
		bindex := len(b) - i - 1
		shifted := val >> (8 * bindex)
		b[i] = byte(shifted & bitmask)
	}
	return e.write(b)
}
