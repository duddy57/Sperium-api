package users

import "errors"

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrCredentialsNotFound = errors.New("email or password is incorrect")
)
