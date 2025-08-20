package webdav

import (
	"errors"
	"fmt"
	"os"
)

var ErrAuthChanged = errors.New("authentication failed, change algorithm")

var ErrTooManyRedirects = errors.New("stopped after 10 redirects")

type StatusError struct {
	Status int
}

func (se StatusError) Error() string {
	return fmt.Sprintf("%d", se.Status)
}

func NewPathError(op string, path string, statusCode int) error {
	return &os.PathError{
		Op:   op,
		Path: path,
		Err:  StatusError{statusCode},
	}
}

func NewPathErrorErr(op string, path string, err error) error {
	return &os.PathError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}
