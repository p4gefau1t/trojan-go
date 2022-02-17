//go:build linux
// +build linux

package tproxy

import (
	"fmt"
	"net"
	"syscall"
)

// Listener describes a TCP Listener
// with the Linux IP_TRANSPARENT option defined
// on the listening socket
type Listener struct {
	base net.Listener
}

// Accept waits for and returns
// the next connection to the listener.
//
// This command wraps the AcceptTProxy
// method of the Listener
func (listener *Listener) Accept() (net.Conn, error) {
	tcpConn, err := listener.base.(*net.TCPListener).AcceptTCP()
	if err != nil {
		return nil, err
	}

	return tcpConn, nil
}

// Addr returns the network address
// the listener is accepting connections
// from
func (listener *Listener) Addr() net.Addr {
	return listener.base.Addr()
}

// Close will close the listener from accepting
// any more connections. Any blocked connections
// will unblock and close
func (listener *Listener) Close() error {
	return listener.base.Close()
}

// ListenTCP will construct a new TCP listener
// socket with the Linux IP_TRANSPARENT option
// set on the underlying socket
func ListenTCP(network string, laddr *net.TCPAddr) (net.Listener, error) {
	listener, err := net.ListenTCP(network, laddr)
	if err != nil {
		return nil, err
	}

	fileDescriptorSource, err := listener.File()
	if err != nil {
		return nil, &net.OpError{Op: "listen", Net: network, Source: nil, Addr: laddr, Err: fmt.Errorf("get file descriptor: %s", err)}
	}
	defer fileDescriptorSource.Close()

	if err = syscall.SetsockoptInt(int(fileDescriptorSource.Fd()), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1); err != nil {
		return nil, &net.OpError{Op: "listen", Net: network, Source: nil, Addr: laddr, Err: fmt.Errorf("set socket option: IP_TRANSPARENT: %s", err)}
	}

	return &Listener{listener}, nil
}

const (
	IP6T_SO_ORIGINAL_DST = 80
	SO_ORIGINAL_DST      = 80
)

// getOriginalTCPDest retrieves the original destination address from
// NATed connection.  Currently, only Linux iptables using DNAT/REDIRECT
// is supported.  For other operating systems, this will just return
// conn.LocalAddr().
//
// Note that this function only works when nf_conntrack_ipv4 and/or
// nf_conntrack_ipv6 is loaded in the kernel.
func getOriginalTCPDest(conn *net.TCPConn) (*net.TCPAddr, error) {
	return conn.LocalAddr().(*net.TCPAddr), nil
}
