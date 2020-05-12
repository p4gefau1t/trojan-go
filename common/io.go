package common

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/golang/snappy"
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
			panic("already disabled")
		}
		r.buffered = false
		r.buf = nil
		r.bufReadIdx = 0
		r.bufferSize = 0
	} else {
		if r.buffered {
			panic("is already buffering")
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
	R *RewindReader
	net.Conn
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

const (
	compressionThreshold = 512
)

type compFrame struct {
	Algorithm      byte
	CompressedSize int
	Compressed     []byte

	OriginalSize int
	Original     []byte
}

func (h *compFrame) ReadFrom(r io.Reader) (int64, error) {
	buf := [3]byte{}
	n, err := r.Read(buf[:])
	if err != nil {
		return 0, err
	}
	if n != 3 {
		return 0, NewError("Too short")
	}
	h.Algorithm = buf[0]
	if h.Algorithm == 0 {
		h.OriginalSize = int(binary.LittleEndian.Uint16(buf[1:]))
		h.Original = make([]byte, h.OriginalSize)
		n, err := r.Read(h.Original)
		return int64(n), err
	}
	h.CompressedSize = int(binary.LittleEndian.Uint16(buf[1:]))
	h.Compressed = make([]byte, h.CompressedSize)
	if err != nil {
		return 0, err
	}
	n, err = r.Read(h.Compressed)
	if err != nil {
		return 0, err
	}
	if n != h.CompressedSize {
		return 0, NewError("Too short")
	}
	h.Original, err = snappy.Decode(nil, h.Compressed)
	if err != nil {
		return 0, err
	}
	h.OriginalSize = len(h.Original)
	log.Trace("Compression ratio", float32(h.CompressedSize)/float32(h.OriginalSize))
	return int64(h.OriginalSize), nil
}

func (h *compFrame) WriteTo(w io.Writer) (int64, error) {
	h.OriginalSize = len(h.Original)
	buf := [3]byte{}
	if h.OriginalSize < compressionThreshold {
		buf[0] = 0
		binary.LittleEndian.PutUint16(buf[1:], uint16(h.OriginalSize))
		n, err := w.Write(append(buf[:], h.Original...))
		n -= 3
		if n < 0 {
			n = 0
		}
		return int64(n), err
	}
	// TODO
	if h.CompressedSize > 0xffff {
		panic("todo")
	}
	buf[0] = 1
	h.Compressed = snappy.Encode(nil, h.Original)
	h.CompressedSize = len(h.Compressed)
	binary.LittleEndian.PutUint16(buf[1:], uint16(h.CompressedSize))
	n, err := w.Write(append(buf[:], h.Compressed...))
	n -= 3
	if n < 0 {
		n = 0
	}
	log.Trace("Compression ratio", float32(h.CompressedSize)/float32(h.OriginalSize))
	return int64(n), err
}

type CompReader struct {
	rawReader io.Reader
	frame     *compFrame
	readBytes int
}

func (r *CompReader) readFrame(p []byte) (int, error) {
	n := copy(p, r.frame.Original[r.readBytes:])
	r.readBytes += n
	if r.readBytes >= r.frame.OriginalSize {
		//buffer is bigger than the frame
		r.frame = nil
		r.readBytes = 0
	}
	return n, nil
}

func (r *CompReader) Read(p []byte) (int, error) {
	if r.frame != nil {
		return r.readFrame(p)
	}
	r.frame = new(compFrame)
	_, err := r.frame.ReadFrom(r.rawReader)
	if err != nil {
		return 0, err
	}
	return r.readFrame(p)
}

type CompWriter struct {
	rawWriter io.Writer
}

func (w *CompWriter) Write(p []byte) (int, error) {
	frame := &compFrame{
		OriginalSize: len(p),
		Original:     p,
	}
	_, err := frame.WriteTo(w.rawWriter)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

type CompReadWriteCloser struct {
	*CompReader
	*CompWriter
	rawRWC io.ReadWriteCloser
}

func (rwc *CompReadWriteCloser) Close() error {
	return rwc.rawRWC.Close()
}

func NewCompReadWriteCloser(rwc io.ReadWriteCloser) (io.ReadWriteCloser, error) {
	return &CompReadWriteCloser{
		CompReader: &CompReader{
			rawReader: rwc,
		},
		CompWriter: &CompWriter{
			rawWriter: rwc,
		},
		rawRWC: rwc,
	}, nil
}
