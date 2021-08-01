package lib

import "io"

type NoopWriteCloser struct {
	W io.Writer
}

func (n *NoopWriteCloser) Write(p []byte) (int, error) {
	return n.W.Write(p)
}

func (n *NoopWriteCloser) Close() error {
	return nil
}
