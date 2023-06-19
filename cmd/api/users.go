package main

import (
	"errors"
	"greenlight.badrchoubai.dev/internal/data"
	"greenlight.badrchoubai.dev/internal/validator"
	"net/http"
)

func (application *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := application.readJSON(w, r, &input)
	if err != nil {
		application.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		application.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		application.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = application.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists")
			application.failedValidationResponse(w, r, v.Errors)
		default:
			application.serverErrorResponse(w, r, err)
		}
		return
	}

	err = application.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		application.serverErrorResponse(w, r, err)
	}
}
