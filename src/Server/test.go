package main

import (
	"fmt"
	"time"
)

type data1 struct {
	d []int
	x int
}

func _main() {
	testM := make(map[int][]int)
	testM[1] = make([]int, 0) //[]int{90}
	testM[1] = append(testM[1], 100)
	fmt.Println(testM)

	v0 := make(map[int]int)
	v1 := v0
	v1[1] = 2

	fmt.Println(v0, "   ", v1)

	if v, e := testM[1]; e {
		v = append(v, 110)
	}

	fmt.Println(testM)

	d1 := []int{1}
	ch := make(chan data1, 10)
	d0 := data1{d1, 100}
	ch <- d0
	d2 := <-ch
	d2.d[0] = 3
	d2.x = 200
	fmt.Println(d0, "  ", d2)

	d3 := []int{1}
	chh := make(chan []int, 10)
	chh <- d3
	d4 := <-chh
	d4[0] = 3
	fmt.Println(d3, "  ", d4)

	time.Sleep(1 * time.Second)
}
