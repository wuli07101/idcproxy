package mika

import (
	"idcproxy/protocols"
	"net"
)

// Mika dails connection between mika server and mika client.
type Mika struct {
	*Conn
	header     *header
	serverSide bool
}

// NewMika wraps a new Mika connection.
// Notice, if header is nil, Mika coonection would be on server side otherwise client side.
func NewMika(conn protocols.Protocol, header *header) (*Mika, error) {
	ss := &Conn{
		Conn: conn,
	}

	mika := &Mika{
		Conn:       ss,
		header:     header,
		serverSide: header == nil,
	}

	if mika.serverSide {
		// On server side, we should get header first.
		header, err := getHeader(ss)
		if err != nil {
			return nil, err
		}
		mika.header = header
	} else {
		// On client side, send header as quickly.
		data := header.Bytes()
		ss.Write(data)
	}

	return mika, nil
}

// Close closes connection and releases buf.
// TODO check close state to avoid close twice.
func (c *Mika) Close() error {
	return c.Conn.Close()
}

// Write writes data to connection.
func (c *Mika) Write(b []byte) (n int, err error) {
	return c.Conn.Write(b)
}

func (c *Mika) Read(b []byte) (n int, err error) {
	return c.Conn.Read(b)
}

func DailWithRawAddrHTTP(network string, server string, rawAddr []byte) (protocols.Protocol, error) {
	var conn net.Conn
	var err error
	conn, err = net.Dial(network, server)
	if err != nil {
		return nil, err
	}

	header := newHeader(rawAddr)
	return NewMika(conn, header)
}
