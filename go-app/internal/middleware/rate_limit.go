package middleware

import (
	"BruceGoodGuy/flover/pkg/cache"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/redis"
)

func RateLimit() gin.HandlerFunc {
	// Allow maximum 100 reqs per second.
	rate, err := limiter.NewRateFromFormatted("100-s")
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
