// +build linux darwin

package tproxy

import (
	"net"
	"os"
	"unsafe"

	syscall "golang.org/x/sys/unix"
)

const (
	IP6T_SO_ORIGINAL_DST = 80
	SO_ORIGINAL_DST      = 80
)

func getsockopt(s int, level int, optname int, optval unsafe.Pointer, optlen *uint32) (err error) {
	_, _, e := syscall.Syscall6(
		syscall.SYS_GETSOCKOPT, uintptr(s), uintptr(level), uintptr(optname),
		uintptr(optval), uintptr(unsafe.Pointer(optlen)), 0)
	if e != 0 {
		return e
	}
	return
}

// getOriginalTCPDest retrieves the original destination address from
// NATed connection.  Currently, only Linux iptables using DNAT/REDIRECT
// is supported.  For other operating systems, this will just return
// conn.LocalAddr().
//
// Note that this function only works when nf_conntrack_ipv4 and/or
// nf_conntrack_ipv6 is loaded in the kernel.
func getOriginalTCPDest(conn *net.TCPConn) (*net.TCPAddr, error) {
	f, err := conn.File()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fd := int(f.Fd())
	// revert to non-blocking mode.
	// see http://stackoverflow.com/a/28968431/1493661
	if err = syscall.SetNonblock(fd, true); err != nil {
		return nil, os.NewSyscallError("setnonblock", err)
	}

	v6 := conn.LocalAddr().(*net.TCPAddr).IP.To4() == nil
	if v6 {
		var addr syscall.RawSockaddrInet6
		var len uint32
		len = uint32(unsafe.Sizeof(addr))
		err = getsockopt(fd, syscall.IPPROTO_IPV6, IP6T_SO_ORIGINAL_DST,
			unsafe.Pointer(&addr), &len)
		if err != nil {
			return nil, os.NewSyscallError("getsockopt", err)
		}
		ip := make([]byte, 16)
		for i, b := range addr.Addr {
			ip[i] = b
		}
		pb := *(*[2]byte)(unsafe.Pointer(&addr.Port))
		return &net.TCPAddr{
			IP:   ip,
			Port: int(pb[0])*256 + int(pb[1]),
		}, nil
	}

	// IPv4
	var addr syscall.RawSockaddrInet4
	var len uint32
	len = uint32(unsafe.Sizeof(addr))
	err = getsockopt(fd, syscall.IPPROTO_IP, SO_ORIGINAL_DST,
		unsafe.Pointer(&addr), &len)
	if err != nil {
		return nil, os.NewSyscallError("getsockopt", err)
	}
	ip := make([]byte, 4)
	for i, b := range addr.Addr {
		ip[i] = b
	}
	pb := *(*[2]byte)(unsafe.Pointer(&addr.Port))
	return &net.TCPAddr{
		IP:   ip,
		Port: int(pb[0])*256 + int(pb[1]),
	}, nil
}
