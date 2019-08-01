package kafka

import (
	"github.com/Shopify/sarama"
	"newgateway/config"
	"newgateway/logger"
	"sync"
	"time"
)

var kafkaProducer = initProducer()

func init() {
	go Tick()
}

func initProducer() *sarama.SyncProducer {
	cfg := sarama.NewConfig()
	// 等待服务器所有副本都保存成功后的响应
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	// 随机的分区类型：返回一个分区器，该分区器每次选择一个随机分区
	cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	// 是否等待成功和失败后的响应
	cfg.Producer.Return.Successes = true

	// 使用给定代理地址和配置创建一个同步生产者
	producer, err := sarama.NewSyncProducer(config.GetConfig().Kafka.ServerList, cfg)
	if err != nil {
		logger.Fatal(err.Error())
		panic(err)
	}
	return &producer
}

//返回partition, offset, error
func Publish(topic, value string) (int32, int64, error) {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: int32(10),
		Key:       sarama.StringEncoder("key"),
		Value:     sarama.ByteEncoder(value),
	}
	return (*kafkaProducer).SendMessage(msg)
}

func BatchPublish(msgs []*sarama.ProducerMessage) error {
	//msgs := make([]*sarama.ProducerMessage, 0)
	//for _, v := range (value) {
	//	msgs = append(msgs, &sarama.ProducerMessage{
	//		Topic:     topic,
	//		Partition: int32(10),
	//		Key:       sarama.StringEncoder("key"),
	//		Value:     sarama.ByteEncoder(v),
	//	})
	//}
	return (*kafkaProducer).SendMessages(msgs)
}

var buffer = make(map[string][]*sarama.ProducerMessage)

func AsyncPublish(topic string, value string) {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: int32(10),
		Key:       sarama.StringEncoder("key"),
		Value:     sarama.ByteEncoder(value),
	}
	mu.Lock()
	buffer[topic] = append(buffer[topic], msg)
	mu.Unlock()
}

var mu = sync.Mutex{}

func Tick() {
	tick := time.Tick(time.Duration(100) * time.Millisecond)
	for {
		select {
		case <-tick:
			mu.Lock()
			for _, v := range (buffer) {
				BatchPublish(v)
			}
			buffer = make(map[string][]*sarama.ProducerMessage)
			mu.Unlock()
		}
	}
}
