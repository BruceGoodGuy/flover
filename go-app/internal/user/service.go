package user

import "context"

type UserServiceInt interface {
}

type UserService struct {
	r *UserRepository
}

func NewUserService(r *UserRepository) *UserService {
	return &UserService{
		r,
	}
}

func CreateUser(ctx *context.Context, userData CreateRequest) {

}
