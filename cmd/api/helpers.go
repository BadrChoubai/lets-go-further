package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"strconv"
)

// Retrieve the "id" URL parameter from the current request context, then convert it to
// an integer and return it. If the operation isn't successful, return 0 and an error.
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// Define an envelope type
type envelope map[string]any

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	JSON, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	JSON = append(JSON, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(JSON)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	err := json.NewDecoder(r.Body).Decode(dst)

	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	var invalidUnmarshalError *json.InvalidUnmarshalError

	switch {
	case errors.As(err, &syntaxError):
		return fmt.Errorf(
			"JSON Body: contains malformed JSON (at character %d)",
			syntaxError.Offset,
		)

	case errors.As(err, &unmarshalTypeError):
		if unmarshalTypeError.Field != "" {
			return fmt.Errorf(
				"JSON  Body: contains incorrect JSON type (T: %q) for field %q",
				unmarshalTypeError.Type, unmarshalTypeError.Field,
			)
		}
		return fmt.Errorf(
			"JSON Body: contains incorrect JSON type (at character %q)",
			unmarshalTypeError.Offset,
		)

	case errors.As(err, &invalidUnmarshalError):
		panic(err)

	case errors.Is(err, io.ErrUnexpectedEOF):
		return fmt.Errorf("JSON Body: contains malformed JSON")

	case errors.Is(err, io.EOF):
		return fmt.Errorf("JSON Body: must not be empty")

	default:
		return err
	}
}
