package message

import (
	"fmt"
)

type Client struct {
	pool *MsgStreamPool
}

func NewClient() *Client {
	return &Client{
		pool: NewMsgStreamPool("tcp", "localhost", "3233", 100),
	}
}

type CodeError struct {
	code uint32
}

func (e CodeError) Error() string {
	return fmt.Sprintf("Error code : %d", e.code)
}

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
		return "", &CodeError{logRes.Response.Code}
	}
	return logRes.Token, nil
}

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
		return nil, &CodeError{getRes.Response.Code}
	}
	return getRes.User, nil
}

func (c *Client) Authenticate(token string) (bool, error) {
	stream, err := c.pool.GetMsgStream()
	if err != nil {
		return false, err
	}
	err = stream.WriteMsg(&AuthRequest{Token: token})
	if err != nil {
		c.pool.destroyMsgStream(stream)
		return false, err
	}
	resMsg, err := stream.ReadMsg()
	if err != nil {
		c.pool.destroyMsgStream(stream)
		return false, err
	}
	c.pool.closeMsgStream(stream)

	authMsg := resMsg.(*Response)
	return authMsg.Code == uint32(0), nil
}

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

	editRes := resMsg.(*Response)
	if editRes.Code > uint32(0) {
		return &CodeError{editRes.Code}
	}
	return nil
}
