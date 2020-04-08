package message

import (
	"fmt"
	"net"
	sync "sync"
	"sync/atomic"
	"time"
)

// MsgStreamPool provides pool of MsgStream to requset and response message.
// After fetching a MsgStream from the pool, it need to be returned by CloseMsgStream method
// after finished using it. If there is some problem(connection error, read error..) in the MsgStream,
// it need to be destroyed by DestroyMsgSteram method so that prevent MsgStreamPool wasting it's max capacity and providing stale stream.
// MsgStreamPool periodically check whether idle stream is stale or not by sending predefined healthcheck message to it's connection.
// This periodical stale check is to minmize MsgStreamPool providing stale stream to user.
type MsgStreamPool struct {
	pool                         chan *MsgStream //channel where idle connection resides
	mutex                        sync.Mutex      //mutex for synchronization between getMsgStream and removal of stale connection
	size                         int32           //total number of connection in pool
	cap                          int32           //maximum number of connection
	idle                         int32           //number of idle connection which means not in use and reside in the channel
	connType, connHost, connPort string          //connection info
}

// Create new message stream
func NewMsgStreamPool(connType, connHost, connPort string, cap int32) *MsgStreamPool {
	pool := &MsgStreamPool{
		pool:     make(chan *MsgStream, cap),
		size:     0,
		cap:      cap,
		idle:     0,
		connType: connType,
		connHost: connHost,
		connPort: connPort,
	}
	go pool.checkStale()
	return pool
}

// give back message stream to stream pool
func (msp *MsgStreamPool) closeMsgStream(stream *MsgStream) {
	msp.pool <- stream
	atomic.AddInt32(&msp.idle, 1)
}

// remove message stream from stream pool
func (msp *MsgStreamPool) destroyMsgStream(stream *MsgStream) {
	stream.Close()
	atomic.AddInt32(&msp.size, -1)
}

// add new message stream to stream pool
func (msp *MsgStreamPool) pushMsgStream(stream *MsgStream) {
	msp.pool <- stream
	atomic.AddInt32(&msp.size, 1)
	atomic.AddInt32(&msp.idle, 1)
}

// Get a msgStream from the pool, if there is idle one, return it.
// if all stream are being used and there is space for new one, create new one and return.
// if there is no idel stream and space, it just wait for a stream to be idle
func (msp *MsgStreamPool) GetMsgStream() (*MsgStream, error) {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()
	// always try to reuse one we already have
	if msp.idle == 0 && msp.size < msp.cap {
		conn, err := net.Dial(msp.connType, net.JoinHostPort(msp.connHost, msp.connPort))
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return nil, err
		}
		stream, err := NewMsgStream(conn, 60)
		if err != nil {
			return nil, err
		}
		msp.pushMsgStream(stream)
	}
	stream := <-msp.pool
	atomic.AddInt32(&msp.idle, -1)
	return stream, nil
}

// periodically remove stale connections from pool.
func (msp *MsgStreamPool) checkStale() {
	// check every stale connections every 20 seconds
	ticker := time.NewTicker(time.Second * 20)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			msp.removeStaleStreams()
		}
	}
}

//remove stale stream from pool, only check idle streams.
func (msp *MsgStreamPool) removeStaleStreams() {
	msp.mutex.Lock()
	idleCnt := msp.idle
	streams := make([]*MsgStream, idleCnt)
	//take all idle stream
	for i := int32(0); i < idleCnt; i++ {
		streams[i] = <-msp.pool
		atomic.AddInt32(&msp.idle, -1)
	}
	msp.mutex.Unlock()
	removed := 0
	fmt.Println("idle connections : ", msp.idle, msp.size)
	for _, stream := range streams {
		//check if stale, remove stale connection, put back others
		if isStaleStream(stream) {
			msp.destroyMsgStream(stream)
			removed++
		} else {
			msp.closeMsgStream(stream)
		}
	}
	fmt.Printf("removed %d stale streams.\n", removed)
}

// send predefined healtcheck msg to check health of connection
func isStaleStream(s *MsgStream) bool {
	err := s.WriteMsg(&HealthcheckMessage{})
	if err != nil {
		return true
	}
	_, err = s.ReadMsg()
	if err != nil {
		return true
	}
	return false
}
