package middleware

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Timer(c fiber.Ctx) error {
	start := time.Now()
	requestId := "[" + randomHex(start.Nanosecond()+rand.Intn(1000)) + "] "
	c.Locals("X-Request-Id", requestId)

	logger.Info(requestId+"[\x1b[95mSTART\x1b[0m] "+c.Method()+" "+c.OriginalURL(), nil)

	err := c.Next()

	duration := float64(time.Since(start).Microseconds()) / 1000.0
	roundedDuration := math.Round(duration*100) / 100

	if roundedDuration > 1000 {
		logger.Warn(requestId+"[\x1b[95mEND\x1b[0m] Time used: "+fmt.Sprintf("%.2f", roundedDuration/1000)+" s", nil)
	} else {
		logger.Info(requestId+"[\x1b[95mEND\x1b[0m] Time used: "+fmt.Sprintf("%.2f", roundedDuration)+" ms", nil)
	}

	return err
}

func randomHex(number int) string {
	return fmt.Sprintf("%x", number)
}
