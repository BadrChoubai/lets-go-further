package data

import (
	"greenlight.badrchoubai.dev/internal/validator"
	"math"
	"strings"
)

type (
	IFilters interface {
		sortColumn() string
		sortDirection() string
		limit() int
		offset() int
	}

	Filters struct {
		Page           int
		PageSize       int
		Sort           string
		SortableValues []string
	}
)

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 1000, "page", "must be less than a thousand")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	v.Check(validator.PermittedValue(f.Sort, f.SortableValues...), "sort", "invalid sort values")
}

func (f Filters) sortColumn() string {
	for _, sortableValue := range f.SortableValues {
		if f.Sort == sortableValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

type (
	Metadata struct {
		CurrentPage  int `json:"current_page,omitempty"`
		PageSize     int `json:"page_size,omitempty"`
		FirstPage    int `json:"first_page,omitempty"`
		LastPage     int `json:"last_page,omitempty"`
		TotalRecords int `json:"total_records,omitempty"`
	}
)

func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
