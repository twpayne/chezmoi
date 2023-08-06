package truncate

import (
	"bytes"
	"io"

	"github.com/mattn/go-runewidth"

	"github.com/muesli/reflow/ansi"
)

type Writer struct {
	width uint
	tail  string

	ansiWriter *ansi.Writer
	buf        bytes.Buffer
	ansi       bool
}

func NewWriter(width uint, tail string) *Writer {
	w := &Writer{
		width: width,
		tail:  tail,
	}
	w.ansiWriter = &ansi.Writer{
		Forward: &w.buf,
	}
	return w
}

func NewWriterPipe(forward io.Writer, width uint, tail string) *Writer {
	return &Writer{
		width: width,
		tail:  tail,
		ansiWriter: &ansi.Writer{
			Forward: forward,
		},
	}
}

// Bytes is shorthand for declaring a new default truncate-writer instance,
// used to immediately truncate a byte slice.
func Bytes(b []byte, width uint) []byte {
	return BytesWithTail(b, width, []byte(""))
}

// Bytes is shorthand for declaring a new default truncate-writer instance,
// used to immediately truncate a byte slice. A tail is then added to the
// end of the byte slice.
func BytesWithTail(b []byte, width uint, tail []byte) []byte {
	f := NewWriter(width, string(tail))
	_, _ = f.Write(b)

	return f.Bytes()
}

// String is shorthand for declaring a new default truncate-writer instance,
// used to immediately truncate a string.
func String(s string, width uint) string {
	return StringWithTail(s, width, "")
}

// StringWithTail is shorthand for declaring a new default truncate-writer instance,
// used to immediately truncate a string. A tail is then added to the end of the
// string.
func StringWithTail(s string, width uint, tail string) string {
	return string(BytesWithTail([]byte(s), width, []byte(tail)))
}

// Write truncates content at the given printable cell width, leaving any
// ansi sequences intact.
func (w *Writer) Write(b []byte) (int, error) {
	tw := ansi.PrintableRuneWidth(w.tail)
	if w.width < uint(tw) {
		return w.buf.WriteString(w.tail)
	}

	w.width -= uint(tw)
	var curWidth uint

	for _, c := range string(b) {
		if c == ansi.Marker {
			// ANSI escape sequence
			w.ansi = true
		} else if w.ansi {
			if ansi.IsTerminator(c) {
				// ANSI sequence terminated
				w.ansi = false
			}
		} else {
			curWidth += uint(runewidth.RuneWidth(c))
		}

		if curWidth > w.width {
			n, err := w.buf.WriteString(w.tail)
			if w.ansiWriter.LastSequence() != "" {
				w.ansiWriter.ResetAnsi()
			}
			return n, err
		}

		_, err := w.ansiWriter.Write([]byte(string(c)))
		if err != nil {
			return 0, err
		}
	}

	return len(b), nil
}

// Bytes returns the truncated result as a byte slice.
func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}

// String returns the truncated result as a string.
func (w *Writer) String() string {
	return w.buf.String()
}
