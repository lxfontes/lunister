package main

import (
	"fmt"
)

func main() {
	b, err := NewBridge("clienttap", "trunktap", 10)
	if err != nil {
		fmt.Println(err)
		return
	}

	b.Start()

	for {
		select {}
	}
}
