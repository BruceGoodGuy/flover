package user

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type UserServiceInt interface {
	CreateUser(ctx context.Context, userData CreateRequest) error
}

type UserService struct {
	r *UserRepository
}

func NewUserService(r *UserRepository) *UserService {
	return &UserService{
		r,
	}
}

func (s *UserService) VerifyEmailExist(ctx context.Context, email string, checkCacheOnly bool) (bool, error) {
	return s.r.CheckEmailExist(ctx, email, checkCacheOnly)
}

func (s *UserService) CreateUser(ctx context.Context, userData CreateRequest) error {
	return s.r.Store(ctx, userData)
}

func (s *UserService) ConfirmAccount(ctx context.Context, token string) (bool, error) {
	return s.r.ConfirmAccount(ctx, token)
}

func (s *UserService) Authenticate(ctx context.Context, userData UserLogin) (Tokens, time.Duration, error) {
	u, err := s.r.FindByEmail(ctx, userData.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Tokens{}, 0, ErrInvalidCredentials
		}
		return Tokens{}, 0, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(userData.Password)); err != nil {
		return Tokens{}, 0, ErrInvalidCredentials
	}

	token, err := s.populateToken(u)

	fmt.Print(token.AccessToken)

	if err != nil {
		return Tokens{}, 0, err
	}

	ttl := (24 * 7 * time.Hour)

	s.r.StoreRefreshToken(ctx, token.RefreshToken, ttl, userData.Email)

	return token, ttl, err
}

func (s *UserService) populateToken(u User) (Tokens, error) {
	now := time.Now()
	role := u.Role
	if role == "" {
		role = "human"
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   u.Id,
		"email": u.Email,
		"role":  role,
		"iat":   now.Unix(),
		"exp":   now.Add(10 * time.Minute).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(os.Getenv("SECRET_STRING")))
	if err != nil {
		fmt.Printf("[PopulateToken] Can't create access token: %v\n", err)
		return Tokens{}, err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   u.Id,
		"email": u.Email,
		"role":  role,
		"iat":   now.Unix(),
		"exp":   now.Add(7 * 24 * time.Hour).Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(os.Getenv("SECRET_REFRESH_STRING")))
	if err != nil {
		fmt.Printf("[PopulateToken] Can't create refresh token: %v\n", err)
		return Tokens{}, err
	}

	return Tokens{RefreshToken: refreshTokenString, AccessToken: accessTokenString}, nil
}
