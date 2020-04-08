package message

import (
	"net"
	"testing"
	"time"
)

func TestMsgStreamHealthCheck(t *testing.T) {
	client, server := net.Pipe()
	clientStream, _ := NewMsgStream(server, 60*time.Second)
	serverStream, _ := NewMsgStream(client, 60*time.Second)
	done := make(chan bool)
	go func() {
		clientStream.WriteMsg(&HealthcheckMessage{})
		resMsg, _ := clientStream.ReadMsg()
		_, ok := resMsg.(*HealthcheckMessage)
		if !ok {
			t.Fail()
		}
		done <- true
	}()
	go func() {
		msg, _ := serverStream.ReadMsg()
		_, ok := msg.(*HealthcheckMessage)
		if !ok {
			t.Fail()
		}
		serverStream.WriteMsg(&HealthcheckMessage{})
	}()
	<-done
}

func TestMsgStreamLogin(t *testing.T) {
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
		if logRes.Token != "abcd" {
			t.Fail()
		}
		done <- true
	}()
	go func() {
		msg, _ := serverStream.ReadMsg()
		logReq, ok := msg.(*LoginRequest)
		if !ok {
			t.Fail()
		}
		t.Logf("id:%s, passwd:%s\n", logReq.Id, logReq.Password)
		if logReq.Id != "hoiek12" || logReq.Password != "abc" {
			t.Fail()
		}
		serverStream.WriteMsg(&LoginResponse{
			Token: "abcd",
		})
	}()
	<-done
}

func TestMsgStreamGetUserInfo(t *testing.T) {
	client, server := net.Pipe()
	clientStream, _ := NewMsgStream(server, 60*time.Second)
	serverStream, _ := NewMsgStream(client, 60*time.Second)
	done := make(chan bool)
	go func() {
		clientStream.WriteMsg(&GetUserInfoRequest{
			Token: "abcd",
		})
		resMsg, _ := clientStream.ReadMsg()
		getRes, ok := resMsg.(*GetUserInfoResponse)
		if !ok {
			t.Fail()
		}
		t.Logf(getRes.String())
		if getRes.User.Nickname != "nick" {
			t.Fail()
		}
		done <- true
	}()
	go func() {
		msg, _ := serverStream.ReadMsg()
		getReq, ok := msg.(*GetUserInfoRequest)
		if !ok {
			t.Fail()
		}
		t.Logf(getReq.String())
		if getReq.Token != "abcd" {
			t.Fail()
		}
		serverStream.WriteMsg(&GetUserInfoResponse{
			User: &User{
				Nickname: "nick",
			},
		})
	}()
	<-done
}

func TestMsgStreamEditUserInfo(t *testing.T) {
	client, server := net.Pipe()
	clientStream, _ := NewMsgStream(server, 60*time.Second)
	serverStream, _ := NewMsgStream(client, 60*time.Second)
	done := make(chan bool)
	go func() {
		clientStream.WriteMsg(&EditUserInfoRequest{
			Token: "abcd",
			User: &User{
				Nickname: "john",
			},
		})
		resMsg, _ := clientStream.ReadMsg()
		res, ok := resMsg.(*Response)
		if !ok {
			t.Fail()
		}
		t.Logf("code:%d", res.Code)
		if res.Code != 0 {
			t.Fail()
		}
		done <- true
	}()
	go func() {
		msg, _ := serverStream.ReadMsg()
		editReq, ok := msg.(*EditUserInfoRequest)
		if !ok {
			t.Fail()
		}
		t.Logf(editReq.String())
		if editReq.Token != "abcd" || editReq.User.Nickname != "john" {
			t.Fail()
		}
		serverStream.WriteMsg(&Response{
			Code: 0,
		})
	}()
	<-done
}

func TestMsgStreamAuthenticate(t *testing.T) {
	client, server := net.Pipe()
	clientStream, _ := NewMsgStream(server, 60*time.Second)
	serverStream, _ := NewMsgStream(client, 60*time.Second)
	done := make(chan bool)
	go func() {
		clientStream.WriteMsg(&AuthRequest{
			Token: "abcd",
		})
		resMsg, _ := clientStream.ReadMsg()
		res, ok := resMsg.(*Response)
		if !ok {
			t.Fail()
		}
		t.Logf("code:%d", res.Code)
		if res.Code != 0 {
			t.Fail()
		}
		done <- true
	}()
	go func() {
		msg, _ := serverStream.ReadMsg()
		authReq, ok := msg.(*AuthRequest)
		if !ok {
			t.Fail()
		}
		t.Logf(authReq.String())
		if authReq.Token != "abcd" {
			t.Fail()
		}
		serverStream.WriteMsg(&Response{
			Code: 0,
		})
	}()
	<-done
}
