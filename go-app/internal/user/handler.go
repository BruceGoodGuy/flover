package user

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"BruceGoodGuy/flover/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandlerInt interface {
	VerifyEmailExist(ctx *gin.Context)
	CreateUser(ctx *gin.Context)
	ConfirmAccount(ctx *gin.Context)
}

type UserHandler struct {
	s *UserService
}

func NewUserHandler(s *UserService) *UserHandler {
	return &UserHandler{
		s,
	}
}

func (h *UserHandler) VerifyEmailExist(ctx *gin.Context) {
	var email EmailRequest
	if err := ctx.ShouldBindQuery(&email); err != nil {
		if response.HandleBindError(ctx, err) {
			return
		}
	}
	if isExist, err := h.s.VerifyEmailExist(ctx, strings.ToLower(email.Email), false); err == nil {
		if isExist {
			ctx.JSON(http.StatusUnprocessableEntity, response.Response{IsSuccess: false, Message: "Duplicate email"})
			return
		}
		ctx.JSON(http.StatusOK, response.Response{IsSuccess: true, Message: "Email Valid"})
	} else {
		ctx.JSON(http.StatusUnprocessableEntity, response.Response{IsSuccess: false, Message: "Something went wrong! Please try again later"})
	}

}

func (h *UserHandler) CreateUser(ctx *gin.Context) {
	var user CreateRequest
	if err := ctx.ShouldBindJSON(&user); err != nil {
		if response.HandleBindError(ctx, err) {
			return
		}
	}

	if isExist, err := h.s.VerifyEmailExist(ctx, user.Email, false); err == nil {
		if isExist {
			ctx.JSON(http.StatusUnprocessableEntity, response.Response{IsSuccess: false, Message: "Duplicate email"})
			return
		}
	} else {
		ctx.JSON(http.StatusUnprocessableEntity, response.Response{IsSuccess: false, Message: "Something went wrong! Please try again later"})
		return
	}

	if err := h.s.CreateUser(ctx, user); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Response{
			IsSuccess: false,
			Message:   "Failed to create user",
			Data:      err.Error(),
		})
		return
	}

	response.Success(ctx, http.StatusCreated, "ok", user)
}

func (h *UserHandler) ConfirmAccount(ctx *gin.Context) {
	var token ConfirmRequest
	if err := ctx.ShouldBindQuery(&token); err != nil {
		if response.HandleBindError(ctx, err) {
			return
		}
	}
	result, _ := h.s.ConfirmAccount(ctx, token.Token)
	if !result {
		ctx.JSON(http.StatusRequestTimeout, response.Response{IsSuccess: false, Message: "Expire token"})
		return
	}

	ctx.JSON(http.StatusCreated, response.Response{
		IsSuccess: true,
		Message:   "Create successfully",
	})

	// ctx.Redirect(http.StatusMovedPermanently, "http://www.google.com/")
}

func (h *UserHandler) Authenticate(ctx *gin.Context) {
	var userData UserLogin
	if err := ctx.ShouldBindJSON(&userData); err != nil {
		if response.HandleBindError(ctx, err) {
			return
		}
	}

	userData.Email = strings.ToLower(userData.Email)
	tokens, ttl, err := h.s.Authenticate(ctx, userData)

	if errors.Is(err, ErrInvalidCredentials) {
		ctx.JSON(http.StatusNotFound, response.Response{IsSuccess: false, Message: "Invalid user data"})
		return
	}
	if err != nil {
		fmt.Printf("%s", err.Error())
		ctx.JSON(http.StatusInternalServerError, response.Response{IsSuccess: false, Message: "Something went wrong! Please try again later"})
		return
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	ctx.JSON(http.StatusOK, response.Response{
		IsSuccess: true,
		Message:   "Ok",
		Data: AuthResponse{
			AccessToken: tokens.AccessToken,
		},
	})
}
