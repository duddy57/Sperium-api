package machines

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestActivateInvalidTokenDoesNotConsumeOrStartTransaction(t *testing.T) {
	service := NewMachineServices(nil, nil, zap.NewNop(), scs.New())

	err := service.Activate(context.Background(), uuid.New(), ActivateMachineReq{Token: "not base64"})

	assert.True(t, errors.Is(err, ErrMachineNotFound))
}

func TestMapMachines(t *testing.T) {
	input := []int{1, 2, 3}
	result := Map(input, func(value int) Machine { return Machine{Name: string(rune('0' + value))} })

	assert.Len(t, result, len(input))
	assert.Equal(t, "1", result[0].Name)
}
