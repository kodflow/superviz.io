package services

import "errors"

var (
	ErrInvalidTarget = errors.New("invalid target format, expected user@host")
	ErrNilConfig     = errors.New("config cannot be nil")
	ErrNilWriter     = errors.New("writer cannot be nil")
)
