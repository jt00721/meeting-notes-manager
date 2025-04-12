package usecase

import "errors"

var (
	ErrEmptyTitle   = errors.New("note title cannot be empty")
	ErrEmptyContent = errors.New("note content cannot be empty")
	ErrNoteNotFound = errors.New("note not found")
)
