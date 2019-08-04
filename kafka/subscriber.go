package kafka

import (
	"github.com/Shopify/sarama"
)

type Subscriber struct {
	Topic    string
	Consumer *Consumer
	PcList   []sarama.PartitionConsumer
}

func (c *Subscriber) Close() {
	for _, v := range (c.PcList) {
		v.Close()
	}
}