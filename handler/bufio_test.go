package handler

import (
	"fmt"
	"strings"
	"testing"
)

func TestBufio(t *testing.T) {
	//var a s



	var quote string
	var name string

	fmt.Print("What is the quote? ")
	fmt.Scanf("%s", &quote)

	fmt.Print("Who said it? ")
	fmt.Scanf("%s", &name)

	fmt.Printf("%s says, \"%s\"", strings.Title(name), quote)
}
