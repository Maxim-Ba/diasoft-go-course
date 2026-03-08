package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args
	fmt.Printf("args: %v\n", args)
	envs, err := ReadDir("./testdata/env")
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	RunCmd(args, envs)
}
