//go:build linux && 386
// +build linux,386

package tproxy

import (
	"syscall"
	"unsafe"
)

const GETSOCKOPT = 15

func getsockopt(fd int, level int, optname int, optval unsafe.Pointer, optlen *uint32) (err error) {
	_, _, e := syscall.Syscall6(
		GETSOCKOPT, uintptr(fd), uintptr(level), uintptr(optname),
		uintptr(optval), uintptr(unsafe.Pointer(optlen)), 0)
	if e != 0 {
		return e
	}
	return
}
