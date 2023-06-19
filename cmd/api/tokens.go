package main

import (
	"errors"
	"greenlight.badrchoubai.dev/internal/data"
	"greenlight.badrchoubai.dev/internal/validator"
	"net/http"
	"time"
)

func (application *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := application.readJSON(w, r, &input)
	if err != nil {
		application.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		application.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := application.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			application.invalidCredentialsResponse(w, r)
		default:
			application.serverErrorResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		application.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		application.invalidCredentialsResponse(w, r)
		return
	}

	token, err := application.models.Token.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		application.serverErrorResponse(w, r, err)
		return
	}

	err = application.writeJSON(
		w,
		http.StatusCreated,
		envelope{"authentication_token": token},
		nil,
	)
}
