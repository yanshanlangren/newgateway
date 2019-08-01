package mqtt

import (
	"errors"
	"newgateway/constant"
	"newgateway/model"
)

func MQTT2ByteArr(msg *model.MQTTMessage) ([]byte, error) {
	arr := make([]byte, 0)
	if msg == nil {
		return arr, errors.New("MQTT message empty")
	}
	//固定头
	arr = appendFixedHeader(msg, arr)
	//可变头
	arr = appendVariableHeader(msg, arr)
	//payload
	arr = appendPayload(msg, arr)
	return arr, nil
}

//创建固定头
func appendFixedHeader(msg *model.MQTTMessage, arr []byte) []byte {
	if msg.FixedHeader != nil {
		//消息类型
		arr = append(arr, byte(msg.FixedHeader.PackageType<<4))
		//剩余长度
		arr = append(arr, byte(msg.FixedHeader.RemainingLength))
	}
	return arr
}

//可变头
func appendVariableHeader(msg *model.MQTTMessage, arr []byte) []byte {
	if msg.FixedHeader != nil && msg.VariableHeader != nil {
		switch msg.FixedHeader.PackageType {
		case constant.MQTT_MSG_TYPE_CONNECTACK:
			arr = append(arr, byte(0))
			arr = append(arr, byte(msg.VariableHeader.ConnectReturnCode))
		case constant.MQTT_MSG_TYPE_PUBLISH:
			arr = append(arr, byte(0))
			arr = append(arr, byte(len(msg.VariableHeader.TopicName)))
			arr = appendArray(arr, []byte(msg.VariableHeader.TopicName))
			if msg.FixedHeader.SpecificToken.Qos > 0 {
				msgInt := int(msg.VariableHeader.MessageId)
				arr = append(arr, byte(msgInt>>8))
				arr = append(arr, byte(msgInt))
			}
		case constant.MQTT_MSG_TYPE_PUBACK, constant.MQTT_MSG_TYPE_PUBREC, constant.MQTT_MSG_TYPE_PUBREL, constant.MQTT_MSG_TYPE_SUBACK, constant.MQTT_MSG_TYPE_UNSUBACK:
			msgInt := int(msg.VariableHeader.MessageId)
			arr = append(arr, byte(msgInt>>8))
			arr = append(arr, byte(msgInt))
		}
	}
	return arr
}

//正文
func appendPayload(msg *model.MQTTMessage, arr []byte) []byte {
	if msg.FixedHeader != nil && msg.Payload != nil {
		switch msg.FixedHeader.PackageType {
		case constant.MQTT_MSG_TYPE_PUBLISH:
			arr = appendArray(arr, []byte(msg.Payload.Data))
		case constant.MQTT_MSG_TYPE_SUBACK:
			arr = append(arr, byte(msg.Payload.SubscribeAckQos))
		}
	}
	return arr
}

func appendArray(arr, tail []byte) []byte {
	for _, v := range (tail) {
		arr = append(arr, v)
	}
	return arr
}
