package mika

import (
	"idcproxy/protocols"
	"idcproxy/utils"
	"net"
)

func Serve(c *Mika) {
	defer c.Close()

	utils.Infof("Connection from %s", c.RemoteAddr())

	address := c.header.Addr

	utils.Infof("Connect to %s", address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		utils.Errorf("Create connection error %s", err)
		return
	}

	go protocols.Pipe(conn, c)
	protocols.Pipe(c, conn)
}
