package logger

import (
	"fmt"
	"strconv"
	"testing"
)

type A struct {
	str   string
	in    int
	inter interface{}
}

func TestA(t *testing.T) {
	m := make(map[int]*A, 10)
	for i := 0; i < 10; i++ {
		m[i] = &A{
			str:   "asd",
			in:    i,
			inter: "qqq" + strconv.Itoa(i),
		}
	}
	q := m[3]
	for k, v := range (m) {
		if v == q {
			fmt.Println(k)
		}
	}
}
