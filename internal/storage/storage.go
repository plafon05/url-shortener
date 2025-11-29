package storage

import "errors"

var (
	ErrURLNotFound   = errors.New("url not found")
	ErrAliasNotFound = errors.New("alias not found")
	ErrURLExists     = errors.New("url already exists")
)
