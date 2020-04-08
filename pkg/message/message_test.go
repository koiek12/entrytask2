package message

import (
	"net"
	"testing"
	"time"
)

//TestMsgStream is test
func TestMsgStream(t *testing.T) {
	client, server := net.Pipe()
	clientStream, _ := NewMsgStream(server, 60*time.Second)
	serverStream, _ := NewMsgStream(client, 60*time.Second)
	done := make(chan bool)
	go func() {
		clientStream.WriteMsg(&LoginRequest{
			Id:       "hoiek12",
			Password: "abc",
		})
		resMsg, _ := clientStream.ReadMsg()
		logRes, ok := resMsg.(*LoginResponse)
		if !ok {
			t.Fail()
		}
		t.Logf("token:%s\n", logRes.Token)
		if token != "abcd" {
			t.Fail()
		}
		done <- true
	}()
	go func() {
		msg, _ := serverStream.ReadMsg()
		logReq, ok := msg.(*LoginRequest)
		if !ok || logReq.Id != "hoiek12" || logReq.Password != "abc" {
			t.Fail()
		}
		t.Logf("id:%s, passwd:%s\n", logReq.Id, logReq.Password)
		serverStream.WriteMsg(&LoginResponse{
			Token: "abcd",
		})
	}()
	<-done
}
