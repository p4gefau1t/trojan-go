package common

import (
	"io"
	"net"

	"github.com/p4gefau1t/trojan-go/log"
)

type RewindReader struct {
	io.Reader
	io.ByteReader

	rawReader  io.Reader
	buf        []byte
	bufReadIdx int
	rewinded   bool
	buffered   bool
	bufferSize int
}

func (r *RewindReader) Read(p []byte) (int, error) {
	if r.rewinded {
		if len(r.buf) > r.bufReadIdx {
			n := copy(p, r.buf[r.bufReadIdx:])
			r.bufReadIdx += n
			return n, nil
		}
		r.rewinded = false //all buffered content has been read
	}
	n, err := r.rawReader.Read(p)
	if r.buffered {
		r.buf = append(r.buf, p[:n]...)
		if len(r.buf) > r.bufferSize*2 {
			log.Debug("read too many bytes!")
		}
	}
	return n, err
}

func (r *RewindReader) ReadByte() (byte, error) {
	buf := [1]byte{}
	_, err := r.Read(buf[:])
	return buf[0], err
}

func (r *RewindReader) Discard(n int) (int, error) {
	buf := [128]byte{}
	if n < 128 {
		return r.Read(buf[:n])
	}
	for discarded := 0; discarded+128 < n; discarded += 128 {
		_, err := r.Read(buf[:])
		if err != nil {
			return discarded, err
		}
	}
	if rest := n % 128; rest != 0 {
		return r.Read(buf[:rest])
	}
	return n, nil
}

func (r *RewindReader) Rewind() {
	if r.bufferSize == 0 {
		panic("has no buffer")
	}
	r.rewinded = true
	r.bufReadIdx = 0
}

func (r *RewindReader) StopBuffering() {
	r.buffered = false
}

func (r *RewindReader) SetBufferSize(size int) {
	if size == 0 { //disable buffering
		if !r.buffered {
			panic("reader is already disabled")
		}
		r.buffered = false
		r.buf = nil
		r.bufReadIdx = 0
		r.bufferSize = 0
	} else {
		if r.buffered {
			panic("reader is already buffering")
		}
		r.buffered = true
		r.bufReadIdx = 0
		r.bufferSize = size
		r.buf = make([]byte, 0, size)
	}
}

func NewRewindReader(r io.Reader) *RewindReader {
	return &RewindReader{
		rawReader: r,
	}
}

type RewindReadWriteCloser struct {
	rawRWC io.ReadWriteCloser
	*RewindReader
}

func (rwc *RewindReadWriteCloser) Write(p []byte) (int, error) {
	return rwc.rawRWC.Write(p)
}

func (rwc *RewindReadWriteCloser) Close() error {
	return rwc.rawRWC.Close()
}

func NewRewindReadWriteCloser(rwc io.ReadWriteCloser) *RewindReadWriteCloser {
	return &RewindReadWriteCloser{
		rawRWC:       rwc,
		RewindReader: NewRewindReader(rwc),
	}
}

func ReadByte(r io.Reader) (byte, error) {
	buf := [1]byte{}
	_, err := r.Read(buf[:])
	return buf[0], err
}

type RewindConn struct {
	net.Conn
	R *RewindReader
}

func (c *RewindConn) Read(p []byte) (int, error) {
	return c.R.Read(p)
}

func NewRewindConn(conn net.Conn) *RewindConn {
	return &RewindConn{
		Conn: conn,
		R:    NewRewindReader(conn),
	}
}

type StickyWriter struct {
	io.Writer

	rawWriter   io.Writer
	writeBuffer []byte
	MaxBuffered int
}

func (w *StickyWriter) Write(p []byte) (int, error) {
	if w.MaxBuffered > 0 {
		w.MaxBuffered--
		w.writeBuffer = append(w.writeBuffer, p...)
		if w.MaxBuffered != 0 {
			return len(p), nil
		}
		w.MaxBuffered = 0
		_, err := w.rawWriter.Write(w.writeBuffer)
		w.writeBuffer = nil
		return len(p), err
	}
	return w.rawWriter.Write(p)
}
