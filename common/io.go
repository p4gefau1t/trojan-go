package common

import (
	"io"
	"net"

	"github.com/p4gefau1t/trojan-go/log"
)

type RewindReader struct {
	rawReader  io.Reader
	buf        []byte
	bufReadIdx int
	rewinded   bool
	buffering  bool
	bufferSize int
}

func (r *RewindReader) Read(p []byte) (int, error) {
	if r.rewinded {
		if len(r.buf) > r.bufReadIdx {
			n := copy(p, r.buf[r.bufReadIdx:])
			r.bufReadIdx += n
			return n, nil
		}
		r.rewinded = false //all buffering content has been read
	}
	n, err := r.rawReader.Read(p)
	if r.buffering {
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
	r.buffering = false
}

func (r *RewindReader) SetBufferSize(size int) {
	if size == 0 { //disable buffering
		if !r.buffering {
			panic("reader is already disabled")
		}
		r.buffering = false
		r.buf = nil
		r.bufReadIdx = 0
		r.bufferSize = 0
	} else {
		if r.buffering {
			panic("reader is already buffering")
		}
		r.buffering = true
		r.bufReadIdx = 0
		r.bufferSize = size
		r.buf = make([]byte, 0, size)
	}
}

type RewindConn struct {
	net.Conn
	*RewindReader
}

func (c *RewindConn) Read(p []byte) (int, error) {
	return c.RewindReader.Read(p)
}

func NewRewindConn(conn net.Conn) *RewindConn {
	return &RewindConn{
		Conn: conn,
		RewindReader: &RewindReader{
			rawReader: conn,
		},
	}
}

type StickyWriter struct {
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
