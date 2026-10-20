package utf8x

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"testing/iotest"
)

func TestValid(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		want bool
	}{
		{"empty", nil, true},
		{"ascii", []byte("sample"), true},
		{"multibyte", []byte("é€😀"), true},
		{"lone continuation", []byte{'a', 0x80}, false},
		{"overlong", []byte{0xC0, 0xAF}, false},
		{"surrogate", []byte{0xED, 0xA0, 0x80}, false},
		{"cut at end", []byte{'a', 0xE2, 0x82}, false},
		{"bad byte", []byte{0xFF}, false},
	}
	for _, c := range cases {
		for _, rd := range []struct {
			name string
			r    io.Reader
		}{
			{"whole", bytes.NewReader(c.in)},
			{"one byte", iotest.OneByteReader(bytes.NewReader(c.in))},
			{"half", iotest.HalfReader(bytes.NewReader(c.in))},
		} {
			got, err := Valid(rd.r)
			if err != nil || got != c.want {
				t.Errorf("%s/%s: got %v, %v; want %v", c.name, rd.name, got, err, c.want)
			}
		}
	}
}

func TestReadError(t *testing.T) {
	boom := errors.New("boom")
	if _, err := Valid(iotest.ErrReader(boom)); !errors.Is(err, boom) {
		t.Fatalf("got %v, want the read error", err)
	}
}

func TestReaderPassesThrough(t *testing.T) {
	in := bytes.Repeat([]byte("é😀a"), 5000)
	got, err := io.ReadAll(NewReader(iotest.HalfReader(bytes.NewReader(in))))
	if err != nil || !bytes.Equal(got, in) {
		t.Fatalf("got %d bytes, %v", len(got), err)
	}
}
