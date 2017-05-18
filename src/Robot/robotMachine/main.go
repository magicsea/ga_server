package main

import (
	. "Robot/agent"
	"fmt"
	_ "sync"
	"time"
)

//var wg sync.WaitGroup

func main() {
	var actionTime time.Duration = time.Millisecond * 0
	var robotCount = 500
	var taskCount = 1000
	var finishCount = 0
	var errorCount = 0
	ch := make(chan string, 100)

	fmt.Println("start...count:", robotCount)
	now := time.Now().UnixNano()
	//wg.Add(robotCount)
	for i := 0; i < robotCount; i++ {
		go Run(i, taskCount, actionTime, ch)
	}
	for r := range ch {
		if r == "OK" {
			finishCount++
		} else {
			errorCount++
		}

		fmt.Println("r:", r, " finish:", finishCount, "/", errorCount, "/", robotCount)
		if finishCount+errorCount == robotCount {
			break
		}
	}
	//wg.Wait()
	use := time.Now().UnixNano() - now
	ms := use / int64(time.Millisecond)
	qps := float32(taskCount*finishCount) / float32(ms) * 1000
	fmt.Println("end. robotCount=", robotCount, "time=ms", ms, "all_qps=", qps)
}

func Run(index, taskCount int, actionTime time.Duration, ch chan string) {
	now := time.Now().UnixNano()
	acc := fmt.Sprintf("magicsea_%d", index)
	r := NewRobot(acc, "111", actionTime)
	result := r.Start(taskCount)
	use := time.Now().UnixNano() - now
	ms := use / int64(time.Millisecond)
	qps := float32(taskCount) / float32(ms) * 1000
	fmt.Println("task over=>", index, " result:", result, taskCount, "  usetime:ms", use/int64(time.Millisecond), " qps:", qps)
	ch <- result
	//wg.Done()
}
