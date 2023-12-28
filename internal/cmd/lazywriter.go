package cmd

import "io"

// A lazyWriter only opens its destination on first write.
type lazyWriter struct {
	openFunc    func() (io.WriteCloser, error)
	writeCloser io.WriteCloser
}

func newLazyWriter(openFunc func() (io.WriteCloser, error)) *lazyWriter {
	return &lazyWriter{
		openFunc: openFunc,
	}
}

func (w *lazyWriter) Close() error {
	if w.writeCloser == nil {
		return nil
	}
	return w.writeCloser.Close()
}

func (w *lazyWriter) Write(p []byte) (int, error) {
	if w.writeCloser == nil {
		writeCloser, err := w.openFunc()
		w.openFunc = nil
		if err != nil {
			return 0, err
		}
		w.writeCloser = writeCloser
	}
	return w.writeCloser.Write(p)
}
