package main

import (
	"fmt"
)

func main() {
	fmt.Println("start...")
	r := NewRobot("magicsea_1", "111")
	r.Start()
	fmt.Println("end")
}
