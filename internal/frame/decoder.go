package frame

import (
	"bytes"
	"fmt"
	"io"
)

type decryptor interface {
	Decrypt(in io.Reader, out io.Writer) (n int64, err error)
}

type FrameDecoder struct {
	r   io.Reader
	dec decryptor
}

func NewDecoder(r io.Reader, dec decryptor) *FrameDecoder {
	return &FrameDecoder{r: r, dec: dec}
}

func (f *FrameDecoder) Decode() (*Frame, error) {
	// First 4-bytes is the payload len
	payloadlen, err := f.readUInt32()
	if err != nil {
		return nil, fmt.Errorf("read frame len: %w", err)
	}

	// Next 4-bytes is the kind of frame
	frmkindNum, err := f.readUInt32()
	if err != nil {
		return nil, fmt.Errorf("read frame kind: %w", err)
	}
	frmkind := FrameKind(frmkindNum)
	if err := frmkind.validate(); err != nil {
		return nil, err
	}

	// Early exit when payload is zero
	if payloadlen == 0 {
		return &Frame{Kind: frmkind}, nil
	}

	// Finally, read the payload of the frame
	var payload bytes.Buffer
	if f.dec != nil {
		if _, err := f.dec.Decrypt(io.LimitReader(f.r, int64(payloadlen)), &payload); err != nil {
			return nil, err
		}
	} else {
		if _, err := io.CopyN(&payload, f.r, int64(payloadlen)); err != nil {
			return nil, err
		}
	}
	return &Frame{
		Kind:    frmkind,
		Payload: payload.Bytes(),
	}, nil
}

func (f *FrameDecoder) readUInt32() (uint32, error) {
	var (
		val uint32
		buf = make([]byte, _uint32Len)
	)
	if _, err := io.ReadFull(f.r, buf); err != nil {
		return 0, err
	}

	// For each byte we recv, bitwise OR it into the zero-valued frmlen,
	// resulting in converting a byte buffer of len "_uint32Len" into a
	// scalar value "val". Reading happens in big-endian.
	for i := 0; i < len(buf); i++ {
		val |= uint32(buf[i]) << (8 * uint32(_uint32Len-i-1))
	}
	return val, nil
}
