package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrNotFound           = errors.New("user not found")
	ErrInvalidPattern     = errors.New("invalid pattern")
	ErrNoTicketsAvailable = errors.New("no tickets available")
)
