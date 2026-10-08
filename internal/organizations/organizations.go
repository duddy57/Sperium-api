package organizations

import (
	"context"
	"time"
	"uuid"
)

type Organization struct {
	ID uuid.UUID `json:"id,omitempty"`

	Name   string `json:"name,omitempty"`
	Domain string `json:"domain,omitempty"`
	Slug   string `json:"slug,omitempty"`

	Plan string `json:"plan,omitempty"`

	Members []Member `json:"members,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Member struct {
	MemberID uuid.UUID `json:"member_id"`
	ID       uuid.UUID `json:"id"`

	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateOrganizationReq struct {
	Name   string `json:"name" validate:"required,min=2,max=50"`
	Domain string `json:"domain" validate:"required,min=2,max=50"`
}

type UpdateOrganizationReq struct {
	Name   string `json:"name" validate:"required,min=2,max=50"`
	Domain string `json:"domain" validate:"required,min=2,max=50"`
}

type GetOrganizationReq struct {
	Message      string       `json:"message"`
	Organization Organization `json:"organization"`
}

type organizationService interface {
	Create(context.Context, CreateOrganizationReq) (uuid.UUID, string, error)
	Details(context.Context, uuid.UUID) (Organization, error)
	Delete(context.Context, uuid.UUID) error
	GetDefault(context.Context) (Organization, error)
	Update(context.Context, uuid.UUID, UpdateOrganizationReq) error
}

type createOrganizationResp struct {
	Message string    `json:"message"`
	Slug    string    `json:"slug"`
	ID      uuid.UUID `json:"id"`
}

type genericResp struct {
	Message string `json:"message"`
}
