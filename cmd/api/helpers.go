package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"greenlight.badrchoubai.dev/internal/validator"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Define an envelope type
type envelope map[string]any

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

func (app *application) readStringValue(qs url.Values, key, defaultValue string) string {
	val := qs.Get(key)

	if val == "" {
		return defaultValue
	}

	return val
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
	}

	return i
}

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
	// Request may be a maximum of 1MB
	const MAX_BYTES = 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(MAX_BYTES))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

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

		case errors.As(err, &maxBytesError):
			return fmt.Errorf(
				"JSON Body: May not be larger then %d bytes",
				maxBytesError.Limit,
			)

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf(
				"JSON Body: contains unknown key %s",
				fieldName,
			)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("JSON Body: contains malformed JSON")

		case errors.Is(err, io.EOF):
			return fmt.Errorf("JSON Body: must not be empty")

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New(
			"request body: must only contain a single JSON object",
		)
	}

	return nil
}
