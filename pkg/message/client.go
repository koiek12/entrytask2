package message

import (
	"fmt"
)

// Client to request and get response from backend TCP server
// Client use tcp connection pool(MsgStreamPool) whenever there is request.
type Client struct {
	pool *MsgStreamPool
}

// Create New Client to connect server.
func NewClient(host, port string, maxConn int) *Client {
	return &Client{
		pool: NewMsgStreamPool("tcp", host, port, int32(maxConn)),
	}
}

type AuthError struct{}

func (e AuthError) Error() string {
	return fmt.Sprintf("Authentication fail")
}

type DBError struct{}

func (e DBError) Error() string {
	return "DB error"
}

type InputError struct{}

func (e InputError) Error() string {
	return "Wrong input"
}

type UnknownError struct{}

func (e UnknownError) Error() string {
	return "Unknown Error"
}

// function to create error from TCP message error code
func getErrorFromCode(code uint32) error {
	switch code {
	case 1:
		return AuthError{}
	case 2:
		return DBError{}
	case 3:
		return InputError{}
	default:
		return UnknownError{}
	}
}

// try login in backend server, If success, token is returned.
// return error on network or backend server failure, in this case token is empty string
func (c *Client) Login(id, password string) (string, error) {
	stream, err := c.pool.GetMsgStream()
	if err != nil {
		return "", err
	}
	err = stream.WriteMsg(&LoginRequest{
		Id:       id,
		Password: password,
	})
	if err != nil {
		c.pool.destroyMsgStream(stream)
		return "", err
	}
	resMsg, err := stream.ReadMsg()
	if err != nil {
		c.pool.destroyMsgStream(stream)
		return "", err
	}
	c.pool.closeMsgStream(stream)

	logRes := resMsg.(*LoginResponse)
	if logRes.Response.Code > uint32(0) {
		return "", getErrorFromCode(logRes.Response.Code)
	}
	return logRes.Token, nil
}

// Get user information from backend TCP server.
// return error on network or backend server failure
func (c *Client) GetUserInfo(token string) (*User, error) {
	stream, err := c.pool.GetMsgStream()
	if err != nil {
		return nil, err
	}
	err = stream.WriteMsg(&GetUserInfoRequest{Token: token})
	if err != nil {
		c.pool.destroyMsgStream(stream)
		return nil, err
	}
	resMsg, err := stream.ReadMsg()
	if err != nil {
		c.pool.destroyMsgStream(stream)
		return nil, err
	}
	c.pool.closeMsgStream(stream)

	getRes := resMsg.(*GetUserInfoResponse)
	if getRes.Response.Code > uint32(0) {
		return nil, getErrorFromCode(getRes.Response.Code)
	}
	return getRes.User, nil
}

// Authenticate JWT access token
// return error on network or backend server failure
func (c *Client) Authenticate(token string) error {
	stream, err := c.pool.GetMsgStream()
	if err != nil {
		return err
	}
	err = stream.WriteMsg(&AuthRequest{Token: token})
	if err != nil {
		c.pool.destroyMsgStream(stream)
		return err
	}
	resMsg, err := stream.ReadMsg()
	if err != nil {
		c.pool.destroyMsgStream(stream)
		return err
	}
	c.pool.closeMsgStream(stream)

	res := resMsg.(*Response)
	if res.Code > uint32(0) {
		return getErrorFromCode(res.Code)
	}
	return nil
}

// Edit User information from backend TCP server
// return error on network or backend server failure
func (c *Client) EditUserInfo(token string, user *User) error {
	stream, err := c.pool.GetMsgStream()
	if err != nil {
		return err
	}
	err = stream.WriteMsg(&EditUserInfoRequest{
		Token: token,
		User:  user,
	})
	if err != nil {
		c.pool.destroyMsgStream(stream)
		return err
	}
	resMsg, err := stream.ReadMsg()
	if err != nil {
		c.pool.destroyMsgStream(stream)
		return err
	}
	c.pool.closeMsgStream(stream)

	res := resMsg.(*Response)
	if res.Code > uint32(0) {
		return getErrorFromCode(res.Code)
	}
	return nil
}
