package users

import (
	"context"
	"time"
	"uuid"
)

type Users struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email'"`
	Password string    `json:"-,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type CreateUserReq struct {
	Name     string `json:"name" validate:"required,min=3,max=32"`
	Email    string `json:"email" validate:"required,email,min=3,max=32"`
	Password string `json:"password" validate:"required,min=8,max=32"`
}

type LoginUserReq struct {
	Email    string `json:"email" validate:"required,email,min=3,max=32"`
	Password string `json:"password" validate:"required,min=8,max=32"`
}

type UpdateUserReq struct {
	Name string `json:"name" validate:"required,min=3,max=32"`
}

type userService interface {
	Create(context.Context, CreateUserReq) (uuid.UUID, error)
	Login(context.Context, LoginUserReq) error
	Logout(context.Context) error
	Details(context.Context) (Users, error)
	Delete(context.Context) error
	Update(context.Context, UpdateUserReq) error
}

type createUserResp struct {
	Message string    `json:"message"`
	ID      uuid.UUID `json:"id"`
}

type genericResp struct {
	Message string `json:"message"`
}
