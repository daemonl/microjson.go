// Package apiutil provides utility functions for building a JSON api
package microjson

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HTTPError interface {
	error
	HTTPStatus() int
	ResponseBody() interface{}
}

const (
	ErrorMessageGeneric         = "Unknown Server Error"
	ErrorMessageBadRequest      = "Bad Request"
	ErrorMessageInvalidJSON     = "Invalid JSON"
	ErrorMessageSchemaViolation = "Request did not match schema"
)

func SendError(rw http.ResponseWriter, req *http.Request, err error) {
	httpError, ok := err.(HTTPError)
	if !ok {
		RequestLogEntry(req).Errorf("Unhandled Error: %s", err.Error())
		httpError = WrapError(err).Message("Unknown Error")
	}
	status := httpError.HTTPStatus()
	SendObject(rw, req, status, httpError.ResponseBody())
	return
}

func SendObject(rw http.ResponseWriter, req *http.Request, code int, body interface{}) {
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(code)
	json.NewEncoder(rw).Encode(body)
	return
}

type builtError struct {
	message    string
	err        error
	status     int
	tags       []string
	customData map[string]interface{}
	plainText  bool
}

func (err builtError) WrappedError() error {
	return err.err
}

func (err builtError) Error() string {
	if err.err != nil {
		return err.err.Error()
	}
	return err.message
}

func (err builtError) HTTPStatus() int {
	return err.status
}

func (err builtError) ResponseBody() interface{} {
	return map[string]interface{}{
		"status":  http.StatusText(err.status),
		"message": err.message,
	}
}

func (err *builtError) Tags() []string {
	return err.tags
}

func (err *builtError) Tag(tag ...string) Builder {
	err.tags = append(err.tags, tag...)
	return err
}

func (err *builtError) Message(format string, params ...interface{}) Builder {
	err.message = fmt.Sprintf(format, params...)
	return err
}

func (err *builtError) Status(status int) Builder {
	err.status = status
	if err.message == ErrorMessageGeneric && status < 500 {
		err.message = http.StatusText(status)
	}
	return err
}

func (err *builtError) Wrap(causingError error) Builder {
	err.err = causingError
	return err
}

func (err *builtError) PlainText(format string, params ...interface{}) Builder {
	err.plainText = true
	err.Message(format, params...)
	return err
}

func (err *builtError) CustomField(key string, value interface{}) Builder {
	err.customData[key] = value
	return err
}

func (err *builtError) GetCustomFields() map[string]interface{} {
	return err.customData
}

// Builder builds a Tagged HTTPError
type Builder interface {
	HTTPError
	Tag(...string) Builder
	Message(string, ...interface{}) Builder
	PlainText(string, ...interface{}) Builder
	Status(int) Builder
	Wrap(error) Builder
	CustomField(string, interface{}) Builder
}

// New returns a new empty Builder
func NewError() Builder {

	return &builtError{
		message:    ErrorMessageGeneric,
		err:        nil,
		status:     http.StatusInternalServerError,
		tags:       []string{},
		customData: map[string]interface{}{},
	}
}

// Wrap wraps an existing error with a builder
func WrapError(err error) Builder {
	return NewError().Wrap(err)
}
