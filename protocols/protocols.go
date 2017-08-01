// Package protocols defines protocol interface.
package protocols

import (
	"io"
	"net"
)

type Protocol interface {
	Write(b []byte) (n int, err error)
	Read(b []byte) (n int, err error)
	RemoteAddr() net.Addr
	Close() error
}

// // Pipe pipes two protocols which should be in same layer.
func Pipe(dst, src Protocol) {
	io.Copy(dst, src)
}
