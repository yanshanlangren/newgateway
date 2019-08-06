package mqtt

import (
	"errors"
	"fmt"
	"newgateway/constant"
	"newgateway/logger"
	"newgateway/model"
)

func ParseMQTTMessage(byteArr []byte) ([]*model.MQTTMessage, error) {
	msgArr := make([]*model.MQTTMessage, 0)

	//捕获可能异常
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	for offset := 0; offset < len(byteArr)-2; {
		msg := &model.MQTTMessage{}
		//解析固定头
		fixedHeader := parseFixedHeader(byteArr[offset : offset+2])
		msg.FixedHeader = fixedHeader

		//解析可变头
		variableHeader, x, err := parseVariableHeader(byteArr[offset+2:], msg)
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}
		msg.VariableHeader = variableHeader

		//解析消息体
		payload, err := parsePayload(byteArr[2+x+offset:offset+msg.FixedHeader.RemainingLength+2], msg)
		if err != nil {
			return nil, err
		}
		msg.Payload = payload
		msgArr = append(msgArr, msg)
		offset += msg.FixedHeader.RemainingLength + 2
	}
	return msgArr, nil
}

//解析固定头
func parseFixedHeader(input []byte) *model.FixedHeader {
	header := &model.FixedHeader{}
	header.PackageType = int(input[0]) >> 4
	if header.PackageType == constant.MQTT_MSG_TYPE_PUBLISH {
		tokenInt := int(input[0]) & 15
		token := &model.SpecificToken{
			Retain: tokenInt % 2,
		}
		tokenInt >>= 1
		token.Qos = tokenInt % 4
		tokenInt >>= 2
		token.DUP = tokenInt % 2
		header.SpecificToken = token
	}
	header.RemainingLength = int(input[1])
	return header
}

//解析可变头
func parseVariableHeader(body []byte, msg *model.MQTTMessage) (*model.VariableHeader, int, error) {
	var (
		header *model.VariableHeader
		err    error
		offset int
	)
	if msg.FixedHeader == nil {
		return nil, 0, errors.New("fixed header null")
	}
	switch msg.FixedHeader.PackageType {
	case constant.MQTT_MSG_TYPE_CONNECT: //connect
		header, offset, err = parseConnectVariableHeader(body)
	case constant.MQTT_MSG_TYPE_PUBLISH: //publish
		header, offset, err = parsePublishVariableHeader(body, msg)
	case constant.MQTT_MSG_TYPE_SUBSCRIBE: //subscribe
		header, offset, err = parseSubscribeVariableHeader(body)
	case constant.MQTT_MSG_TYPE_UNSUBSCRIBE: //unsubscribe
		header, offset, err = parseUnsubscribeVariableHeader(body)
	}
	if err != nil {
		return nil, 0, err
	}
	return header, offset, nil
}

//处理connect消息的可变头
func parseConnectVariableHeader(body []byte) (*model.VariableHeader, int, error) {
	header := &model.VariableHeader{}
	length := int(body[1])
	//默认以UTF-8编码解析协议名
	header.ProtocolName = string(body[2 : 2+length])
	//协议版本
	header.ProtocolVersion = int(body[2+length])
	//连接标识
	header.ConnectFlags = &model.ConnectFlags{}
	connectFlags, err := parseConnectFlags(body[3+length])
	if err != nil {
		return nil, 0, err
	}
	header.ConnectFlags = connectFlags
	//keepalive
	header.KeepAliveTimer = int(body[4+length])<<8 + int(body[5+length])
	return header, 6 + length, nil
}

//处理publish消息的可变头
func parsePublishVariableHeader(body []byte, msg *model.MQTTMessage) (*model.VariableHeader, int, error) {
	header := &model.VariableHeader{}
	offset := 1
	topicLen := int(body[offset])
	offset++
	header.TopicName = string(body[offset : offset+topicLen])
	offset += topicLen
	if msg.FixedHeader.SpecificToken.Qos > 0 {
		header.MessageId = int(body[offset])<<8 + int(body[offset+1])
		offset += 2
	}
	return header, offset, nil
}

//处理subscribe消息的可变头
func parseSubscribeVariableHeader(body []byte) (*model.VariableHeader, int, error) {
	header := &model.VariableHeader{}
	header.MessageId = int(body[0])<<8 + int(body[1])
	return header, 2, nil
}

//处理unsubscribe消息的可变头
func parseUnsubscribeVariableHeader(body []byte) (*model.VariableHeader, int, error) {
	header := &model.VariableHeader{}
	header.MessageId = int(body[0])<<8 + int(body[1])
	return header, 2, nil
}

//解析标识符
func parseConnectFlags(body byte) (*model.ConnectFlags, error) {
	flags := &model.ConnectFlags{}
	tmp := int(body)
	tmp >>= 1
	flags.CleanSession = tmp % 2
	tmp >>= 1
	flags.WillFlag = tmp % 2
	tmp >>= 1
	flags.WillQos = tmp % 4
	tmp >>= 2
	flags.WillQos = tmp % 2
	tmp >>= 1
	flags.PasswordFlag = tmp % 2
	tmp >>= 1
	flags.UserNameFlag = tmp % 2
	return flags, nil
}

//解析消息头
func parsePayload(body []byte, msg *model.MQTTMessage) (*model.Payload, error) {
	var (
		payload *model.Payload
		err     error
	)
	switch msg.FixedHeader.PackageType {
	case constant.MQTT_MSG_TYPE_CONNECT:
		payload, err = parseConnectPayload(body, msg)
	case constant.MQTT_MSG_TYPE_SUBSCRIBE:
		payload, err = parseSubscribePayload(body, msg)
	case constant.MQTT_MSG_TYPE_UNSUBSCRIBE:
		payload, err = parseUnsubscribePayload(body, msg)
	case constant.MQTT_MSG_TYPE_PUBLISH:
		payload, err = parsePublishPayload(body, msg)
	}
	return payload, err
}

//解析connect payload
func parseConnectPayload(body []byte, msg *model.MQTTMessage) (*model.Payload, error) {
	//clientId
	clientIdLen := int(body[1])
	payload := &model.Payload{
		ClientId: string(body[2 : clientIdLen+2]),
	}
	offset := 3 + clientIdLen
	//Will topic & Will message
	if msg.VariableHeader.ConnectFlags.WillFlag == 1 {
		//will topic
		topicLen := int(body[offset])
		offset++
		payload.WillTopic = string(body[offset : offset+topicLen])
		logger.Debug("will topic:" + string(body[offset:offset+topicLen]))
		offset += topicLen
		offset++

		//will message
		messageLen := int(body[offset])
		offset++
		payload.WillMessage = string(body[offset : offset+messageLen])
		logger.Debug("will message:" + string(body[offset:offset+messageLen]))
		offset += messageLen
		offset++
	}
	//username
	if msg.VariableHeader.ConnectFlags.UserNameFlag == 1 {
		userLen := int(body[offset])
		offset++
		payload.UserName = string(body[offset : offset+userLen])
		logger.Debug("username:" + string(body[offset:offset+userLen]))
		offset += userLen
		offset++
	}
	//password
	if msg.VariableHeader.ConnectFlags.PasswordFlag == 1 {
		passLen := int(body[offset])
		offset++
		payload.Password = string(body[offset : offset+passLen])
		logger.Debug("password:" + string(body[offset:offset+passLen]))
		offset += passLen
		offset++
	}
	return payload, nil
}

//解析subscribe payload
func parseSubscribePayload(body []byte, msg *model.MQTTMessage) (*model.Payload, error) {
	subLen := int(body[1])
	payload := &model.Payload{
		SubscribePayload: string(body[2 : 2+subLen]),
		SubscribeAckQos:  int(body[subLen+2]),
	}
	return payload, nil
}

//解析unsubscribe payload
func parseUnsubscribePayload(body []byte, msg *model.MQTTMessage) (*model.Payload, error) {
	topicLen := body[1]
	payload := &model.Payload{
		UnsubscribeTopics: string(body[2 : 2+topicLen]),
	}
	return payload, nil
}

//解析publish payload
func parsePublishPayload(body []byte, msg *model.MQTTMessage) (*model.Payload, error) {
	payload := &model.Payload{
		Data: string(body),
	}
	return payload, nil
}
