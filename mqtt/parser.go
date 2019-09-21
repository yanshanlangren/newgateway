package mqtt

import (
	"errors"
	"newgateway/constant"
	"newgateway/logger"
	"newgateway/model"
	"newgateway/utils"
	"strconv"
)

func (c *Client) DealByteArray(byteArr []byte) {
	var offset = 0
	//捕获可能异常
	defer func() {
		if err := recover(); err != nil {
			if offset > 0 {
				logger.Warn("parse error, array length = ", len(byteArr), " offset = ", offset, " msg = ", string(byteArr))
				copy(c.Buffer[c.BufferOffset:], byteArr[offset:])
				c.BufferOffset += len(byteArr) - offset
				c.IsBufferEmpty = false
			} else {
				logger.Warn("emergency error, array length = ", len(byteArr), " offset = ", offset, " msg(from offset) = ", string(byteArr[offset:]))
				////试探下一个完整的数据节点
				for offset++; offset < len(byteArr); offset++ {
					if msg, err := ParseByteArray(byteArr[offset:]); msg != nil && err == nil {
						break
					}
				}
				logger.Info("new pointer found, array length = ", len(byteArr), " offset = ", offset, " msg = ", string(byteArr[offset:]))
				c.DealByteArray(byteArr[offset:])

				//丢掉出问题的数据
				c.BufferOffset = 0
				c.IsBufferEmpty = true
			}
		}
	}()
	if !c.IsBufferEmpty {
		byteArr = appendArray(c.Buffer[:c.BufferOffset], byteArr)
		c.BufferOffset = 0
	}

	//Parse(byteArr,&offset,func(msg *model.MQTTMessage){
	//	//处理消息业务逻辑
	//	if msg.FixedHeader != nil {
	//		logger.Debug("message type: " + strconv.Itoa(msg.FixedHeader.PackageType))
	//		resMsg := c.DealMQTTMessage(msg)
	//		if resMsg != nil {
	//			go c.Write(resMsg)
	//		}
	//	}
	//})

	for offset = 0; offset < len(byteArr); {
		msg := &model.MQTTMessage{}
		//解析固定头
		fixedHeader, fixHeaderLen := parseFixedHeader(byteArr[offset:])
		msg.FixedHeader = fixedHeader

		//解析可变头
		variableHeader, varHeaderLen := parseVariableHeader(byteArr[offset+fixHeaderLen:offset+fixHeaderLen+msg.FixedHeader.RemainingLength], msg)
		msg.VariableHeader = variableHeader

		//解析消息体
		msg.Payload = parsePayload(byteArr[fixHeaderLen+varHeaderLen+offset:offset+msg.FixedHeader.RemainingLength+fixHeaderLen], msg)

		logger.Debug("receive message type: "+strconv.Itoa(msg.FixedHeader.PackageType)+", body: ", string(byteArr[0:msg.FixedHeader.RemainingLength+fixHeaderLen]))

		offset += msg.FixedHeader.RemainingLength + fixHeaderLen
		//处理消息业务逻辑
		if msg.FixedHeader != nil {
			resMsg := c.DealMQTTMessage(msg)
			if resMsg != nil {
				go c.Write(resMsg)
			}
		}
	}
	c.IsBufferEmpty = true
}

func Parse(byteArr []byte, offset *int, f func(msg *model.MQTTMessage)) {
	for *offset = 0; *offset < len(byteArr); {
		msg := &model.MQTTMessage{}
		//解析固定头
		fixedHeader, fixHeaderLen := parseFixedHeader(byteArr[:])
		msg.FixedHeader = fixedHeader

		//解析可变头
		variableHeader, varHeaderLen := parseVariableHeader(byteArr[fixHeaderLen:fixHeaderLen+msg.FixedHeader.RemainingLength], msg)
		msg.VariableHeader = variableHeader

		//解析消息体
		msg.Payload = parsePayload(byteArr[fixHeaderLen+varHeaderLen:msg.FixedHeader.RemainingLength+fixHeaderLen], msg)
		*offset += msg.FixedHeader.RemainingLength + fixHeaderLen
		f(msg)
	}
}

func ParseByteArray(byteArr []byte) (*model.MQTTMessage, error) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	msg := &model.MQTTMessage{}
	//解析固定头
	fixedHeader, fixHeaderLen := parseFixedHeader(byteArr[:])
	msg.FixedHeader = fixedHeader

	//解析可变头
	if len(byteArr) < fixHeaderLen+msg.FixedHeader.RemainingLength {
		return nil, errors.New("invalid message")
	}
	variableHeader, varHeaderLen := parseVariableHeader(byteArr[fixHeaderLen:fixHeaderLen+msg.FixedHeader.RemainingLength], msg)
	msg.VariableHeader = variableHeader

	//解析消息体
	if len(byteArr) < fixHeaderLen+varHeaderLen {
		return nil, errors.New("invalid message")
	}
	msg.Payload = parsePayload(byteArr[fixHeaderLen+varHeaderLen:msg.FixedHeader.RemainingLength+fixHeaderLen], msg)

	logger.Debug("message type: " + strconv.Itoa(msg.FixedHeader.PackageType))
	//处理消息业务逻辑
	return msg, nil
}

//解析固定头
func parseFixedHeader(input []byte) (*model.FixedHeader, int) {
	header := &model.FixedHeader{
		PackageType: int(input[0]) >> 4,
	}
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
	flag := int(input[1]) >> 7
	header.RemainingLength = int(input[1] & 127)
	var i int
	for i = 2; flag == 1; i++ {
		header.RemainingLength += int(input[i]&127) * utils.Pow(128, (i - 1))
		flag = int(input[i]) >> 7
	}

	return header, i
}

//解析可变头
func parseVariableHeader(body []byte, msg *model.MQTTMessage) (*model.VariableHeader, int) {
	var (
		header *model.VariableHeader
		offset int
	)
	switch msg.FixedHeader.PackageType {
	case constant.MQTT_MSG_TYPE_CONNECT: //connect
		header, offset = parseConnectVariableHeader(body)
	case constant.MQTT_MSG_TYPE_PUBLISH: //publish
		header, offset = parsePublishVariableHeader(body, msg)
	case constant.MQTT_MSG_TYPE_PUBREL: //pubrel
		header, offset = parsePubrelVariableHeader(body)
	case constant.MQTT_MSG_TYPE_SUBSCRIBE: //subscribe
		header, offset = parseSubscribeVariableHeader(body)
	case constant.MQTT_MSG_TYPE_UNSUBSCRIBE: //unsubscribe
		header, offset = parseUnsubscribeVariableHeader(body)
	}
	return header, offset
}

//处理connect消息的可变头
func parseConnectVariableHeader(body []byte) (*model.VariableHeader, int) {
	header := &model.VariableHeader{}
	length := int(body[1])
	//默认以UTF-8编码解析协议名
	header.ProtocolName = string(body[2 : 2+length])
	//协议版本
	header.ProtocolVersion = int(body[2+length])
	//连接标识
	header.ConnectFlags = parseConnectFlags(body[3+length])
	//keepalive
	header.KeepAliveTimer = int(body[4+length])<<8 + int(body[5+length])
	return header, 6 + length
}

//处理publish消息的可变头
func parsePublishVariableHeader(body []byte, msg *model.MQTTMessage) (*model.VariableHeader, int) {
	header := &model.VariableHeader{}
	topicLen := int(body[0])<<8 + int(body[1])
	header.TopicName = string(body[2 : 2+topicLen])
	offset := 2 + topicLen
	if msg.FixedHeader.SpecificToken.Qos > 0 {
		header.MessageId = int(body[offset])<<8 + int(body[offset+1])
		offset += 2
	}
	return header, offset
}

//处理subscribe消息的可变头
func parsePubrelVariableHeader(body []byte) (*model.VariableHeader, int) {
	header := &model.VariableHeader{}
	header.MessageId = int(body[0])<<8 + int(body[1])
	return header, 2
}

//处理subscribe消息的可变头
func parseSubscribeVariableHeader(body []byte) (*model.VariableHeader, int) {
	header := &model.VariableHeader{}
	header.MessageId = int(body[0])<<8 + int(body[1])
	return header, 2
}

//处理unsubscribe消息的可变头
func parseUnsubscribeVariableHeader(body []byte) (*model.VariableHeader, int) {
	header := &model.VariableHeader{}
	header.MessageId = int(body[0])<<8 + int(body[1])
	return header, 2
}

//解析标识符
func parseConnectFlags(body byte) *model.ConnectFlags {
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
	return flags
}

//解析消息头
func parsePayload(body []byte, msg *model.MQTTMessage) *model.Payload {
	var payload *model.Payload
	switch msg.FixedHeader.PackageType {
	case constant.MQTT_MSG_TYPE_CONNECT:
		payload = parseConnectPayload(body, msg)
	case constant.MQTT_MSG_TYPE_SUBSCRIBE:
		payload = parseSubscribePayload(body, msg)
	case constant.MQTT_MSG_TYPE_UNSUBSCRIBE:
		payload = parseUnsubscribePayload(body, msg)
	case constant.MQTT_MSG_TYPE_PUBLISH:
		payload = parsePublishPayload(body, msg)
	}
	return payload
}

//解析connect payload
func parseConnectPayload(body []byte, msg *model.MQTTMessage) *model.Payload {
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
	return payload
}

//解析subscribe payload
func parseSubscribePayload(body []byte, msg *model.MQTTMessage) *model.Payload {
	subLen := int(body[1])
	payload := &model.Payload{
		SubscribePayload: string(body[2 : 2+subLen]),
		SubscribeAckQos:  int(body[subLen+2]),
	}
	return payload
}

//解析unsubscribe payload
func parseUnsubscribePayload(body []byte, msg *model.MQTTMessage) *model.Payload {
	return &model.Payload{
		UnsubscribeTopics: string(body[2 : 2+int(body[1])]),
	}
}

//解析publish payload
func parsePublishPayload(body []byte, msg *model.MQTTMessage) *model.Payload {
	return &model.Payload{
		Data: string(body),
	}
}
