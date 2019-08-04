package kafka

import (
	"github.com/Shopify/sarama"
	"newgateway/config"
	"regexp"
	"sync"
)

var consumerPool *ConsumerPool

type Consumer struct {
	consumer *sarama.Consumer
}

func init() {
	var size = config.GetConfig().Kafka.ConsumerPoolSize
	consumerPool = &ConsumerPool{
		size:      size,
		consumers: make(map[int]*Consumer, size),
		isIdle:    make(map[int]bool, size),
		mu:        sync.Mutex{},
	}
	for i := 0; i < size; i++ {
		consumerPool.consumers[i] = newConsumer()
		consumerPool.isIdle[i] = true
	}
}

type ConsumerPool struct {
	size      int
	consumers map[int]*Consumer
	isIdle    map[int]bool
	mu        sync.Mutex
}

func newConsumer() *Consumer {
	consumer, err := sarama.NewConsumer(config.GetConfig().Kafka.ServerList, nil)
	if err != nil {
		panic(err)
		return nil
	}
	return &Consumer{
		consumer: &consumer,
	}
}

func GetConsumer() *Consumer {
	mu.Lock()
	defer mu.Unlock()
	for k, v := range (consumerPool.isIdle) {
		if v {
			consumerPool.isIdle[k] = false
			return consumerPool.consumers[k]
		}
	}
	return newConsumer()
}

func (c *Consumer) Release() {
	mu.Lock()
	defer mu.Unlock()
	for k, v := range (consumerPool.consumers) {
		if v == c {
			consumerPool.isIdle[k] = true
			return
		}
	}
	(*c.consumer).Close()
}

func (c *Consumer) NewSubscriber(topic string, consumerBufferSize int) (*Subscriber, error) {
	//Partitions(topic):该方法返回了该topic的所有分区id
	partitionList, err := (*c.consumer).Partitions(topic)
	if err != nil {
		return nil, err
	}

	sub := &Subscriber{
		Topic:    topic,
		Consumer: c,
		PcList:   make([]sarama.PartitionConsumer, 0),
	}

	for partition := range partitionList {
		//ConsumePartition方法根据主题，分区和给定的偏移量创建创建了相应的分区消费者
		//如果该分区消费者已经消费了该信息将会返回error
		//sarama.OffsetNewest:表明了为最新消息
		pc, err := (*c.consumer).ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			return nil, err
		}

		//添加到列表中
		sub.PcList = append(sub.PcList, pc)
	}
	return sub, nil
}

func (c *Consumer) NewSubscribers(tpc string, consumerBufferSize int) ([]*Subscriber, error) {
	//Partitions(topic):该方法返回了该topic的所有分区id
	var subs []*Subscriber
	topicList, err := (*c.consumer).Topics()
	if err != nil {
		return nil, err
	}
	for _, topic := range (topicList) {
		if match, err := regexp.Match(tpc, []byte(topic)); err == nil && match {

			partitionList, err := (*c.consumer).Partitions(topic)
			if err != nil {
				return nil, err
			}
			sub := &Subscriber{
				Topic:    topic,
				Consumer: c,
				PcList:   make([]sarama.PartitionConsumer, 0),
			}

			for partition := range partitionList {
				//ConsumePartition方法根据主题，分区和给定的偏移量创建创建了相应的分区消费者
				//如果该分区消费者已经消费了该信息将会返回error
				//sarama.OffsetNewest:表明了为最新消息
				pc, err := (*c.consumer).ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
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
