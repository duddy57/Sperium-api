package organizations

import "errors"

var (
	ErrOrganizationAlreadyExists       = errors.New("organization already exists")
	ErrOrganizationNotFound            = errors.New("organization not found")
	ErrOrganizationInternalServerError = errors.New("organization internal server error")
)
