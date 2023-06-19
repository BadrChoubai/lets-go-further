package main

import (
	"errors"
	"fmt"
	"greenlight.badrchoubai.dev/internal/data"
	"greenlight.badrchoubai.dev/internal/validator"
	"net/http"
)

func (application *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var qsValues struct {
		Title  string
		Genres []string
		data.FilterOptions
	}

	v := validator.New()
	qs := r.URL.Query()

	// Data values
	qsValues.Title = application.readStringValue(qs, "title", "")
	qsValues.Genres = application.readCSV(qs, "genres", []string{})

	// Pagination values
	qsValues.FilterOptions.Page = application.readInt(qs, "page", 1, v)
	qsValues.FilterOptions.PageSize = application.readInt(qs, "page_size", 20, v)
	qsValues.FilterOptions.Sort = application.readStringValue(qs, "sort", "id")

	qsValues.FilterOptions.SortableValues = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	if data.ValidateFilters(v, qsValues.FilterOptions); !v.Valid() {
		application.failedValidationResponse(w, r, v.Errors)
		return
	}

	movies, metadata, err := application.models.Movies.GetAll(qsValues.Title, qsValues.Genres, qsValues.FilterOptions)
	if err != nil {
		application.serverErrorResponse(w, r, err)
		return
	}

	err = application.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "movies": movies}, nil)
	if err != nil {
		application.serverErrorResponse(w, r, err)
	}
}

func (application *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := application.readJSON(w, r, &input)
	if err != nil {
		application.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	if data.ValidateMovie(v, movie); !v.Valid() {
		application.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = application.models.Movies.Insert(movie)
	if err != nil {
		application.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/movies/%d", movie.ID))

	err = application.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		application.serverErrorResponse(w, r, err)
	}
}

func (application *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := application.readIDParam(r)
	if err != nil || id < 1 {
		application.notFoundResponse(w, r)
		return
	}

	movie, err := application.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			application.notFoundResponse(w, r)
		default:
			application.serverErrorResponse(w, r, err)
		}
		return
	}

	err = application.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		application.serverErrorResponse(w, r, err)
	}
}

func (application *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := application.readIDParam(r)
	if err != nil || id < 1 {
		application.notFoundResponse(w, r)
		return
	}

	movie, err := application.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			application.notFoundResponse(w, r)
		default:
			application.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	err = application.readJSON(w, r, &input)
	if err != nil {
		application.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		application.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = application.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			application.editConflictResponse(w, r)
		default:
			application.serverErrorResponse(w, r, err)
		}
		return
	}

	err = application.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		application.serverErrorResponse(w, r, err)
	}
}

func (application *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := application.readIDParam(r)
	if err != nil {
		application.notFoundResponse(w, r)
		return
	}

	err = application.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			application.notFoundResponse(w, r)
		default:
			application.serverErrorResponse(w, r, err)
		}
		return
	}

	err = application.writeJSON(w, http.StatusOK, envelope{"message": "movie deleted successfully"}, nil)
	if err != nil {
		application.serverErrorResponse(w, r, err)
	}
}
