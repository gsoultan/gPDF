package runlength

import (
	"bytes"
	"fmt"
	"io"

	"gpdf/stream"
)

const eodMarker = 128

// Filter implements stream.Filter for RunLengthDecode (PDF run-length encoding).
type Filter struct{}

// NewFilter returns a RunLengthDecode filter.
func NewFilter() stream.Filter {
	return Filter{}
}

// Decode decompresses PDF run-length encoded data.
// Format: each run begins with a length byte n.
//   - n == 128: end of data
//   - 0 <= n <= 127: copy the next n+1 literal bytes
//   - 129 <= n <= 255: repeat the next byte (257-n) times
func (Filter) Decode(dst io.Writer, src io.Reader, _ string) error {
	data, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("runlength: read: %w", err)
	}
	decoded, err := decode(data)
	if err != nil {
		return err
	}
	_, err = dst.Write(decoded)
	return err
}

// Encode compresses data using PDF run-length encoding.
func (Filter) Encode(dst io.Writer, src io.Reader, _ string) error {
	data, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("runlength: read: %w", err)
	}
	encoded := encode(data)
	_, err = dst.Write(encoded)
	return err
}

func decode(data []byte) ([]byte, error) {
	var out bytes.Buffer
	i := 0
	for i < len(data) {
		n := int(data[i])
		i++
		if n == eodMarker {
			break
		}
		if err := decodeLiteral(&out, data, &i, n); err != nil {
			return nil, err
		}
	}
	return out.Bytes(), nil
}

func decodeLiteral(out *bytes.Buffer, data []byte, i *int, n int) error {
	if n < eodMarker {
		return copyLiteralRun(out, data, i, n+1)
	}
	return copyRepeatedRun(out, data, i, 257-n)
}

func copyLiteralRun(out *bytes.Buffer, data []byte, i *int, count int) error {
	if *i+count > len(data) {
		return fmt.Errorf("runlength: literal run extends past end of data")
	}
	out.Write(data[*i : *i+count])
	*i += count
	return nil
}

func copyRepeatedRun(out *bytes.Buffer, data []byte, i *int, count int) error {
	if *i >= len(data) {
		return fmt.Errorf("runlength: repeat run missing byte")
	}
	b := data[*i]
	*i++
	for range count {
		out.WriteByte(b)
	}
	return nil
}

func encode(data []byte) []byte {
	var out bytes.Buffer
	i := 0
	for i < len(data) {
		if canEncodeRepeatRun(data, i) {
			i = writeRepeatRun(&out, data, i)
		} else {
			i = writeLiteralRun(&out, data, i)
		}
	}
	out.WriteByte(eodMarker)
	return out.Bytes()
}

func canEncodeRepeatRun(data []byte, i int) bool {
	return i+1 < len(data) && data[i] == data[i+1]
}

func writeRepeatRun(out *bytes.Buffer, data []byte, i int) int {
	b := data[i]
	count := 1
	for count < 128 && i+count < len(data) && data[i+count] == b {
		count++
	}
	out.WriteByte(byte(257 - count))
	out.WriteByte(b)
	return i + count
}

func writeLiteralRun(out *bytes.Buffer, data []byte, i int) int {
	start := i
	for i < len(data) && i-start < 128 {
		if i+1 < len(data) && data[i] == data[i+1] {
			break
		}
		i++
	}
	count := i - start
	out.WriteByte(byte(count - 1))
	out.Write(data[start:i])
	return i
}
