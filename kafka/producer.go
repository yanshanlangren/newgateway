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
	}
	return &producer
}

func initAsyncProducer() *sarama.AsyncProducer {

	cfg := sarama.NewConfig()
	// 等待服务器所有副本都保存成功后的响应
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	// 随机的分区类型：返回一个分区器，该分区器每次选择一个随机分区
	cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	// 是否等待成功和失败后的响应
	cfg.Producer.Return.Successes = false

	producer, err := sarama.NewAsyncProducer(config.GetConfig().Kafka.ServerList, cfg)
	if err != nil {
		panic(err)
	}
	return &producer
}

var asyncProducer = initAsyncProducer()

func Async(msg *sarama.ProducerMessage) {
	(*asyncProducer).Input() <- msg
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
	return (*kafkaProducer).SendMessages(msgs)
}

var lock = sync.RWMutex{}

var buffer = make(map[string]*Buff)

type Buff struct {
	mu   sync.Mutex
	data []*sarama.ProducerMessage
}

func (b *Buff) Read() []*sarama.ProducerMessage {
	b.mu.Lock()
	defer b.mu.Unlock()
	arr := b.data
	b.data = make([]*sarama.ProducerMessage, 0)
	return arr
}

func (b *Buff) Write(msg *sarama.ProducerMessage) {
	b.mu.Lock()
	b.data = append(b.data, msg)
	b.mu.Unlock()
}

func AsyncPublish(topic string, value string) {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: int32(10),
		Key:       sarama.StringEncoder("key"),
		Value:     sarama.ByteEncoder(value),
	}
	lock.RLock()
	buf, ok := buffer[topic]
	lock.RUnlock()
	if !ok {
		buf = &Buff{
			mu:   sync.Mutex{},
			data: make([]*sarama.ProducerMessage, 0),
		}
		lock.Lock()
		buffer[topic] = buf
		lock.Unlock()
	}
	buf.Write(msg)
}

func Tick() {
	tick := time.Tick(time.Duration(config.GetConfig().Kafka.ProducerTickerInterval) * time.Millisecond)
	for {
		select {
		case <-tick:
			if len(buffer) > 0 {
				lock.RLock()
				for _, v := range (buffer) {
					BatchPublish(v.Read())
				}
				lock.RUnlock()
			}
		}
	}
}
