package main

import (
	"fmt"
	"time"
)

// 消息接收计时器
type timer struct {
	id       int64
	start    time.Time
	end      time.Time
	interval time.Duration
}

func newTimer(interval time.Duration) *timer {
	return &timer{interval: interval}
}

func (t *timer) run() {
	fmt.Println("开始计时")
	t.start = time.Now()
	t.end = t.start
	t.id = t.start.Unix()
	go func() {
		ticker := time.NewTicker(t.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fmt.Println(t.id, " 时差(毫秒):", t.end.Sub(t.start).Milliseconds())
			}
		}
	}()
}

// 更新 end 时间
func (t *timer) updateEndTime() {
	t.end = time.Now()
}
