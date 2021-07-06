package util

import (
	"fmt"
	"math"
	"time"
)

var (
	minute     = "0"
	hour       = "*"
	dayOfMonth = "*"
	month      = "*"
	dayOfWeek  = "*"
)

func SplitDurationIntoCron(duration time.Duration, desiredIterations int) string {
	minutesPerIteration := int(math.Round(duration.Minutes() / float64(desiredIterations)))

	if minutesPerIteration <= 1 {
		minute = "*"
	} else if minutesPerIteration <= 30 {
		minute = fmt.Sprintf("*/%d", minutesPerIteration)
	} else {
		intervalString := fmt.Sprintf("%dm", minutesPerIteration)
		intervalDuration, _ := time.ParseDuration(intervalString)
		minute = fmt.Sprint(time.Now().Add(intervalDuration).Minute())
	}

	return fmt.Sprintf("%s %s %s %s %s", minute, hour, dayOfMonth, month, dayOfWeek)
}
