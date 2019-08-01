package mq

import "qingcloud.com/qing-cloud-mq/consumer"

type Config struct {
	Host    string
	Port    string
	GroupId string
	MsgMode consumer.GET_MSG_MODE
	Topic   string
}
