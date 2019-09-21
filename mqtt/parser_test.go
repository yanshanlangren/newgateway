package mqtt

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	//a := []byte{48, 150, 1}
	//b, len := parseFixedHeader(a)
	//fmt.Println(time.Now().Nanosecond(), b.RemainingLength, len)

	//fmt.Println(1 ^ 1)
	deparser(8555555)
}

func deparser(a int) {
	x := a
	flag := true
	for ; flag; {
		flag = x/128 > 0
		a := x % 128
		if flag {
			a += 128
		}
		fmt.Println(a)
		x /= 128
	}
}
