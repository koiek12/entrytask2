package message

import (
	"fmt"
	"net"
	sync "sync"
	"sync/atomic"
	"time"
)

type MsgStreamPool struct {
	pool                         chan *MsgStream
	mutex                        sync.Mutex
	size                         int32
	cap                          int32
	active, idle                 int32
	connType, connHost, connPort string
}

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

// pop a message stream from stream pool for usage
func (msp *MsgStreamPool) popMsgStream() *MsgStream {
	stream := <-msp.pool
	atomic.AddInt32(&msp.idle, -1)
	return stream
}

func (msp *MsgStreamPool) GetMsgStream() (*MsgStream, error) {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()
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
	return msp.popMsgStream(), nil
}

func (msp *MsgStreamPool) checkStale() {
	ticker := time.NewTicker(time.Second * 20)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			msp.removeStaleStreams()
		}
	}
}

func (msp *MsgStreamPool) removeStaleStreams() {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()
	removed := 0
	fmt.Println("idle connections : ", msp.idle, msp.size)
	idleCnt := msp.idle
	for i := int32(0); i < idleCnt; i++ {
		stream := msp.popMsgStream()
		if isStaleStream(stream) {
			msp.destroyMsgStream(stream)
			removed++
		} else {
			msp.closeMsgStream(stream)
		}
	}
	fmt.Printf("removed %d stale streams.\n", removed)
}

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
