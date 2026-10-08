package machines

import "errors"

var (
	ErrMachineAlreadyExists       = errors.New("machine already exists")
	ErrMachineNotFound            = errors.New("machine not found")
	ErrMachineInternalServerError = errors.New("machine internal server error")
)
