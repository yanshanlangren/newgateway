package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"testing"
	"time"
)

func TestProducer(t *testing.T) {
	var msgs []*sarama.ProducerMessage
	for i := 0; i < 500000; i++ {
		msgs = append(msgs, &sarama.ProducerMessage{
			Topic:     "test",
			Partition: int32(10),
			Key:       sarama.StringEncoder("key"),
			Value:     sarama.ByteEncoder("hello"),
		})
	}
	a := time.Now().UnixNano()
	for _, v := range (msgs) {
		Async(v)
	}
	//BatchPublish(msgs)
	b := time.Now().UnixNano()
	newConsumer()
	fmt.Printf("%v ns costs", b-a)
}
