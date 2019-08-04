package mqtt

import (
	"errors"
	"github.com/Shopify/sarama"
	"net"
	"newgateway/common"
	"newgateway/constant"
	"newgateway/kafka"
	"newgateway/logger"
	"newgateway/model"
	"regexp"
	"strings"
	"sync"
	"time"
)

// 客户端连接的抽象
type Client struct {
	// tcp 连接
	Conn net.Conn
	// 当服务端开始发送数据时进入waiting, 阻止其它goroutine关闭连接
	Waiting common.Wait
	//Qos=2的消息
	Assure sync.Map
	//Client下的订阅
	SubscribeMap sync.Map
	//Assure定时发送的时间间隔
	Ticker int
	//Client关闭信号
	Closing chan bool
	//Client是否关闭
	Closed bool
	//关闭Assure的信号
	AssureClosing chan bool
	//kafka消费
	Consumer *kafka.Consumer
}

// 关闭客户端连接
func (c *Client) Close() error {
	// 等待数据发送完成或超时
	c.Waiting.WaitWithTimeout(10 * time.Second)
	//关闭订阅组
	c.SubscribeMap.Range(func(k, v interface{}) bool {
		sub := v.(*kafka.Subscriber)
		sub.Close()
		return true
	})
	//关闭Assure
	c.AssureClosing <- true
	c.Conn.Close()
	if c.Consumer != nil {
		c.Consumer.Release()
	}
	c.Closed = true
	return nil
}

//处理字节数组
func (c *Client) Deal(arr []byte) {
	//logger.Info("message body:" + string(arr))
	reqArr, err1 := ParseMQTTMessage(arr)
	if err1 != nil {
		logger.Fatal(err1.Error())
		return
	}
	for _, reqMsg := range (reqArr) {
		//logger.Info("message type:" + strconv.Itoa(reqMsg.FixedHeader.PackageType))
		//处理消息业务逻辑
		resMsg, err2 := c.DealMQTTMessage(reqMsg)
		if err2 != nil {
			logger.Fatal(err2.Error())
			return
		}
		go c.Write(resMsg)
	}
}

func (c *Client) Write(msg *model.MQTTMessage) {
	resByte, err3 := MQTT2ByteArr(msg)
	if err3 != nil {
		logger.Fatal(err3.Error())
		return
	}
	if resByte != nil && len(resByte) > 0 {
		//logger.Info("return type: " + strconv.Itoa(msg.FixedHeader.PackageType) + "\r\nreturn message: " + string(resByte))
		// 发送数据前先置为waiting状态
		c.Waiting.Add(1)
		//写返回
		c.Conn.Write(resByte)
		// 发送完毕, 结束waiting
		c.Waiting.Done()
	}
}

//针对publish qos=2的消息,启动轮询发送确保的消息
func (c *Client) Tick() {
	tick := time.Tick(time.Duration(c.Ticker) * time.Second)
	for {
		select {
		case <-tick:
			c.Assure.Range(func(key, value interface{}) bool {
				msg := key.(*model.MQTTMessage)
				c.Write(msg)
				return true
			})
		case <-c.AssureClosing:
			return
		}
	}
}

//取消订阅
func (c *Client) Unsubscribe(topicName string) {
	val, ok := c.SubscribeMap.Load(topicName)
	if !ok {
		return
	}
	sub := val.(*kafka.Subscriber)
	sub.Close()
	c.SubscribeMap.Delete(topicName)
}

//订阅
func (c *Client) Subscribe(s *kafka.Subscriber, msg *model.MQTTMessage) {
	c.SubscribeMap.Store(s.Topic, s)
	count := 0
	for _, pc := range (s.PcList) {
		go func(sarama.PartitionConsumer) {
			//Messages()该方法返回一个消费消息类型的只读通道，由代理产生
			for message := range pc.Messages() {
				//fmt.Printf("%s---Partition:%d, Offset:%d, Key:%s, Value:%s\n", msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
				arr := message.Value
				pub := &model.MQTTMessage{
					FixedHeader: &model.FixedHeader{
						PackageType: constant.MQTT_MSG_TYPE_PUBLISH,
						SpecificToken: &model.SpecificToken{
							DUP:    0,
							Qos:    msg.Payload.SubscribeAckQos,
							Retain: 0,
						},
						RemainingLength: 2 + len(msg.Payload.SubscribePayload) + len(arr),
					},
					VariableHeader: &model.VariableHeader{
						TopicName: msg.Payload.SubscribePayload,
						MessageId: count,
					},
					Payload: &model.Payload{
						Data: string(arr),
					},
				}
				if msg.Payload.SubscribeAckQos > 0 {
					pub.FixedHeader.RemainingLength = 4 + len(msg.Payload.SubscribePayload) + len(arr)
				}
				count++
				go c.Write(pub)
			}
		}(pc)
	}
}

//处理业务逻辑并返回
func (cli *Client) DealMQTTMessage(msg *model.MQTTMessage) (*model.MQTTMessage, error) {
	var retMsg *model.MQTTMessage
	if msg.FixedHeader == nil {
		return nil, errors.New("fixed header is null")
	}
	switch msg.FixedHeader.PackageType {
	case constant.MQTT_MSG_TYPE_DISCONNECT: //disconnect
		retMsg = cli.dealDisconnect(msg)
	case constant.MQTT_MSG_TYPE_PUBLISH: //publish
		retMsg = cli.dealPublish(msg)
	case constant.MQTT_MSG_TYPE_PUBREL: //pubrel
		retMsg = cli.dealPubrel(msg)
	case constant.MQTT_MSG_TYPE_SUBSCRIBE: //subscribe
		retMsg = cli.dealSubscribe(msg)
	case constant.MQTT_MSG_TYPE_UNSUBSCRIBE: //unsubscribe
		retMsg = cli.dealUnsubscribe(msg)
	case constant.MQTT_MSG_TYPE_PINGREQ: //pingreq
		retMsg = cli.dealPing(msg)
	}
	return retMsg, nil
}

//Publish
func (cli *Client) dealPublish(msg *model.MQTTMessage) *model.MQTTMessage {
	//发布消息
	//par, offset, err := kafka.Publish(msg.VariableHeader.TopicName, msg.Payload.Data)
	go kafka.AsyncPublish(msg.VariableHeader.TopicName, msg.Payload.Data)
	//err := kafka.BatchPublish(msg.VariableHeader.TopicName,)
	//if err != nil {
	//	logger.Fatal(err.Error())
	//} else {
	//	logger.Info("Partition:" + strconv.Itoa(int(par)) + ", offset:" + strconv.FormatInt(offset, 10))
	//}

	//产生返回值
	switch msg.FixedHeader.SpecificToken.Qos {
	case 0:
		return &model.MQTTMessage{}
	case 1:
		//返回PUBACK消息
		return &model.MQTTMessage{
			FixedHeader: &model.FixedHeader{
				PackageType:     constant.MQTT_MSG_TYPE_PUBACK,
				RemainingLength: 2,
			},
			VariableHeader: &model.VariableHeader{
				MessageId: msg.VariableHeader.MessageId,
			},
		}
	case 2:
		//生产一条PUBREC消息, 发送给消息发送方, 并期待接收到PUBREL消息
		pubrec := &model.MQTTMessage{
			FixedHeader: &model.FixedHeader{
				PackageType:     constant.MQTT_MSG_TYPE_PUBREC,
				RemainingLength: 2,
			},
			VariableHeader: &model.VariableHeader{
				MessageId: msg.VariableHeader.MessageId,
			},
		}
		cli.Assure.Store(pubrec, 1)

		//返回PUBACK消息
		return &model.MQTTMessage{
			FixedHeader: &model.FixedHeader{
				PackageType:     constant.MQTT_MSG_TYPE_PUBACK,
				RemainingLength: 2,
			},
			VariableHeader: &model.VariableHeader{
				MessageId: msg.VariableHeader.MessageId,
			},
		}
	}
	return nil
}

//Subscribe
func (cli *Client) dealSubscribe(msg *model.MQTTMessage) *model.MQTTMessage {
	//订阅
	if cli.Consumer == nil {
		c := kafka.GetConsumer()
		if c == nil {
			return &model.MQTTMessage{}
		}
		cli.Consumer = c
	}
	if strings.Index(msg.Payload.SubscribePayload, "*") != -1 {
		s, err := cli.Consumer.NewSubscribers(msg.Payload.SubscribePayload, 200)
		if err != nil {
			logger.Fatal(err.Error())
			return &model.MQTTMessage{}
		}
		for _, sub := range (s) {
			go cli.Subscribe(sub, msg)
		}
	} else {
		s, err := cli.Consumer.NewSubscriber(msg.Payload.SubscribePayload, 200)
		if err != nil {
			logger.Fatal(err.Error())
			return &model.MQTTMessage{}
		}
		go cli.Subscribe(s, msg)
	}

	//产生SUBACK消息
	return &model.MQTTMessage{
		FixedHeader: &model.FixedHeader{
			PackageType:     constant.MQTT_MSG_TYPE_SUBACK,
			RemainingLength: 2,
		},
		VariableHeader: &model.VariableHeader{
			MessageId: msg.VariableHeader.MessageId,
		},
	}
}

//TODO Unsubscribe
func (cli *Client) dealUnsubscribe(msg *model.MQTTMessage) *model.MQTTMessage {
	//删除订阅
	for _, topic := range (strings.Split(msg.Payload.UnsubscribeTopics, ",")) {
		cli.SubscribeMap.Range(func(key, val interface{}) bool {
			if match, err := regexp.Match(topic, []byte(key.(string))); err == nil && match {
				cli.Unsubscribe(key.(string))
			}
			return true
		})
	}
	//产生UNSUBACK消息
	res := &model.MQTTMessage{
		FixedHeader: &model.FixedHeader{
			PackageType:     constant.MQTT_MSG_TYPE_UNSUBACK,
			RemainingLength: 2,
		},
		VariableHeader: &model.VariableHeader{
			MessageId: msg.VariableHeader.MessageId,
		},
	}
	return res
}

//Ping
func (cli *Client) dealPing(msg *model.MQTTMessage) *model.MQTTMessage {
	//产生PINGRESP消息
	return &model.MQTTMessage{
		FixedHeader: &model.FixedHeader{
			PackageType:     constant.MQTT_MSG_TYPE_PINGRESP,
			RemainingLength: 0,
		},
	}
}

//Pubrel publish端发过来的Pubrec消息的返回
func (cli *Client) dealPubrel(msg *model.MQTTMessage) *model.MQTTMessage {
	//TODO 处理Pubrel消息
	//产生一条空消息
	return &model.MQTTMessage{}
}

//Disconnect
func (cli *Client) dealDisconnect(msg *model.MQTTMessage) *model.MQTTMessage {
	//断开连接
	cli.Closing <- true
	//产生一条空消息
	return &model.MQTTMessage{}
}
