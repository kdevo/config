package config

import (
	"fmt"
)

// Errors is a map of field name to Error.
type Errors struct {
	errors map[string]*Error
}

func (e *Errors) Add(err *Error) *Errors {
	if e.errors == nil {
		e.errors = make(map[string]*Error)
	}
	e.errors[err.Field] = err
	return e
}

func (e *Errors) HasField(fieldName string) bool {
	if e.errors == nil {
		return false
	}
	_, ok := e.errors[fieldName]
	return ok
}

func (e *Errors) Error() string {
	var fieldNames []string
	for _, err := range e.errors {
		fieldNames = append(fieldNames, err.Field)
	}
	return fmt.Sprintf("invalid config for fields: %v", fieldNames)
}

func (e *Errors) AsError() error {
	if e == nil || len(e.errors) == 0 {
		return nil
	}
	return e
}

type Error struct {
	Field   string
	Value   interface{}
	Message string
	Inner   error
}

func Err(field string, val interface{}, message string) *Error {
	return &Error{Field: field, Value: val, Message: message}
}

func (e *Error) WithInner(err error) *Error {
	e.Inner = err
	return e
}

func EmptyErr(field string, val string) *Error {
	return &Error{Field: field, Value: val, Message: "must not be empty"}
}

func (e *Error) Error() string {
	return fmt.Sprintf("field %s (with val: %s): %s", e.Field, e.Value, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Inner
}
