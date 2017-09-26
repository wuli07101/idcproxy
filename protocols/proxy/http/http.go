package http

import (
	"bufio"
	"idcproxy/protocols"
	"idcproxy/protocols/mika"
	"idcproxy/utils"
	"net/http"
	"net/http/httputil"
)

type Relay struct {
	conn   protocols.Protocol
	closed bool
}

func NewRelay(conn protocols.Protocol) *Relay {
	return &Relay{
		conn: conn,
	}
}

// Serve parse data and then send to idcproxy server.
func (h *Relay) Serve() {
	client := h.conn
	bf := bufio.NewReader(h.conn)
	req, err := http.ReadRequest(bf)
	if err != nil {
		utils.Errorf("Read request error %s", err)
		return
	}

	idcName := req.Header.Get("idcName")
	mikaConn, err := mika.DailWithRawAddrHTTP(idcName, "kcp", utils.ToAddr(req.URL.Host))
	if err != nil {
		utils.Errorf("Close connection error %v\n", err)
		return
	}

	defer func() {
		if !h.closed {
			err := mikaConn.Close()
			utils.Errorf("Close connection error %v\n", err)
		}
	}()

	rmProxyHeaders(req)

	if req.Method == "CONNECT" {
		client.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	} else {
		dump, _ := httputil.DumpRequestOut(req, true)
		mikaConn.Write(dump)
	}

	p1die := make(chan struct{})
	go func() { protocols.Pipe(client, mikaConn); close(p1die) }()

	p2die := make(chan struct{})
	go func() { protocols.Pipe(mikaConn, client); close(p2die) }()

	// wait for tunnel termination
	select {
	case <-p1die:
	case <-p2die:
	}

	h.closed = true
}

// rmProxyHeaders remove Hop-by-hop headers.
func rmProxyHeaders(req *http.Request) {
	req.RequestURI = ""
	req.Header.Del("Proxy-Connection")
	req.Header.Del("Connection")
	req.Header.Del("Keep-Alive")
	req.Header.Del("Proxy-Authenticate")
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("TE")
	req.Header.Del("Trailers")
	req.Header.Del("Transfer-Encoding")
	req.Header.Del("Upgrade")
}
