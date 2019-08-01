package mq

import (
	"fmt"
	"git.internal.yunify.com/MDMP2/cloudevents/pkg/cloudevents/client"
	"git.internal.yunify.com/MDMP2/cloudevents/pkg/cloudevents/transport/qmq"
	"github.com/golang/glog"
	"qingcloud.com/qing-cloud-mq/consumer"
	"qingcloud.com/qing-cloud-mq/producer"
)

func NewConsumerClient(cfg Config) client.Client {
	optC := consumer.NewOptions()
	optC.NSServer = cfg.Host + ":" + cfg.Port
	optC.GroupID = cfg.GroupId
	optC.MsgMode = consumer.PUSH

	c, err := qmq.NewConsumer(optC)
	if err != nil {
		glog.Errorf("failed to create nats transport, %s", err.Error())
		panic("failed to create nats transport")
	}
	cloudClient, err := client.New(c)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return cloudClient
}

func NewProducerClient(cfg Config) client.Client {
	optP := producer.NewOptions()
	optP.NsServerList = cfg.Host + ":" + cfg.Port
	optP.Topic = cfg.Topic

	p, err := qmq.NewProducer(optP)
	if err != nil {
		glog.Errorf("failed to create nats transport, %s", err.Error())
		panic("failed to create nats transport")
	}
	cloudClient, err := client.New(p)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return cloudClient
}
