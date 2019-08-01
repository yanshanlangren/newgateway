package kafka

import (
	"github.com/Shopify/sarama"
	"newgateway/config"
	"regexp"
)

type Subscriber struct {
	Topic    string
	Consumer *sarama.Consumer
	Ch       chan []byte
	PcList   []sarama.PartitionConsumer
}

func (c *Subscriber) Close() {
	close(c.Ch)
	for _, v := range (c.PcList) {
		v.Close()
	}
}

var kafkaConsumer = initConsumer()

func initConsumer() *sarama.Consumer {
	consumer, err := sarama.NewConsumer(config.GetConfig().Kafka.ServerList, nil)
	if err != nil {
		panic(err)
		return nil
	}
	return &consumer
}

func NewSubscriber(topic string, consumerBufferSize int) (*Subscriber, error) {
	//Partitions(topic):该方法返回了该topic的所有分区id
	partitionList, err := (*kafkaConsumer).Partitions(topic)
	if err != nil {
		return nil, err
	}

	sub := &Subscriber{
		Topic:    topic,
		Consumer: kafkaConsumer,
		PcList:   make([]sarama.PartitionConsumer, 0),
		Ch:       make(chan []byte, consumerBufferSize),
	}

	for partition := range partitionList {
		//ConsumePartition方法根据主题，分区和给定的偏移量创建创建了相应的分区消费者
		//如果该分区消费者已经消费了该信息将会返回error
		//sarama.OffsetNewest:表明了为最新消息
		pc, err := (*kafkaConsumer).ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			return nil, err
		}

		//添加到列表中
		sub.PcList = append(sub.PcList, pc)
	}
	return sub, nil
}

func NewSubscribers(tpc string, consumerBufferSize int) ([]*Subscriber, error) {
	//Partitions(topic):该方法返回了该topic的所有分区id
	var subs []*Subscriber
	topicList, err := (*kafkaConsumer).Topics()
	if err != nil {
		return nil, err
	}
	for _, topic := range (topicList) {
		if match, err := regexp.Match(tpc, []byte(topic)); err == nil && match {

			partitionList, err := (*kafkaConsumer).Partitions(topic)
			if err != nil {
				return nil, err
			}
			sub := &Subscriber{
				Topic:    topic,
				Consumer: kafkaConsumer,
				PcList:   make([]sarama.PartitionConsumer, 0),
				Ch:       make(chan []byte, consumerBufferSize),
			}

			for partition := range partitionList {
				//ConsumePartition方法根据主题，分区和给定的偏移量创建创建了相应的分区消费者
				//如果该分区消费者已经消费了该信息将会返回error
				//sarama.OffsetNewest:表明了为最新消息
				pc, err := (*kafkaConsumer).ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
				if err != nil {
					return nil, err
				}

				//添加到列表中
				sub.PcList = append(sub.PcList, pc)
			}
			subs = append(subs, sub)
		}
	}

	return subs, nil
}
