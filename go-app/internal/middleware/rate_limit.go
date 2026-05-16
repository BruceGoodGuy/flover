package middleware

import (
	"BruceGoodGuy/flover/pkg/cache"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/redis"
)

// RateLimit creates a rate-limiting middleware.
// numberRequest: maximum number of requests allowed.
// period: time window — "S" (second), "M" (minute), "H" (hour), "D" (day).
func RateLimit(numberRequest int, period string) gin.HandlerFunc {
	rate, err := limiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", numberRequest, period))
	if err != nil {
		panic(err)
	}

	client := cache.Connect()
	store, err := redis.NewStore(client)
	if err != nil {
		panic(err)
	}

	instance := limiter.New(store, rate)
	return mgin.NewMiddleware(instance)
}
