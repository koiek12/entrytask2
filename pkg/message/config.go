package message

import (
	reflect "reflect"

	"google.golang.org/protobuf/proto"
)

// return message type number of message
func GetMsgNum(msg proto.Message) (uint, error) {
	msgType := reflect.TypeOf(msg).String()
	num, ok := msgNums[msgType]
	if !ok {
		return 0, nil
	}
	return num, nil
}

// find corresponding container by message type number and return it.
func GetMsgContainer(typeNum uint) (proto.Message, error) {
	containerFunc, ok := msgContainerFunc[typeNum]
	if !ok {
		return nil, nil
	}
	return containerFunc(), nil
}

// Mapping from message class name to message number.
// This map is used to determine which type number to write given message.
// Type number written will later be used by message reader to determine how to
// translate it. When adding new message, assign new number to message and configure this map
var msgNums = map[string]uint{
	"*message.HealthCheckMessage":  0,
	"*message.LoginRequest":        1,
	"*message.GetUserInfoRequest":  2,
	"*message.EditUserInfoRequest": 3,
	"*message.AuthRequest":         4,
	"*message.Response":            5,
	"*message.LoginResponse":       6,
	"*message.GetUserInfoResponse": 7,
}

// Mapping from message number to it's corresponding container generater.
// This map is used to determine which container to generate when reading the message given type number.
// When adding new message, assign new number to message and configure this map.
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
	5: func() proto.Message {
		return &Response{}
	},
	6: func() proto.Message {
		return &LoginResponse{}
	},
	7: func() proto.Message {
		return &GetUserInfoResponse{}
	},
}
