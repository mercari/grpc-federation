package resolver

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mercari/grpc-federation/source"
)

// Error represents resolver package's error.
// this has multiple error instances from resolver.
type Error struct {
	Errs []error
}

func (e *Error) Error() string {
	errs := make([]string, 0, len(e.Errs))
	for _, err := range e.Errs {
		errs = append(errs, err.Error())
	}
	return fmt.Sprintf("grpc-federation: %s", strings.Join(errs, ": "))
}

// LocationError holds error message with location.
type LocationError struct {
	Location *source.Location
	Message  string
}

func (e *LocationError) Error() string {
	return e.Message
}

// ExtractIndividualErrors extracts all error instances from Error type.
func ExtractIndividualErrors(err error) []error {
	var e *Error
	if !errors.As(err, &e) {
		return nil
	}
	if e == nil {
		return nil
	}
	return e.Errs
}

// ToLocationError convert err into LocationError if error instances the one.
func ToLocationError(err error) *LocationError {
	var e *LocationError
	if !errors.As(err, &e) {
		return nil
	}
	return e
}

// ErrWithLocation creates LocationError instance from message and location.
func ErrWithLocation(msg string, loc *source.Location) *LocationError {
	return &LocationError{
		Location: loc,
		Message:  msg,
	}
}

type errorBuilder struct {
	errs []error
}

func (b *errorBuilder) add(err error) {
	if err == nil {
		return
	}
	b.errs = append(b.errs, err)
}

func (b *errorBuilder) build() error {
	if len(b.errs) == 0 {
		return nil
	}
	return &Error{Errs: b.errs}
}
