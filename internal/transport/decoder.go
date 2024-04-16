package transport

import (
	"fmt"
	"io"
)

type FrameDecoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *FrameDecoder {
	return &FrameDecoder{r: r}
}

func (f *FrameDecoder) Decode() (*Frame, error) {
	return f.decode()
}

func (f *FrameDecoder) decode() (*Frame, error) {
	// First 4-bytes is the payload len of frame
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
	var (
		payload []byte
		buf     = make([]byte, payloadlen)
	)
	for {
		n, err := f.r.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("read frame payload: %w", err)
		}

		// If entire frame is read at once, just use the temporary buffer as
		// the payload. Otherwise, correctly provision the payload slice
		if len(payload) == 0 {
			if n == cap(payload) {
				payload = buf
				break
			}
			payload = make([]byte, 0, payloadlen)
		}

		payload = append(payload, buf[:n]...)
		if len(payload) == int(payloadlen) {
			break
		}
	}

	return &Frame{
		Kind:    frmkind,
		Payload: payload,
	}, nil
}

func (f *FrameDecoder) readUInt32() (uint32, error) {
	var (
		val       uint32
		recvcount int
		buf       = make([]byte, _uint32Len)
	)
	for {
		n, err := f.r.Read(buf)
		if err != nil {
			return 0, err
		}
		// For each byte we recv, bitwise OR it into the zero-valued frmlen,
		// resulting in converting a byte buffer of len "_uint32Len" into a
		// scalar value "val". Reading happens in big-endian.
		for i := 0; i < n; i++ {
			val |= uint32(buf[i]) << (8 * uint32(_uint32Len-recvcount-1))
			recvcount++
		}
		if recvcount == _uint32Len {
			break
		}
	}
	return val, nil
}
