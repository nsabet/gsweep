package main

import (
	"time"
)

const (
	millisPerSecond     = int64(time.Second / time.Millisecond)
	nanosPerMillisecond = int64(time.Millisecond / time.Nanosecond)
)

func msToTime(msInt int64) time.Time {
	return time.Unix(msInt/millisPerSecond, (msInt%millisPerSecond)*nanosPerMillisecond)
}
