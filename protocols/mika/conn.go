package mika

import (
	"idcproxy/protocols"
	"net"
)

// Conn dails connection between ss server and ss client.
type Conn struct {
	Conn protocols.Protocol
}

// NewConn creates a new shadowsocks connection.
func NewConn(conn protocols.Protocol) protocols.Protocol {
	return &Conn{
		Conn: conn,
	}
}

// RemoteAddr gets remote connection address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// Close closes connection and releases buf.
// TODO check close state to avoid close twice.
func (c *Conn) Close() error {
	return c.Conn.Close()
}

// Write writes data to connection.
func (c *Conn) write(b []byte) (n int, err error) {
	return c.Conn.Write(b)
}

// Write writes data to connection.
func (c *Conn) Write(b []byte) (n int, err error) {
	return c.Conn.Write(b)
}

// Read reads data from connection.
func (c *Conn) Read(b []byte) (n int, err error) {
	return c.Conn.Read(b)
}
