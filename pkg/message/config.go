package message

import (
	reflect "reflect"

	"google.golang.org/protobuf/proto"
)

func GetMsgNum(msg proto.Message) (uint, error) {
	msgType := reflect.TypeOf(msg).String()
	num, ok := msgNums[msgType]
	if !ok {
		return 0, nil
	}
	return num, nil
}

func GetMsgContainer(typeNum uint) (proto.Message, error) {
	containerFunc, ok := msgContainerFunc[typeNum]
	if !ok {
		return nil, nil
	}
	return containerFunc(), nil
}

var msgNums = map[string]uint{
	"*message.HealthCheckMessage":  0,
	"*message.LoginRequest":        1,
	"*message.GetUserInfoRequest":  2,
	"*message.EditUserInfoRequest": 3,
	"*message.AuthRequest":         4,
	"*message.Response":            100,
	"*message.LoginResponse":       101,
	"*message.GetUserInfoResponse": 102,
}

var msgContainerFunc = map[uint]func() proto.Message{
	0: func() proto.Message {
		return &HealthcheckMessage{}
	},
	1: func() proto.Message {
		return &LoginRequest{}
	},
	2: func() proto.Message {
		return &GetUserInfoRequest{}
	},
	3: func() proto.Message {
		return &EditUserInfoRequest{}
	},
	4: func() proto.Message {
		return &AuthRequest{}
	},
	100: func() proto.Message {
		return &Response{}
	},
	101: func() proto.Message {
		return &LoginResponse{}
	},
	102: func() proto.Message {
		return &GetUserInfoResponse{}
	},
}
