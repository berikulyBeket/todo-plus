package utils

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserNotOwner     = errors.New("user is not the owner")
	ErrListNotFound     = errors.New("list not found")
	ErrListEmptyRequest = errors.New("list update structure has no values")
	ErrItemNotFound     = errors.New("item not found")
	ErrItemEmptyRequest = errors.New("item update structure has no values")
)
