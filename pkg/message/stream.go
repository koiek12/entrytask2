package message

import (
	"bufio"
	"encoding/binary"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
)

// MsgStream abstract stream where messages for communication come and go.
type MsgStream struct {
	conn net.Conn
	in   *bufio.Reader
	out  *bufio.Writer
	tmp  []byte
}

func NewMsgStream(conn net.Conn, timeout time.Duration) (*MsgStream, error) {
	conn.SetReadDeadline(time.Now().Add(time.Second * timeout))
	return &MsgStream{conn, bufio.NewReader(conn), bufio.NewWriter(conn), make([]byte, 32)}, nil
}

func (ms *MsgStream) Close() {
	ms.conn.Close()
}

func (ms *MsgStream) ReadVint() (uint, error) {
	val, err := binary.ReadUvarint(ms.in)
	return uint(val), err
}

func (ms *MsgStream) WriteVint(val uint) error {
	n := binary.PutUvarint(ms.tmp, uint64(val))
	n, err := ms.out.Write(ms.tmp[:n])
	return err
}

func (ms *MsgStream) readLenDelimData() ([]byte, error) {
	size, err := binary.ReadUvarint(ms.in)
	if err != nil {
		return nil, err
	}
	data := make([]byte, size)
	ptr := uint64(0)
	var n int
	for ptr < size {
		n, err = ms.in.Read(data[ptr:len(data)])
		if err != nil {
			return nil, err
		}
		ptr += uint64(n)
	}
	return data, nil
}

func (ms *MsgStream) writeLenDelimData(data []byte) error {
	n := binary.PutUvarint(ms.tmp, uint64(len(data)))
	n, err := ms.out.Write(ms.tmp[:n])
	if err != nil {
		return err
	}
	n, err = ms.out.Write(data)
	return err
}

func (ms *MsgStream) ReadMsg() (proto.Message, error) {
	typeNum, err := ms.ReadVint()
	if err != nil {
		return nil, err
	}
	data, err := ms.readLenDelimData()
	if err != nil {
		return nil, err
	}
	container, err := GetMsgContainer(typeNum)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(data, container)
	if err != nil {
		return nil, err
	}
	return container, nil
}

func (ms *MsgStream) WriteMsg(msg proto.Message) error {
	typeNum, err := GetMsgNum(msg)
	if err != nil {
		return err
	}
	err = ms.WriteVint(typeNum)
	if err != nil {
		return err
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	err = ms.writeLenDelimData(data)
	if err != nil {
		return err
	}
	ms.out.Flush()
	return err
}
