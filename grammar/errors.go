package grammar

import "errors"

var (
	ErrTableRequired   error = errors.New("table name is required")
	ErrColumnsRequired error = errors.New("columns are required")
	ErrKeysRequired    error = errors.New("keys are required")
)
