package test

type CreateRequest struct {
	Name  string `json:"name" binding:"required,gt=2,lt=20"`
	Value string `json:"value" binding:"required,gt=2,lt=20"`
}

type Response struct {
	Success bool
	Message string
	Data    any
}
