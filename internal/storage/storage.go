package storage

import "errors"

var (
	ErrURLNotFound   = errors.New("URL не найден")
	ErrAliasNotFound = errors.New("алиас не найден")
	ErrURLExists     = errors.New("URL уже существует")
)
