package main

import (
	"fmt"
	kcp "github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
	"idcproxy/protocols"
	"idcproxy/protocols/mika"
	"idcproxy/protocols/transfer/tcp"
	"idcproxy/utils"
	"net"
	"runtime"
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
			handle(utils.NewCompStream(tcpConn))
		}()
	}
}

func handleMux(conn protocols.Protocol) {
	// stream multiplex
	smuxConfig := smux.DefaultConfig()
	smuxConfig.MaxReceiveBuffer = 4194304
	smuxConfig.KeepAliveInterval = time.Duration(10) * time.Second

	mux, err := smux.Server(conn, smuxConfig)
	if err != nil {
		conn.Close()
		utils.Errorf("Create smux.Server error %s", err)
	}
	defer mux.Close()
	for {
		p1, err := mux.AcceptStream()
		if err != nil {
			utils.Errorf("mux.AcceptStream error %s", err)
			return
		}
		go handle(p1)
	}
}

func listenKcp(serverInfo *utils.ServerConf) {
	nl, err := kcp.ListenWithOptions(fmt.Sprintf("%s:%d", serverInfo.Address, serverInfo.Port), nil, 10, 3)
	if err != nil {
		utils.Fatalf("Create server error %s", err)
	}
	if err := nl.SetDSCP(0); err != nil {
		utils.Fatalf("SetDSCP error %s", err)
	}
	if err := nl.SetReadBuffer(4194304); err != nil {
		utils.Fatalf("SetReadBuffer error %s", err)
	}
	if err := nl.SetWriteBuffer(4194304); err != nil {
		utils.Fatalf("SetWriteBuffer error %s", err)
	}

	utils.Infof("Listen on kcp://%s:%d\n", serverInfo.Address, serverInfo.Port)
	for {
		conn, err := nl.AcceptKCP()
		if err != nil {
			utils.Errorf("Accept connection error %s", err)
			continue
		}

		go func() {
			conn.SetStreamMode(true)
			conn.SetNoDelay(1, 10, 2, 1)
			conn.SetACKNoDelay(true)
			conn.SetWindowSize(1024, 1024)
			conn.SetWriteDelay(true)
			conn.SetMtu(1350)

			kcpConn := &tcp.Conn{conn, time.Duration(serverInfo.Timeout) * time.Second}
			handleMux(utils.NewCompStream(kcpConn))
		}()
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	conf = utils.ParseSeverConf()
	//TODO check conf

	//注册服务
	for _, serverInfo := range conf.Server {
		if serverInfo.Timeout <= 0 {
			serverInfo.Timeout = 30
		}

		if serverInfo.Protocol == "kcp" {
			listenKcp(serverInfo)
		} else {
			listen(serverInfo)
		}
	}
}
