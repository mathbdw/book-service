package postgres

import (
	"fmt"

	"github.com/mathbdw/book/internal/domain/entities"
	"github.com/mathbdw/book/internal/errors"
)

type servicePagination struct{}

type ServicePagination interface {
	CreateCursor(model entities.Paginatable, field string) (string, error)
	CreatePageInfo(model []entities.Paginatable, params entities.PaginationParams) (entities.PageInfo, error)
}

// NewService - Constructor servicePagination
func NewService() ServicePagination {
	return &servicePagination{}
}

// CreateCursor - Creates a cursor from a Paginatable field
func (s *servicePagination) CreateCursor(model entities.Paginatable, field string) (string, error) {
	value, err := model.GetFieldAsString(field)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("servicePostgres.CreateCursor: field not found - %s", field))
	}

	var strCursor string
	if _, ok := entities.SortTypesUnique[field]; ok {
		strCursor, err = EncodeCursor(value, nil)
		if err != nil {
			return "", errors.Wrap(err, "servicePostgres.CreateCursor: failed to encode cursor SortTypesUnique")
		}
	} else {
		createdAt := model.GetCreatedAt()
		strCursor, err = EncodeCursor(value, &createdAt)
		if err != nil {
			return "", errors.Wrap(err, "servicePostgres.CreateCursor: failed to encode cursor")
		}
	}

	return strCursor, nil
}

// CreatePageInfo - Creates the PageInfo from a slice of Paginatable
func (s *servicePagination) CreatePageInfo(model []entities.Paginatable, params entities.PaginationParams) (entities.PageInfo, error) {
	var nextCursor string
	var err error

	if uint64(len(model)) > params.Limit {
		nextCursor, err = s.CreateCursor(model[params.Limit-1], string(params.SortBy))
		if err != nil {
			return entities.PageInfo{}, errors.Wrap(err, "servicePostgres.CreatePageInfo: error nextCursor")
		}
	}

	return entities.PageInfo{NextCursor: nextCursor}, nil
}
