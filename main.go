package main

import (
	"fmt"
	"os"

	"github.com/moolmanruan/gifmaker/gif"
)

const version = "v0.1.1"

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("gifmaker (%s)\n\nUsage: gifmaker <input> <output>\n", version)
		return
	}

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
