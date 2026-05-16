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

func (r *UserRepository) SeedBloomFilter(ctx context.Context) {
	const batchSize = 1000

	exists, err := r.cache.Exists(ctx, "users:email").Result()
	if err != nil {
		fmt.Printf("[SeedBloomFilter] Redis error: %v\n", err)
		return
	}
	if exists > 0 {
		fmt.Println("[SeedBloomFilter] Bloom filter already exists, skipping seed")
		return
	}

	type emailRow struct {
		ID    uint
		Email string
	}

	var lastID uint = 0
	for {
		var rows []emailRow
		err := r.db.WithContext(ctx).
			Model(&User{}).
			Select("id, email").
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(batchSize).
			Scan(&rows).Error
		if err != nil {
			fmt.Printf("[SeedBloomFilter] DB error: %v\n", err)
			break
		}
		if len(rows) == 0 {
			break
		}

		pipe := r.cache.Pipeline()
		for _, row := range rows {
			pipe.Do(ctx, "BF.ADD", "users:email", row.Email)
		}
		if _, err := pipe.Exec(ctx); err != nil {
			fmt.Printf("[SeedBloomFilter] Redis pipeline error: %v\n", err)
		}

		lastID = rows[len(rows)-1].ID
		fmt.Printf("[SeedBloomFilter] Seeded %d emails, last ID: %d\n", len(rows), lastID)
	}
	fmt.Println("[SeedBloomFilter] Done")
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

	ttl := 5 * time.Minute

	// Lock the email to prevent duplicate pending registrations.
	r.cache.Set(ctx, "register:email:"+u.Email, key, ttl)
	r.cache.Set(ctx, "register:token:"+key, string(jsonData), ttl)

	emailData := map[string]interface{}{
		"name":            u.FirstName,
		"activation_link": os.Getenv("APP_HOST") + "/user/confirm?token=" + key,
	}

	mailContext := context.Background()

	go r.mb.Send(mailContext, u.Email, "Welcome to Flover! Confirm your email", "verify", emailData)
	return nil
}

// CheckEmailExist is purely for API live-check (e.g. typing email on register form)
func (r *UserRepository) CheckEmailExist(ctx context.Context, email string, checkCacheOnly bool) (bool, error) {
	val, err := r.cache.Exists(ctx, "register:email:"+email).Result()
	if err == nil && val > 0 {
		return true, nil
	}

	exists, err := r.cache.BFExists(ctx, "users:email", email).Result()
	if err != nil {
		fmt.Printf("[BloomFilter] Check error, fallback to DB: %v\n", err)
		if checkCacheOnly {
			return false, nil
		}
		return r.checkEmailExistInDB(ctx, email)
	}

	if !exists {
		fmt.Printf("[BloomFilter] Email %s definitely not exist\n", email)
		return false, nil
	}

	if checkCacheOnly {
		return true, nil
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

func (r *UserRepository) ConfirmAccount(ctx context.Context, token string) (bool, error) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[ActivateAccount] DB error, Can't activate account: %v", r)
		}
	}()

	data, err := r.cache.Get(ctx, "register:token:"+token).Result()

	if err == redis.Nil {
		fmt.Printf("[ActivateAccount] Token not found")
		return false, err
	} else if err != nil {
		fmt.Printf("[ActivateAccount] Check error, Can't active account")
		return false, err
	}
	var userData User

	err = json.Unmarshal([]byte(data), &userData)

	if err != nil {
		fmt.Printf("[ActivateAccount] Can't decode user data")
		return false, err
	}

	fmt.Printf("[ActivateAccount] Decoded user: %s\n", userData.Email)

	userData.Status = "active"

	if err := r.db.WithContext(ctx).Create(&userData).Error; err != nil {
		fmt.Printf("[ActivateAccount] DB error: %v\n", err)
		return false, err
	}

	if ok, err := r.cache.BFAdd(ctx, "users:email", userData.Email).Result(); err != nil {
		fmt.Printf("[ActivateAccount] Can't add %s to bloom filter: %v\n", userData.Email, err)
	} else if !ok {
		fmt.Printf("[ActivateAccount] Email %s already exists in bloom filter\n", userData.Email)
	}

	if _, err := r.cache.Del(ctx, "register:token:"+token, "register:email:"+userData.Email).Result(); err != nil {
		fmt.Printf("[ActivateAccount] Can't remove cache keys\n")
	}

	return true, nil
}

func (r *UserRepository) Authenticate(ctx context.Context, userData UserLogin) (bool, error) {
	type a struct {
		ID       int
		Email    string
		Password string
	}

	var data a
	if err := r.db.WithContext(ctx).Model(&User{}).Where("email = ?", userData.Email).First(&data).Error; err != nil {
		return false, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(data.Password), []byte(userData.Password)); err != nil {
		return false, err
	}

	return true, nil
}
