package main

import (
	"fmt"

	"golang.org/x/example/hello/reverse"
)

func main() {
	baseStr := "Hello, DIASOFT!"
	fmt.Println(reverse.String(baseStr))
}
