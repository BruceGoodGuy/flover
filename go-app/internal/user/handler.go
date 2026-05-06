package user

import (
	"net/http"

	"BruceGoodGuy/flover/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandlerInt interface {
}

type UserHandler struct {
	s *UserService
}

func NewUserHandler(s *UserService) *UserHandler {
	return &UserHandler{
		s,
	}
}

func (h *UserHandler) CreateUser(ctx *gin.Context) {
	var user CreateRequest
	if err := ctx.ShouldBindJSON(&user); err != nil {
		if response.HandleBindError(ctx, err) {
			return
		}
	}

	response.Success(ctx, http.StatusCreated, "ok", user)
}
