package app

import (
	"BruceGoodGuy/flover/pkg/mail"
	"context"

	"BruceGoodGuy/flover/internal/test"
	"BruceGoodGuy/flover/internal/user"

	"BruceGoodGuy/flover/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AppContainer struct {
	TestHandler *test.Handler
	UserHandler *user.UserHandler
	userRepo    *user.UserRepository
}

func NewAppContainer(db *gorm.DB, rdb *redis.Client, mb *mail.Mail) *AppContainer {
	testRepo := test.NewRepository(db, rdb)

	testService := test.NewService(testRepo)

	userRepo := user.NewUserRepository(db, rdb, mb)

	userService := user.NewUserService(userRepo)

	return &AppContainer{
		TestHandler: test.NewHandler(testService),
		UserHandler: user.NewUserHandler(userService),
		userRepo:    userRepo,
	}
}

func (a *AppContainer) Bootstrap() {
	ctx := context.Background()
	go a.userRepo.SeedBloomFilter(ctx)
}

func (a *AppContainer) Routes() {
	r := gin.Default()

	r.Use(middleware.RateLimit(100, "m"))

	v1 := r.Group("v1")
	{
		v1.GET("/test", a.TestHandler.RetrieveTests)
		v1.POST("/test", a.TestHandler.StoreTest)
	}
	user := v1.Group("user")
	{
		user.POST("/create", a.UserHandler.CreateUser)
		user.GET("/verify", middleware.RateLimit(5, "m"), a.UserHandler.VerifyEmailExist)
		user.GET("/confirm", middleware.RateLimit(100, "m"), a.UserHandler.ConfirmAccount)
		user.POST("/authenticate", a.UserHandler.Authenticate)
	}
	r.Run(":8080")
}
