package models

import "errors"

var (
	ErrForeignKeyViolation = errors.New("violates foreign key constraint")
)
