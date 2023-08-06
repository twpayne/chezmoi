package gojq

import "unicode/utf8"

// Preview returns the preview string of v. The preview string is basically the
// same as the jq-flavored JSON encoding returned by [Marshal], but is truncated
// by 30 bytes, and more efficient than truncating the result of [Marshal].
//
// This method is used by error messages of built-in operators and functions,
// and accepts only limited types (nil, bool, int, float64, *big.Int, string,
// []any, and map[string]any). Note that the maximum width and trailing strings
// on truncation may be changed in the future.
func Preview(v any) string {
	bs := jsonLimitedMarshal(v, 32)
	if l := 30; len(bs) > l {
		var trailing string
		switch v.(type) {
		case string:
			trailing = ` ..."`
		case []any:
			trailing = " ...]"
		case map[string]any:
			trailing = " ...}"
		default:
			trailing = " ..."
		}
		for len(bs) > l-len(trailing) {
			_, size := utf8.DecodeLastRune(bs)
			bs = bs[:len(bs)-size]
		}
		bs = append(bs, trailing...)
	}
	return string(bs)
}

func jsonLimitedMarshal(v any, n int) (bs []byte) {
	w := &limitedWriter{buf: make([]byte, n)}
	defer func() {
		_ = recover()
		bs = w.Bytes()
	}()
	(&encoder{w: w}).encode(v)
	return
}

type limitedWriter struct {
	buf []byte
	off int
}

func (w *limitedWriter) Write(bs []byte) (int, error) {
	n := copy(w.buf[w.off:], bs)
	if w.off += n; w.off == len(w.buf) {
		panic(struct{}{})
	}
	return n, nil
}

func (w *limitedWriter) WriteByte(b byte) error {
	w.buf[w.off] = b
	if w.off++; w.off == len(w.buf) {
		panic(struct{}{})
	}
	return nil
}

func (w *limitedWriter) WriteString(s string) (int, error) {
	n := copy(w.buf[w.off:], s)
	if w.off += n; w.off == len(w.buf) {
		panic(struct{}{})
	}
	return n, nil
}

func (w *limitedWriter) Bytes() []byte {
	return w.buf[:w.off]
}
