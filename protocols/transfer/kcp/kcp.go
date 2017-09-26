package kcp

import (
	"errors"
	cmap "github.com/streamrail/concurrent-map"
	kcp "github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
	"idcproxy/utils"
	"math/rand"
	"sync"
	"time"
)

var (
	idcObjects               = cmap.New()
	lock       *sync.RWMutex = new(sync.RWMutex)
)

type smuxObj struct {
	session    *smux.Session
	ttl        time.Time
	serverAddr string
}

type idcObject struct {
	idcName string
	muxes   []smuxObj
}

func ResetIdcConn(idcName string, server []string, conn int) (bool, error) {
	lock.RLock()
	defer lock.RUnlock()

	serverLen := len(server)
	numconn := uint16(conn * serverLen)
	muxes := make([]smuxObj, numconn)

	k := 0
	for i := 0; i < serverLen; i++ {
		for j := 0; j < conn; j++ {
			muxes[k].session = waitKcpConn(server[i])
			muxes[k].ttl = time.Now().Add(time.Duration(30) * time.Second)
			muxes[k].serverAddr = server[i]

			k++
		}
	}

	// for k := range muxes {
	// 	muxes[k].session = waitKcpConn(server)
	// 	muxes[k].ttl = time.Now().Add(time.Duration(30) * time.Second)
	// 	muxes[k].serverAddr = server
	// }

	obj := &idcObject{
		idcName: idcName,
		muxes:   muxes,
	}

	addIdcObject(obj)

	return true, nil
}

// wait until a connection is ready
func waitKcpConn(server string) *smux.Session {
	for {
		if session, err := createKcpConn(server); err == nil {
			return session
		} else {
			time.Sleep(time.Second)
		}
	}
}

func createKcpConn(server string) (*smux.Session, error) {
	kcpconn, err := kcp.DialWithOptions(server, nil, 10, 3)
	if err != nil {
		return nil, err
	}
	kcpconn.SetStreamMode(true)
	kcpconn.SetWriteDelay(true)
	kcpconn.SetNoDelay(1, 10, 2, 1)
	kcpconn.SetWindowSize(128, 512)
	kcpconn.SetMtu(1350)
	kcpconn.SetACKNoDelay(true)

	if err := kcpconn.SetDSCP(0); err != nil {
		return nil, err
	}
	if err := kcpconn.SetReadBuffer(4194304); err != nil {
		return nil, err
	}
	if err := kcpconn.SetWriteBuffer(4194304); err != nil {
		return nil, err
	}

	smuxConfig := smux.DefaultConfig()
	smuxConfig.MaxReceiveBuffer = 4194304
	smuxConfig.KeepAliveInterval = time.Duration(10) * time.Second

	// stream multiplex
	var session *smux.Session
	session, err = smux.Client(utils.NewCompStream(kcpconn), smuxConfig)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func findIdcObject(idcName string) (*idcObject, bool) {
	if idcName == "" {
		return nil, false
	}
	if v, found := idcObjects.Get(idcName); found {
		return v.(*idcObject), true
	}
	return nil, false
}

func addIdcObject(obj *idcObject) {
	idcObjects.Set(obj.idcName, obj)
}

func removeIdcObject(idcName string) {
	idcObjects.Remove(idcName)
}

func GetKcpSmuxSession(idcName string) (*smux.Session, error) {
	if idcObject, result := findIdcObject(idcName); result == true {
		idx := rand.Intn(len(idcObject.muxes))
		if idcObject.muxes[idx].session.IsClosed() || time.Now().After(idcObject.muxes[idx].ttl) {
			idcObject.muxes[idx].session = waitKcpConn(idcObject.muxes[idx].serverAddr)
			idcObject.muxes[idx].ttl = time.Now().Add(time.Duration(10) * time.Second)
		}
		return idcObject.muxes[idx].session, nil
	}

	var err error = errors.New("idc not exist")

	return nil, err
}
