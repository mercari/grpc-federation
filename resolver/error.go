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

func (e *Error) add(err error) {
	if err == nil {
		return
	}

	e.Errs = append(e.Errs, err)
}

// LocationError holds error message with location.
type LocationError interface {
	error
	GetLocation() *source.Location
	GetMessage() string
}

type LocationErrorImpl struct {
	Location *source.Location
	Message  string
}

func (e *LocationErrorImpl) GetLocation() *source.Location {
	return e.Location
}

func (e *LocationErrorImpl) GetMessage() string {
	return e.Message
}

func (e *LocationErrorImpl) Error() string {
	return e.Message
}

type CyclicDependencyError struct {
	Location *source.Location
	Message  string
}

func (e *CyclicDependencyError) GetLocation() *source.Location {
	return e.Location
}

func (e *CyclicDependencyError) GetMessage() string {
	return e.Message
}

func (e *CyclicDependencyError) Error() string {
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
func ToLocationError(err error) LocationError {
	var e LocationError
	if !errors.As(err, &e) {
		return nil
	}
	return e
}

// ErrWithLocation creates LocationError instance from message and location.
func ErrWithLocation(msg string, loc *source.Location) LocationError {
	return &LocationErrorImpl{
		Location: loc,
		Message:  msg,
	}
}
