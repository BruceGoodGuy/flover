package app

import (
	"BruceGoodGuy/flover/pkg/mail"

	"BruceGoodGuy/flover/internal/test"
	"BruceGoodGuy/flover/internal/user"

	"BruceGoodGuy/flover/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type AppContainer struct {
	TestHandler *test.Handler
	UserHandler *user.UserHandler
}

func NewAppContainer(db *gorm.DB, rdb *redis.Client, mb *mail.Mail) *AppContainer {
	testRepo := test.NewRepository(db, rdb)

	testService := test.NewService(testRepo)

	userRepo := user.NewUserRepository(db, rdb)

	userService := user.NewUserService(userRepo)

	return &AppContainer{
		TestHandler: test.NewHandler(testService),
		UserHandler: user.NewUserHandler(userService),
	}
}

func (a *AppContainer) Routes() {
	r := gin.Default()

	r.Use(middleware.RateLimit())

	v1 := r.Group("v1")
	{
		v1.GET("/test", a.TestHandler.RetrieveTests)
		v1.POST("/test", a.TestHandler.StoreTest)
	}
	user := v1.Group("user")
	{
		user.POST("create", a.UserHandler.CreateUser)
	}
	r.Run(":8080")
}
