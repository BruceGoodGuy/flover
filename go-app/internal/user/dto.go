package user

type CreateRequest struct {
	FirstName  string `json:"first_name" binding:"required,lt=20"`
	LastName   string `json:"last_name" binding:"required,lt=20"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,gt=8,lt=20"`
	RePassword string `json:"re_password" binding:"required,gt=8,lt=20,eqfield=Password"`
}

type EmailRequest struct {
	Email string `json:"email" form:"email" binding:"required,email"`
}

type ConfirmRequest struct {
	Token string `form:"token" json:"token" binding:"len=26"`
}

type UserLogin struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,gt=8,lt=20"`
}

type Tokens struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}
