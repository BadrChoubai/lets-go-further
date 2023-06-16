package data

import "greenlight.badrchoubai.dev/internal/validator"

type Filters struct {
	Page           int
	PageSize       int
	Sort           string
	SortableValues []string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 1000, "page", "must be less than a thousand")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	v.Check(validator.PermittedValue(f.Sort, f.SortableValues...), "sort", "invalid sort values")
}
