package user

import (
	"BruceGoodGuy/flover/pkg/mail"
	"context"
	"fmt"
	"os"
	"time"

	"encoding/json"

	"crypto/rand"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepositoryInt interface {
	Store(ctx context.Context, user CreateRequest) error
}

type UserRepository struct {
	db    *gorm.DB
	cache *redis.Client
	mb    *mail.Mail
}

func NewUserRepository(db *gorm.DB, cache *redis.Client, mb *mail.Mail) *UserRepository {
	return &UserRepository{
		db:    db,
		cache: cache,
		mb:    mb,
	}
}

// Store creates user and updates Bloom Filter if insert is successful
func (r *UserRepository) Store(ctx context.Context, req CreateRequest) error {
	u := User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		fmt.Printf("[HashPassword] Failed to hash password")
		fmt.Printf("Error: %s", err.Error())
		return err
	}

	u.Password = string(hashPassword)

	jsonData, err := json.Marshal(&u)

	if err != nil {
		fmt.Printf("[Jsonconvert] Failed to convert to json")
		fmt.Printf("Error: %s", err.Error())
		return err
	}

	key := rand.Text()

	r.cache.Set(ctx, "register:token:"+key, string(jsonData), 5*time.Minute)

	emailData := map[string]interface{}{
		"name":            u.FirstName,
		"activation_link": os.Getenv("APP_HOST") + "/user/verify?token=" + key,
	}

	mailContext := context.Background()

	go r.mb.Send(mailContext, u.Email, "Welcome to Flover! Confirm your email", "verify", emailData)
	return nil
}

// InitBloomFilter should be called once when application starts
func (r *UserRepository) InitBloomFilter(ctx context.Context) error {
	exists, err := r.cache.Exists(ctx, "users:email").Result()
	if err != nil {
		return fmt.Errorf("redis check exists error: %w", err)
	}

	if exists != 1 {
		_, err := r.cache.BFReserve(ctx, "users:email", 0.01, 10000).Result()
		if err != nil {
			return fmt.Errorf("redis bfreserve error: %w", err)
		}
		fmt.Println("[BloomFilter] Created new filter 'users:email'")
	}
	return nil
}

// CheckEmailExist is purely for API live-check (e.g. typing email on register form)
func (r *UserRepository) CheckEmailExist(ctx context.Context, email string) (bool, error) {
	exists, err := r.cache.BFExists(ctx, "users:email", email).Result()
	if err != nil {
		fmt.Printf("[BloomFilter] Check error, fallback to DB: %v\n", err)
		return r.checkEmailExistInDB(ctx, email)
	}

	if !exists {
		fmt.Printf("[BloomFilter] Email %s definitely not exist\n", email)
		return false, nil
	}

	fmt.Printf("[BloomFilter] Email %s maybe exist, checking DB...\n", email)
	return r.checkEmailExistInDB(ctx, email)
}

func (r *UserRepository) checkEmailExistInDB(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
