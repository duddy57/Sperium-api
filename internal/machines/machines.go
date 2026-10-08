package machines

import (
	"context"
	"time"
	"uuid"
)

type Machine struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`

	Name string `json:"name"`

	Hostname     string `json:"hostname,omitempty"`
	AgentVersion string `json:"agent_version,omitempty"`

	LastSeenAt time.Time `json:"last_seen_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateMachineReq struct {
	Name string `json:"name" validate:"required,min=2,max=50"`
}

type UpdateMachineReq struct {
	Name string `json:"name" validate:"required,min=2,max=50"`
}

type ActivateMachineReq struct {
	Token        string `json:"token" validate:"required,min=2"`
	Hostname     string `json:"hostname" validate:"required,min=2,max=50"`
	AgentVersion string `json:"agent_version" validate:"required,min=2,max=50"`
}

type machineService interface {
	Create(context.Context, CreateMachineReq, uuid.UUID) (uuid.UUID, string, error)
	Details(context.Context, uuid.UUID) (Machine, error)
	Delete(context.Context, uuid.UUID) error
	List(context.Context, uuid.UUID) ([]Machine, error)
	Update(context.Context, uuid.UUID, UpdateMachineReq) error
	Activate(context.Context, uuid.UUID, ActivateMachineReq) error
}

type getMachineResp struct {
	Message string  `json:"message"`
	Machine Machine `json:"machine"`
}
type listMachineResp struct {
	Message string    `json:"message"`
	Machine []Machine `json:"machine"`
}
type createMachineResp struct {
	Message string    `json:"message"`
	ID      uuid.UUID `json:"id"`
	Token   string    `json:"token"`
}

type genericResp struct {
	Message string `json:"message"`
}
