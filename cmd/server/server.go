package main

import (
	"fmt"
	"idcproxy/protocols"
	"idcproxy/protocols/mika"
	"idcproxy/protocols/transfer/tcp"
	"idcproxy/utils"
	"net"
	"time"
)

var conf *utils.Conf

func handle(c protocols.Protocol) {
	mikaConn, err := mika.NewMika(c, nil)
	if err != nil {
		c.Close()
		utils.Errorf("Create mika connection error %s", err)
		return
	}
	mika.Serve(mikaConn)
}

func listen(serverInfo *utils.ServerConf) {
	nl, err := net.Listen("tcp", fmt.Sprintf("%s:%d", serverInfo.Address, serverInfo.Port))
	if err != nil {
		utils.Fatalf("Create server error %s", err)
	}

	utils.Infof("Listen on tcp://%s:%d\n", serverInfo.Address, serverInfo.Port)

	for {
		c, err := nl.Accept()
		if err != nil {
			utils.Errorf("Accept connection error %s", err)
			continue
		}

		go func() {
			tcpConn := &tcp.Conn{c, time.Duration(serverInfo.Timeout) * time.Second}
			handle(tcpConn)
		}()
	}
}

// func listenKcp(serverInfo *utils.ServerConf) {
// 	nl, err := kcp.ListenWithOptions(fmt.Sprintf("%s:%d", serverInfo.Address, serverInfo.Port), nil, 10, 3)
// 	if err != nil {
// 		utils.Fatalf("Create server error %s", err)
// 	}

// 	utils.Infof("Listen on kcp://%s:%d\n", serverInfo.Address, serverInfo.Port)
// 	cg := mika.NewCryptoGenerator(serverInfo.Method, serverInfo.Password)

// 	for {
// 		conn, err := nl.AcceptKCP()
// 		if err != nil {
// 			utils.Errorf("Accept connection error %s", err)
// 			continue
// 		}

// 		go func() {
// 			conn.SetStreamMode(true)
// 			conn.SetNoDelay(1, 20, 2, 1)
// 			conn.SetACKNoDelay(true)
// 			conn.SetWindowSize(1024, 1024)
// 			kcpConn := &tcp.Conn{conn, time.Duration(serverInfo.Timeout) * time.Second}
// 			handle(kcpConn, cg)
// 		}()
// 	}
// }

func main() {
	conf = utils.ParseSeverConf()

	//TODO check conf
	for _, serverInfo := range conf.Server {
		if serverInfo.Timeout <= 0 {
			serverInfo.Timeout = 30
		}

		listen(serverInfo)
	}
}
