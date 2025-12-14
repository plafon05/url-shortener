package storage

import "errors"

var (
	ErrURLNotFound   = errors.New("URL не найден")
	ErrAliasNotFound = errors.New("alias не найден")
	ErrAliasExists   = errors.New("alias уже существует")
)
