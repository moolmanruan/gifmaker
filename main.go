package main

import (
	"fmt"
	"os"

	"github.com/moolmanruan/gifmaker/gif"
)

const version = "v0.0.0"

func main() {
	fmt.Printf("gifmaker %s\n", version)

	input, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	output, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}

	err = gif.Create(string(input), output)
	if err != nil {
		panic(err)
	}
}
