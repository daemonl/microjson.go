package microjson

import (
	"encoding/json"
	"net/http"
	"time"
)

// JSONBodyError attempts to safely extract JSON error details for the caller
func JSONBodyError(err error) error {

	switch err := err.(type) {
	case *json.UnmarshalFieldError, *json.UnmarshalTypeError:
		return WrapError(err).
			Status(http.StatusBadRequest).
			Message(ErrorMessageSchemaViolation)
	case *json.SyntaxError:
		return WrapError(err).
			Status(http.StatusBadRequest).
			Message(ErrorMessageInvalidJSON)
	case *time.ParseError:
		return WrapError(err).
			Status(http.StatusBadRequest).
			Message(ErrorMessageSchemaViolation)
	}

	return WrapError(err)
}
