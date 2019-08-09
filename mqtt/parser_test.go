package mqtt

import (
	"fmt"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	a := []byte{48, 16, 0, 5, 116, 101, 115, 116, 52, 104, 101, 108, 108, 111, 45, 49, 51, 56, 16, 0, 5, 116, 101, 115, 116, 52, 104, 101, 108, 108, 111, 45, 49, 53, 51}
	b := ParseMQTTMessage(a)
	fmt.Println(time.Now().Nanosecond(), b)
}
