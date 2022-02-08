// Package main provides the entry point to the program.
package main

import (
	"fmt"
	"github.com/libp2p/go-libp2p"
)

func main()  {
	node, err := libp2p.New()
	if err != nil {
		panic(err)
	}

	fmt.Println()
	if err := node.Close(); err != nil {
		panic(err)
	}
}