package main

import (
	"errors"
	"fmt"
)

var customErr err = errors.New("test")

func test() error {
	return customErr
}
func main() {
	err := test()
	if errors.Is(err, customError) {
		fmt.Println("yay")
	}
}
