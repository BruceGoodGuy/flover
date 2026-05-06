package test

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HandlerInt interface {
}

type Handler struct {
	s *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{
		s,
	}
}

func (t *Handler) RetrieveTests(ctx *gin.Context) {

	response := t.s.Index(ctx)
	if !response.Success {
		ctx.JSON(http.StatusBadRequest, Response{Success: false, Message: response.Message})
		return
	}

	ctx.JSON(http.StatusOK, Response{Success: true, Message: "Successfully!", Data: response.Data})
}

func (t *Handler) StoreTest(ctx *gin.Context) {
	var data CreateRequest
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, Response{Success: false, Message: err.Error()})
		return
	}

	response := t.s.Store(ctx, data)

	if !response.Success {
		ctx.JSON(http.StatusBadRequest, Response{Success: false, Message: response.Message})
		return
	}

	ctx.JSON(http.StatusAccepted, Response{Success: true, Message: "Successfully!"})

}
