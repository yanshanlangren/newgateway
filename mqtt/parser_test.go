package mqtt

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	q := []byte{48,16,0,5,116,101,115,116,52,104,101,108,108,111,45,51,52,49,16,0,5,116,101,115,116,52,104,101,108,108,111,45,51,55,48,48,16,0,5,116,101,115,116,52,104,101,108,108,111,45,50,50,56}
	asd, err := ParseMQTTMessage(q)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(asd)
}
