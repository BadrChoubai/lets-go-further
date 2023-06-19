package main

import (
	"errors"
	"greenlight.badrchoubai.dev/internal/data"
	"greenlight.badrchoubai.dev/internal/validator"
	"net/http"
	"time"
)

func (application *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := application.readJSON(w, r, &input)
	if err != nil {
		application.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		application.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := application.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			application.failedValidationResponse(w, r, v.Errors)
		default:
			application.serverErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true

	err = application.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			application.editConflictResponse(w, r)
		default:
			application.serverErrorResponse(w, r, err)
		}
	}

	err = application.models.Token.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		application.serverErrorResponse(w, r, err)
		return
	}

	err = application.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		application.serverErrorResponse(w, r, err)
	}
}

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

	token, err := application.models.Token.New(user.ID, 1*24*time.Hour, data.ScopeActivation)
	if err != nil {
		application.serverErrorResponse(w, r, err)
		return
	}

	application.background(func() {
		userActivationInfo := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		err = application.mailer.Send(user.Email, "user_welcome.tmpl", userActivationInfo)
		if err != nil {
			application.log.PrintError(err, nil)
		}
	})

	err = application.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		application.serverErrorResponse(w, r, err)
	}
}
