package pagination

import "errors"

var (
	ErrorMaxPage     = errors.New("page number exceeds maximum page limit")
	ErrorPage        = errors.New("page must be greater than zero")
	ErrorPageEmpty   = errors.New("page cannot be empty")
	ErrorPageInvalid = errors.New("page is invalid, must be a number")
)
