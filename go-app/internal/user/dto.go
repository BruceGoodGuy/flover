package user

type CreateRequest struct {
	FirstName  string `json:"first_name" binding:"required,lt=20"`
	LastName   string `json:"last_name" binding:"required,lt=20"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,gt=8,lt=20"`
	RePassword string `json:"re_password" binding:"required,gt=8,lt=20,eqfield=Password"`
}
