package message

import (
	"bufio"
	"encoding/binary"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
)

// MsgStream represent stream of messages. Message is abstraction of a request from client or a response from server.
// Every message have same format like below.
// Message format : | Message Type(varint) | Protobuf data(len-delim data) |
type MsgStream struct {
	conn net.Conn      // network connection for message
	in   *bufio.Reader // read incoming message using this Reader
	out  *bufio.Writer // write outgoing message using this Writer
	tmp  []byte        // temporal buffer for varint write
}

// NewMsgStream create new instance of MsgStream with network connection and read timeout duration.
func NewMsgStream(conn net.Conn, timeout time.Duration) (*MsgStream, error) {
	// set maxium read deadline for connection
	conn.SetReadDeadline(time.Now().Add(time.Second * timeout))
	return &MsgStream{conn, bufio.NewReader(conn), bufio.NewWriter(conn), make([]byte, 32)}, nil
}

// Close closes stream's undelying network connection.
func (ms *MsgStream) Close() {
	ms.conn.Close()
}

// readVarInt read a variable length integer from undelying network cconnection.
func (ms *MsgStream) readVarInt() (uint, error) {
	val, err := binary.ReadUvarint(ms.in)
	return uint(val), err
}

// writeVarInt write a variable length integer to undelying network cconnection.
func (ms *MsgStream) writeVarInt(val uint) error {
	n := binary.PutUvarint(ms.tmp, uint64(val))
	n, err := ms.out.Write(ms.tmp[:n])
	return err
}

// readLenDelimData read length-delimted data from stream
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

// writeLenDelimData write length-delimted data to stream
func (ms *MsgStream) writeLenDelimData(data []byte) error {
	n := binary.PutUvarint(ms.tmp, uint64(len(data)))
	n, err := ms.out.Write(ms.tmp[:n])
	if err != nil {
		return err
	}
	n, err = ms.out.Write(data)
	return err
}

// ReadMsg read a message from stream. Message consist of message type(varint) + data(protobuf data)
func (ms *MsgStream) ReadMsg() (proto.Message, error) {
	typeNum, err := ms.readVarInt()
	if err != nil {
		return nil, err
	}
	data, err := ms.readLenDelimData()
	if err != nil {
		return nil, err
	}
	// create empty message container
	container, err := getMsgContainer(typeNum)
	if err != nil {
		return nil, err
	}
	// write message protobuf data to empty container
	err = proto.Unmarshal(data, container)
	if err != nil {
		return nil, err
	}
	return container, nil
}

// WriteMsg wrtie a message to stream. Message consist of message type(varint) + data(protobuf data)
func (ms *MsgStream) WriteMsg(msg proto.Message) error {
	typeNum, err := getMsgNum(msg)
	if err != nil {
		return err
	}
	err = ms.writeVarInt(typeNum)
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

// RemoteAddr returns remote address of underlying network connection.
func (ms *MsgStream) RemoteAddr() string {
	return ms.conn.RemoteAddr().String()
}
