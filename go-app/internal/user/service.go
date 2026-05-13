package user

import "context"

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

func (s *UserService) VerifyEmailExist(ctx context.Context, email string) (bool, error) {
	return s.r.CheckEmailExist(ctx, email)
}

func (s *UserService) CreateUser(ctx context.Context, userData CreateRequest) error {
	return s.r.Store(ctx, userData)
}

func (s *UserService) ConfirmAccount(ctx context.Context, token string) (bool, error) {
	return s.r.ConfirmAccount(ctx, token)
}
