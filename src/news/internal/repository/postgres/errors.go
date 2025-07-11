package postgres

import "errors"

var (
	ErrFailedToCreateNews          = errors.New("failed to create news")
	ErrFailedToCreateContentBlock  = errors.New("failed to create content block")
	ErrNoTransactionInContext      = errors.New("no transaction found in context")
	ErrFailedToGetNews             = errors.New("failed to get news")
	ErrFailedToGetContentBlocks    = errors.New("failed to get content blocks")
	ErrFailedToUpdateNews          = errors.New("failed to update news")
	ErrFailedToDeleteContentBlocks = errors.New("failed to delete content blocks")
	ErrNotFound                    = errors.New("not found")
	ErrFailedToDeleteNews          = errors.New("failed to delete news")
)
