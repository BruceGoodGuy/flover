package test

import (
	"context"
)

type ServiceInt interface {
	Index(ctx context.Context) any
	Store(ctx context.Context, data CreateRequest) any
}

type Service struct {
	r *Repository
}

func NewService(r *Repository) *Service {
	return &Service{
		r,
	}
}

func (s *Service) Index(ctx context.Context) Response {
	err, data := s.r.Index(ctx)
	if err != nil {
		return Response{Success: false, Message: err.Error()}
	}

	return Response{Success: true, Data: data}
}

func (s *Service) Store(ctx context.Context, data CreateRequest) Response {
	testData := Test{
		Name:  data.Name,
		Value: data.Value,
	}
	err := s.r.Store(ctx, testData)

	if err != nil {
		return Response{Success: false, Message: err.Error()}
	}

	return Response{Success: true, Data: data}
}
