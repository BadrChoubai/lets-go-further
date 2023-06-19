package data

import (
	"greenlight.badrchoubai.dev/internal/validator"
	"math"
	"strings"
)

type (
	Filters interface {
		sortColumn() string
		sortDirection() string
		limit() int
		offset() int
	}

	FilterOptions struct {
		Page           int
		PageSize       int
		Sort           string
		SortableValues []string
	}
)

func ValidateFilters(v *validator.Validator, filterOpts FilterOptions) {
	v.Check(filterOpts.Page > 0, "page", "must be greater than zero")
	v.Check(filterOpts.Page <= 1000, "page", "must be less than a thousand")
	v.Check(filterOpts.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(filterOpts.PageSize <= 100, "page_size", "must be a maximum of 100")

	v.Check(validator.PermittedValue(filterOpts.Sort, filterOpts.SortableValues...), "sort", "invalid sort values")
}

func (filterOpts FilterOptions) sortColumn() string {
	for _, sortableValue := range filterOpts.SortableValues {
		if filterOpts.Sort == sortableValue {
			return strings.TrimPrefix(filterOpts.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + filterOpts.Sort)
}

func (filterOpts FilterOptions) sortDirection() string {
	if strings.HasPrefix(filterOpts.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func (filterOpts FilterOptions) limit() int {
	return filterOpts.PageSize
}

func (filterOpts FilterOptions) offset() int {
	return (filterOpts.Page - 1) * filterOpts.PageSize
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
