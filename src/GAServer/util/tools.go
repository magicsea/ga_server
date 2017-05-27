package util

import "time"

type TimeTaskFunc func()

func StartLoopTask(t time.Duration, fun TimeTaskFunc) *time.Ticker {
	timeTicker := time.NewTicker(t)
	go func() {
		for {
			select {
			case <-timeTicker.C:
				fun()
			}
		}
	}()
	return timeTicker
}
