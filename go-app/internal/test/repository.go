package test

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type RepositoryInt interface {
	Index(ctx context.Context) ([]Test, error)
	Store(ctx context.Context, data Test) error
}

type Repository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewRepository(db *gorm.DB, cache *redis.Client) *Repository {
	return &Repository{
		db,
		cache,
	}
}

func (t *Repository) Index(ctx context.Context) (error, []Test) {
	var tests []Test
	if error := t.db.WithContext(ctx).Find(&tests).Error; error != nil {
		return error, tests
	}

	return nil, tests
}

func (t *Repository) Store(ctx context.Context, data Test) error {
	if err := t.db.WithContext(ctx).Create(&data).Error; err != nil {
		return err
	}

	return nil
}
