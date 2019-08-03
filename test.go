package main

import (
	"fmt"
	"time"
)

func main() {
	arr := []byte{1,2,3,4,5}
	lenarr := len(arr)
	buf := make([]byte, lenarr)
	buf[3] =33

	fmt.Printf("%#v \n", buf)

	time.Sleep(time.Second *1)

}
