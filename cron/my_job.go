package cron

import (
	"fmt"
	"time"
)

type MyJob struct{}

func (m MyJob) Run() {
	fmt.Println("Custom job running at 123123", time.Now())
}
