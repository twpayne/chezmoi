package cmd

import (
	"io"
	"sync"
)

// A lazyWriter only opens its destination on first write.
type lazyWriter struct {
	mutex       sync.Mutex
	openFunc    func() (io.WriteCloser, error)
	writeCloser io.WriteCloser
	err         error
}

func newLazyWriter(openFunc func() (io.WriteCloser, error)) *lazyWriter {
	return &lazyWriter{
		openFunc: openFunc,
	}
}

func (w *lazyWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.writeCloser == nil {
		return nil
	}
	return w.writeCloser.Close()
}

func (w *lazyWriter) Write(p []byte) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.openFunc != nil {
		w.writeCloser, w.err = w.openFunc()
		w.openFunc = nil
	}
	if w.err != nil {
		return 0, w.err
	}
	return w.writeCloser.Write(p)
}
