package user

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type UserRepositoryInt interface {
}

type UserRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewUserRepository(db *gorm.DB, cache *redis.Client) *UserRepository {
	return &UserRepository{
		db:    db,
		cache: cache,
	}
}
